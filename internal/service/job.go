package service

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/vo"
	"github.com/dronesphere/pkg/coordinate"
	"github.com/dronesphere/pkg/wpml"
	"github.com/jinzhu/copier"
)

type (
	JobSvc interface {
		Repo() JobRepo
		FetchByID(ctx context.Context, id uint) (*entity.Job, error)
		FetchAvailableAreas(ctx context.Context) ([]*entity.Area, error)
		FetchAvailableDrones(ctx context.Context) ([]entity.Drone, error)
		FetchAll(ctx context.Context, jobName, areaName string, scheduleTimeStart, scheduleTimeEnd string) ([]entity.Job, error)
		CreateJob(ctx context.Context, name, description string, areaID uint, scheduleTime time.Time, drones []po.JobDronePO, waylines []po.JobWaylinePO, command_drones []po.JobCommandDronePO) (uint, error)
		ModifyJob(ctx context.Context, id uint, name, description string, areaID uint, scheduleTime time.Time, drones []po.JobDronePO, waylines []po.JobWaylinePO, command_drones []po.JobCommandDronePO) (*entity.Job, error)
	}

	JobRepo interface {
		Save(ctx context.Context, job *po.Job) error
		DeleteByID(ctx context.Context, id uint) error
		FetchPOByID(ctx context.Context, id uint) (*po.Job, error)
		SelectByID(ctx context.Context, id uint) (*po.Job, error)
		SelectAll(ctx context.Context, jobName, areaName string, scheduleTimeStart, scheduleTimeEnd string) ([]po.Job, error)
		SelectPhysicalDrones(ctx context.Context) ([]dto.PhysicalDrone, error)
	}
)

type JobImpl struct {
	jobRepo   JobRepo
	areaRepo  AreaRepo
	droneRepo DroneRepo
	modelRepo ModelRepo
	l         *slog.Logger
}

func NewJobImpl(jobRepo JobRepo, areaRepo AreaRepo, droneRepo DroneRepo, modelRepo ModelRepo, l *slog.Logger) *JobImpl {
	return &JobImpl{
		jobRepo:   jobRepo,
		areaRepo:  areaRepo,
		droneRepo: droneRepo,
		modelRepo: modelRepo,
		l:         l,
	}
}

// toEntity 将 po.Area 转换为 entity.Area
// 在服务层中进行包装，避免循环引用
func (s *JobImpl) toAreaEntity(p *po.Area) *entity.Area {
	if p == nil {
		return nil
	}

	var points []vo.GeoPoint
	for _, point := range p.Points {
		var p vo.GeoPoint
		if err := copier.Copy(&p, point); err != nil {
			s.l.Error("复制点数据失败", slog.Any("error", err))
			return nil
		}
		points = append(points, p)
	}

	var area entity.Area
	if err := copier.Copy(&area, p); err != nil {
		s.l.Error("复制区域数据失败", slog.Any("error", err))
		return nil
	}
	area.Points = points
	return &area
}

func (j *JobImpl) Repo() JobRepo {
	return j.jobRepo
}

func (j *JobImpl) FetchAvailableAreas(ctx context.Context) ([]*entity.Area, error) {
	areas, err := j.areaRepo.FetchAll(ctx, "", "", "")
	if err != nil {
		return nil, err
	}
	var areaEntities []*entity.Area
	for _, area := range areas {
		areaEntity := j.toAreaEntity(area)
		if areaEntity == nil {
			j.l.Error("转换区域数据失败", slog.Any("area", area))
			return nil, err
		}
		areaEntities = append(areaEntities, areaEntity)
	}

	return areaEntities, nil
}

func (j *JobImpl) FetchAvailableDrones(ctx context.Context) ([]entity.Drone, error) {
	// 调用时传递空参数，表示不进行过滤
	drones, err := j.droneRepo.SelectAll(ctx, "", "", 0)
	if err != nil {
		return nil, err
	}
	return drones, nil
}

func (j *JobImpl) FetchByID(ctx context.Context, id uint) (*entity.Job, error) {
	job, err := j.jobRepo.SelectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	area, err := j.areaRepo.SelectByID(ctx, job.AreaID)
	if err != nil {
		return nil, err
	}
	areaEntity := j.toAreaEntity(area)
	var entity entity.Job
	entity.Area = *areaEntity
	entity.ID = job.ID
	entity.Name = job.Name
	entity.Description = job.Description
	entity.ScheduleTime = job.ScheduleTime
	entity.Drones = job.Drones
	entity.Waylines = job.Waylines
	entity.CommandDrones = job.CommandDrones

	return &entity, nil
}

func (j *JobImpl) CreateJob(ctx context.Context, name, description string, areaID uint, scheduleTime time.Time, drones []po.JobDronePO, waylines []po.JobWaylinePO, commandDrones []po.JobCommandDronePO) (uint, error) {
	job := &po.Job{
		Name:          name,
		Description:   description,
		AreaID:        areaID,
		ScheduleTime:  scheduleTime,
		Drones:        drones,
		Waylines:      waylines,
		CommandDrones: commandDrones,
	}
	if err := j.jobRepo.Save(ctx, job); err != nil {
		return 0, err
	}
	j.l.Info("Job created", "job", job)
	// 逐个创建航线文件
	for _, w := range job.Waylines {
		var drone po.JobDronePO
		for _, d := range job.Drones {
			if d.Key == w.DroneKey {
				drone = d
				break
			}
		}
		var variation po.DroneVariation
		variations, err := j.modelRepo.SelectAllDroneVariation(ctx, nil)
		if err != nil {
			j.l.Error("获取无人机变体失败", slog.Any("error", err))
			return 0, err
		}
		for _, v := range variations {
			if v.ID == drone.VariationID {
				variation = v
				break
			}
		}

		waylineDoc, err := j.generateWayline(ctx, drone.PhysicalDroneID, variation, drone.TakeoffPoint, w)
		if err != nil {
			j.l.Error("Failed to generate wayline", slog.Any("error", err))
			return 0, err
		}
		j.l.Info("航线文件已创建", "wayline", waylineDoc)
		template, err := waylineDoc.GenerateXML()
		if err != nil {
			j.l.Error("生成航线文件失败", slog.Any("error", err))
			return 0, err
		}
		j.l.Info("航线文件已生成", "template", template)
		wayline, err := waylineDoc.GenerateXML()
		if err != nil {
			j.l.Error("生成航线文件失败", slog.Any("error", err))
			return 0, err
		}
		j.l.Info("航线文件已生成", "wayline", wayline)

		// 保存航线文件
		// 创建kmz目录
		if err := os.MkdirAll("kmz", os.ModePerm); err != nil {
			j.l.Error("创建kmz目录失败", slog.Any("error", err))
			return 0, err
		}
		kmzFileName := "kmz/" + "job-" + strconv.Itoa(int(job.ID)) + "-" + "drone-key-" + w.DroneKey + "-" + "drone-id-" + strconv.Itoa(int(drone.PhysicalDroneID)) + ".kmz"
		err = wpml.GenerateKMZ(kmzFileName, template, wayline)
		if err != nil {
			j.l.Error("生成航线文件失败", slog.Any("error", err))
			return 0, err
		}
		j.l.Info("航线文件已生成", "kmzFileName", kmzFileName)
	}
	return job.ID, nil
}

func (j *JobImpl) ModifyJob(ctx context.Context, id uint, name, description string, areaID uint, scheduleTime time.Time, drones []po.JobDronePO, waylines []po.JobWaylinePO, command_drones []po.JobCommandDronePO) (*entity.Job, error) {
	// 获取已存在的任务
	p, err := j.jobRepo.SelectByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新任务信息
	p.Name = name
	p.Description = description
	p.AreaID = areaID
	p.ScheduleTime = scheduleTime
	p.Drones = drones
	p.Waylines = waylines
	p.CommandDrones = command_drones

	// 保存更新
	if err := j.jobRepo.Save(ctx, p); err != nil {
		return nil, err
	}

	// 更新航线文件
	// for _, w ÷:= range waylines {
	// var dr dto.JobCreationDrone
	// for _, d := range drones {
	// 	if d.Key == w.DroneKey {
	// 		dr = d
	// 		break
	// 	}
	// }
	// waylineKey, err := j.jobRepo.CreateWaylineFile(ctx, p.Name, dr, w)
	// if err != nil {
	// 	j.l.Error("更新航线文件失败", slog.Any("error", err))
	// 	return nil, err
	// }
	// j.l.Info("航线文件已更新", "waylineKey", waylineKey)
	// }

	// 返回更新后的任务
	return j.FetchByID(ctx, id)
}

func (j *JobImpl) FetchAll(ctx context.Context, jobName, areaName string, scheduleTimeStart, scheduleTimeEnd string) ([]entity.Job, error) {
	// 调用时传递空字符串作为时间参数，表示不按时间筛选
	jobs, err := j.jobRepo.SelectAll(ctx, jobName, areaName, scheduleTimeStart, scheduleTimeEnd)
	if err != nil {
		return nil, err
	}

	var result []entity.Job
	for _, job := range jobs {
		jobEntity := entity.Job{}
		if err := copier.Copy(&jobEntity, job); err != nil {
			j.l.Error("复制任务数据失败", slog.Any("error", err))
			return nil, err
		}
		result = append(result, jobEntity)
	}

	return result, nil
}

func (j *JobImpl) generateWayline(ctx context.Context, droneID uint, droneVariation po.DroneVariation, takeoffPoint po.JobTakeoffPointPO, wayline po.JobWaylinePO) (wpml.Document, error) {
	droneModel := droneVariation.DroneModel
	gimbals := droneVariation.Gimbals
	drone, err := j.droneRepo.SelectByID(ctx, droneID)
	if err != nil {
		j.l.Error("获取无人机信息失败", slog.Any("error", err))
		return wpml.Document{}, err
	}
	j.l.Info("生成航线", "droneID", droneID, "droneModel", droneModel, "gimbals", gimbals, "drone", drone)

	author := "System" // 先创建一个字符串变量
	timeNow := time.Now().UnixMilli()
	doc := wpml.Document{
		Author:     &author, // 使用该变量的地址
		CreateTime: &timeNow,
		UpdateTime: &timeNow,
		Mission: wpml.MissionConfig{
			FlyToWaylineMode:        wpml.FlyToWaylineSafely,            // 安全飞行
			FinishAction:            wpml.FinishActionGoHome,            // 完成后返回
			ExitOnRCLost:            wpml.ExitOnRCLostExecuteLostAction, // 失去信号后执行失控动作
			ExecuteRCLostAction:     wpml.RCLostActionGoBack,            // 失去信号后返回
			TakeOffSecurityHeight:   10,
			GlobalTransitionalSpeed: 2,
			DroneInfo: wpml.DroneInfo{
				DroneEnumValue:    wpml.DroneEnumValue(droneModel.Type),
				DroneSubEnumValue: wpml.DroneSubEnumValue(droneModel.SubType),
			},
			PayloadInfo: wpml.PayloadInfo{
				PayloadEnumValue:     wpml.PayloadEnumValue(gimbals[0].Type),
				PayloadSubEnumValue:  wpml.PayloadSubEnumValue(gimbals[0].SubType),
				PayloadPositionIndex: gimbals[0].Gimbalindex,
			},
			AutoRerouteInfo: nil,
		},
	}

	trueBool := wpml.BoolAsInt(true) // 使用变量的地址
	falseBool := wpml.BoolAsInt(false)
	templateType := wpml.TemplateTypeWaypoint // 先创建一个变量
	templateID := 0
	autoFlightSpeed := 1.0
	GimbalPitchMode := wpml.GimbalPitchModeUsePointSetting
	globalWaypointHeadingParam := wpml.WaypointHeadingParam{
		WaypointHeadingMode: wpml.HeadingFollowWayline, // 航点航线
	}
	globalWaypointTurnMode := wpml.ToPointAndStopWithDiscontinuityCurvature
	folder := wpml.Folder{
		TemplateType: &templateType, // 使用变量的地址 // 航点航线
		TemplateID:   &templateID,
		WaylineCoordinateSysParam: &wpml.WaylineCoordinateSysParam{
			CoordinateMode:  wpml.CoordinateModeWGS84,            // WGS84坐标系
			HeightMode:      wpml.HeightModeRelativeToStartPoint, // 高度相对于起点
			PositioningType: wpml.PositioningTypeRTKBaseStation,  // RTK基站定位
		},
		AutoFlightSpeed:            autoFlightSpeed,             // 自动飞行速度
		GimbalPitchMode:            &GimbalPitchMode,            // 使用点设置
		GlobalWaypointHeadingParam: &globalWaypointHeadingParam, // 航点航线
		GlobalWaypointTurnMode:     &globalWaypointTurnMode,     // 航点转弯模式
		GlobalUseStraightLine:      &trueBool,                   // 使用直线
	}

	for idx, waypoint := range wayline.Waypoints {
		lng, lat := coordinate.GCJ02ToWGS84(waypoint.Lng, waypoint.Lat)
		placemark := wpml.Placemark{
			Point:                 wpml.Point{Coordinates: wpml.FormatCoordinates(float64(lng), float64(lat))},
			Index:                 idx,
			UseGlobalHeight:       &falseBool,
			Height:                &wayline.Altitude,
			EllipsoidHeight:       nil,
			ExecuteHeight:         &wayline.Altitude,
			UseGlobalSpeed:        &trueBool,
			UseGlobalHeadingParam: &trueBool,
			WaypointHeadingParam:  nil,
			UseGlobalTurnParam:    &trueBool,
			WaypointTurnParam:     nil,
			UseStraightLine:       &trueBool,
			GimbalPitchAngle:      -90,
			ActionGroup:           nil,
		}
		folder.Placemarks = append(folder.Placemarks, placemark)
	}

	doc.Folders = append(doc.Folders, folder)

	return doc, nil
}
