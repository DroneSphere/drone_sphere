package service

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/vo"
	"github.com/dronesphere/pkg/coordinate"
	"github.com/dronesphere/pkg/wpml"
	"github.com/jinzhu/copier"
	"gorm.io/datatypes"
)

type (
	JobSvc interface {
		Repo() JobRepo
		FetchByID(ctx context.Context, id uint) (*entity.Job, error)
		FetchAvailableAreas(ctx context.Context) ([]*entity.Area, error)
		FetchAvailableDrones(ctx context.Context) ([]entity.Drone, error)
		FetchAll(ctx context.Context, jobName, areaName string, scheduleTimeStart, scheduleTimeEnd string) ([]entity.Job, error)
		CreateJob(ctx context.Context, name, description string, areaID uint, scheduleTime time.Time, drones []po.JobDronePO, waylines []po.JobWaylinePO, command_drones []po.JobCommandDronePO, waylineGenerationParams po.JobWaylineGenerationParams) (uint, error)
		ModifyJob(ctx context.Context, id uint, name, description string, areaID uint, scheduleTime time.Time, drones []po.JobDronePO, waylines []po.JobWaylinePO, command_drones []po.JobCommandDronePO, waylineGenerationParams po.JobWaylineGenerationParams) (*entity.Job, error)
	}

	JobRepo interface {
		Save(ctx context.Context, job *po.Job) error
		DeleteByID(ctx context.Context, id uint) error
		FetchPOByID(ctx context.Context, id uint) (*po.Job, error)
		SelectByID(ctx context.Context, id uint) (*po.Job, error)
		SelectAll(ctx context.Context, jobName, areaName string, scheduleTimeStart, scheduleTimeEnd string) ([]po.Job, error)
		SelectPhysicalDrones(ctx context.Context) ([]dto.PhysicalDrone, error)
		SaveWayline(ctx context.Context, wayline po.Wayline, kmzFile string) (*po.Wayline, error)
	}
)

type JobImpl struct {
	jobRepo     JobRepo
	areaRepo    AreaRepo
	droneRepo   DroneRepo
	modelRepo   ModelRepo
	waylineRepo WaylineRepo
	waylineSvc  WaylineSvc
	l           *slog.Logger
}

func NewJobImpl(jobRepo JobRepo, areaRepo AreaRepo, droneRepo DroneRepo, modelRepo ModelRepo, waylineRepo WaylineRepo, waylineSvc WaylineSvc, l *slog.Logger) *JobImpl {
	return &JobImpl{
		jobRepo:     jobRepo,
		areaRepo:    areaRepo,
		droneRepo:   droneRepo,
		modelRepo:   modelRepo,
		waylineRepo: waylineRepo,
		waylineSvc:  waylineSvc,
		l:           l,
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
	areas, _, err := j.areaRepo.SelectAll(ctx, "", "", "", 0, 0)
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
	drones, _, err := j.droneRepo.SelectAll(ctx, "", "", 0, 0, 0)
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
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		errChan chan error
	)
	errChan = make(chan error, len(job.Drones))
	for i, dronePO := range job.Drones {
		wg.Add(1)
		go func(index int, dronePO *po.JobDronePO) {
			defer wg.Done()
			droneEntity, err := j.FetchDroneEntity(ctx, job.ID, *dronePO)
			if err != nil {
				j.l.Error("获取无人机实体失败", slog.Any("error", err))
				select {
				case errChan <- err:
				default:
					// 如果通道已满，说明其他 Goroutine 已经发生了错误，当前错误可以忽略
					// 或者更严格地，可以在这里加入日志或者计数器来记录被忽略的错误
				}
				return
			}
			mu.Lock()
			entity.Drones = append(entity.Drones, *droneEntity)
			mu.Unlock()
		}(i, &dronePO)
	}
	wg.Wait()
	close(errChan) // 关闭错误通道，表示没有更多的错误会写入
	for err := range errChan {
		if err != nil {
			return nil, err // 返回第一个遇到的错误
		}
	}
	// for _, dronePO := range job.Drones {
	// 	droneEntity, err := j.FetchDroneEntity(ctx, job.ID, dronePO)
	// 	if err != nil {
	// 		j.l.Error("获取无人机实体失败", slog.Any("error", err))
	// 		return nil, err
	// 	}
	// 	entity.Drones = append(entity.Drones, *droneEntity)
	// }
	entity.Waylines = job.Waylines
	entity.CommandDrones = job.CommandDrones
	entity.WaylineGenerationParams = job.WaylineGenerationParams.Data()

	return &entity, nil
}

func (j *JobImpl) FetchDroneEntity(ctx context.Context, jobID uint, dronePO po.JobDronePO) (*entity.JobDrone, error) {
	var wg sync.WaitGroup
	wg.Add(3) // 3个并发查询任务

	// 创建结果和错误通道
	physicalDroneCh := make(chan *po.Drone, 1)
	droneModelCh := make(chan *entity.DroneModel, 1)
	gimbalModelCh := make(chan *po.GimbalModel, 1)
	// waylineCh := make(chan *entity.Wayline, 1)
	errCh := make(chan error, 4)

	// 1. 查询物理无人机信息
	go func() {
		defer wg.Done()
		physicalDrone, err := j.droneRepo.SelectByIDV2(ctx, dronePO.PhysicalDroneID)
		if err != nil {
			j.l.Error("获取无人机信息失败", slog.Any("error", err))
			errCh <- err
			return
		}
		j.l.Info("获取无人机信息", "physicalDrone", physicalDrone)
		physicalDroneCh <- physicalDrone
	}()

	// 2. 查询无人机型号信息
	go func() {
		defer wg.Done()
		droneModel, err := j.modelRepo.SelectDroneModelByID(ctx, dronePO.ModelID)
		if err != nil {
			j.l.Error("获取无人机型号信息失败", slog.Any("error", err))
			errCh <- err
			return
		}
		j.l.Info("获取无人机型号信息", "droneModel", droneModel)
		droneModelCh <- droneModel
	}()

	// 3. 查询云台型号信息
	go func() {
		defer wg.Done()
		variations, err := j.modelRepo.SelectAllDroneVariation(ctx, nil)
		if err != nil {
			j.l.Error("获取云台型号信息失败", slog.Any("error", err))
			errCh <- err
			return
		}
		var variation po.DroneVariation
		for _, e := range variations {
			if e.DroneModelID == dronePO.ModelID {
				variation = e
				break
			}
		}
		j.l.Info("获取云台型号信息", "gimbalModel", variation.Gimbals[0])
		gimbalModelCh <- &variation.Gimbals[0]
	}()

	// 4. 查询航线信息
	// go func() {
	// 	defer wg.Done()
	// 	// 不再需要等待物理无人机查询结果
	// 	wayline, err := j.waylineSvc.FetchWaylineByJobIDAndDroneKey(ctx, jobID, dronePO.Key)
	// 	if err != nil {
	// 		j.l.Error("获取航线信息失败", slog.Any("error", err))
	// 		errCh <- err
	// 		return
	// 	}
	// 	j.l.Info("获取航线信息", "wayline", wayline)
	// 	waylineCh <- wayline
	// }()

	// 等待所有goroutine完成
	wg.Wait()
	close(errCh)

	// 检查是否有错误
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	// 收集结果
	physicalDrone := <-physicalDroneCh
	droneModel := <-droneModelCh
	gimbalModel := <-gimbalModelCh
	// wayline := <-waylineCh

	// 组装结果
	var drone entity.JobDrone
	if err := copier.Copy(&drone, dronePO); err != nil {
		j.l.Error("复制无人机数据失败", slog.Any("error", err))
		return nil, err
	}
	drone.DroneModel = *droneModel
	drone.GimbalModel = *gimbalModel
	drone.PhysicalDrone = *physicalDrone
	// drone.Wayline = *wayline

	return &drone, nil
}

func (j *JobImpl) CreateJob(ctx context.Context, name, description string, areaID uint, scheduleTime time.Time, drones []po.JobDronePO, waylines []po.JobWaylinePO, commandDrones []po.JobCommandDronePO, waylineGenerationParams po.JobWaylineGenerationParams) (uint, error) {
	job := &po.Job{
		Name:                    name,
		Description:             description,
		AreaID:                  areaID,
		ScheduleTime:            scheduleTime,
		Drones:                  drones,
		Waylines:                waylines,
		CommandDrones:           commandDrones,
		WaylineGenerationParams: datatypes.NewJSONType(waylineGenerationParams),
	}
	if err := j.jobRepo.Save(ctx, job); err != nil {
		return 0, err
	}
	j.l.Info("Job created", "job", job)
	// 逐个创建航线文件
	for _, w := range job.Waylines {
		// 找到对应的无人机
		var drone po.JobDronePO
		for _, d := range job.Drones {
			if d.Key == w.DroneKey {
				drone = d
				break
			}
		}

		// 调用抽取的方法创建和保存航线文件
		_, err := j.createWorkerWaylineFile(ctx, job.ID, job.Name, drone, w, waylineGenerationParams)
		if err != nil {
			j.l.Error("创建航线文件失败", slog.Any("error", err))
			return 0, err
		}
	}

	// 检查是否有指挥机，创建指挥机的航线文件
	if len(commandDrones) > 0 {
		for _, commandDrone := range commandDrones {
			// 找到对应的无人机
			var drone po.JobDronePO
			for _, d := range drones {
				if d.Key == commandDrone.DroneKey {
					drone = d
					break
				}
			}
			// 调用抽取的方法创建和保存指挥机的航线文件
			_, err := j.createCommandWaylineFile(ctx, job.ID, job.Name, drone, commandDrone)
			if err != nil {
				j.l.Error("创建指挥机航线文件失败", slog.Any("error", err), slog.String("droneKey", commandDrone.DroneKey))
				return 0, err
			}
		}
	}
	j.l.Info("指挥机航线文件已创建", slog.Any("commandDrones", commandDrones))

	return job.ID, nil
}

func (j *JobImpl) ModifyJob(ctx context.Context, id uint, name, description string, areaID uint, scheduleTime time.Time, drones []po.JobDronePO, waylines []po.JobWaylinePO, command_drones []po.JobCommandDronePO, waylineGenerationParams po.JobWaylineGenerationParams) (*entity.Job, error) {
	// 获取已存在的任务
	p, err := j.jobRepo.SelectByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldWaylines := p.Waylines
	// 更新任务信息
	p.Name = name
	p.Description = description
	p.AreaID = areaID
	p.ScheduleTime = scheduleTime
	p.Drones = drones
	p.Waylines = waylines
	p.CommandDrones = command_drones
	p.WaylineGenerationParams = datatypes.NewJSONType(waylineGenerationParams)

	// 保存更新
	if err := j.jobRepo.Save(ctx, p); err != nil {
		j.l.Error("保存任务失败", slog.Any("error", err))
		return nil, err
	}

	// 删除旧的航线文件
	var deleteWg sync.WaitGroup
	deleteErrCh := make(chan error, len(oldWaylines))

	for _, w := range oldWaylines {
		deleteWg.Add(1)
		// 使用闭包捕获当前循环变量
		go func(wayline po.JobWaylinePO) {
			defer deleteWg.Done()

			// 通过 JobID 和 DroneKey 查询航线文件
			existWayline, err := j.waylineRepo.SelectByJobIDAndDroneKey(ctx, id, wayline.DroneKey)
			if err != nil {
				j.l.Error("获取航线文件失败", slog.Any("error", err), slog.String("droneKey", wayline.DroneKey))
				deleteErrCh <- err
				return
			}

			// 删除已有的航线文件
			err = j.waylineRepo.DeleteByID(ctx, existWayline.ID)
			if err != nil {
				j.l.Error("删除航线文件失败", slog.Any("error", err), slog.Any("waylineID", existWayline.ID))
				deleteErrCh <- err
				return
			}
			j.l.Info("成功删除航线文件", slog.Any("waylineID", existWayline.ID), slog.String("droneKey", wayline.DroneKey))
		}(w)
	}

	// 等待所有删除操作完成
	deleteWg.Wait()
	close(deleteErrCh)

	// 检查是否有删除错误
	if len(deleteErrCh) > 0 {
		err := <-deleteErrCh
		j.l.Error("删除航线文件过程中出现错误", slog.Any("error", err))
		return nil, err
	}

	// 更新航线文件
	var createWg sync.WaitGroup
	createErrCh := make(chan error, len(waylines))

	for _, w := range waylines {
		createWg.Add(1)
		// 使用闭包捕获当前循环变量
		go func(wayline po.JobWaylinePO) {
			defer createWg.Done()

			// 找到对应的无人机
			var drone po.JobDronePO
			for _, d := range drones {
				if d.Key == wayline.DroneKey {
					drone = d
					break
				}
			}

			// 调用抽取的方法创建和保存航线文件
			_, err := j.createWorkerWaylineFile(ctx, id, p.Name, drone, wayline, waylineGenerationParams)
			if err != nil {
				j.l.Error("创建航线文件失败", slog.Any("error", err), slog.String("droneKey", wayline.DroneKey))
				createErrCh <- err
				return
			}
			j.l.Info("成功创建航线文件", slog.String("droneKey", wayline.DroneKey))
		}(w)
	}

	// 等待所有创建操作完成
	createWg.Wait()
	close(createErrCh)

	// 检查是否有创建错误
	if len(createErrCh) > 0 {
		err := <-createErrCh
		j.l.Error("创建航线文件过程中出现错误", slog.Any("error", err))
		return nil, err
	}

	// 检查是否有指挥机，创建指挥机的航线文件
	if len(command_drones) > 0 {
		for _, commandDrone := range command_drones {
			// 找到对应的无人机
			var drone po.JobDronePO
			for _, d := range drones {
				if d.Key == commandDrone.DroneKey {
					drone = d
					break
				}
			}
			// 调用抽取的方法创建和保存指挥机的航线文件
			_, err := j.createCommandWaylineFile(ctx, id, p.Name, drone, commandDrone)
			if err != nil {
				j.l.Error("创建指挥机航线文件失败", slog.Any("error", err), slog.String("droneKey", commandDrone.DroneKey))
				return nil, err
			}
		}
	}
	j.l.Info("指挥机航线文件已创建", slog.Any("commandDrones", command_drones))

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
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(jobs))

	// 为每个job启动goroutine
	for _, job := range jobs {
		wg.Add(1)
		go func(jobPO po.Job) {
			defer wg.Done()

			// 创建job实体
			jobEntity := entity.Job{}
			if err := copier.Copy(&jobEntity, jobPO); err != nil {
				j.l.Error("复制任务数据失败", slog.Any("error", err))
				errChan <- err
				return
			}

			// 为当前job的所有drone启动并发查询
			var droneWg sync.WaitGroup
			var droneMu sync.Mutex
			droneErrChan := make(chan error, len(jobPO.Drones))

			for _, dronePO := range jobPO.Drones {
				droneWg.Add(1)
				go func(po po.JobDronePO) {
					defer droneWg.Done()

					// 获取无人机实体
					droneEntity, err := j.FetchDroneEntity(ctx, jobPO.ID, po)
					if err != nil {
						j.l.Error("获取无人机实体失败", slog.Any("error", err))
						select {
						case droneErrChan <- err:
						default:
							// 通道已满，忽略重复错误
						}
						return
					}

					// 线程安全地添加到结果中
					droneMu.Lock()
					jobEntity.Drones = append(jobEntity.Drones, *droneEntity)
					droneMu.Unlock()
				}(dronePO)
			}

			// 等待当前job的所有drone查询完成
			droneWg.Wait()
			close(droneErrChan)

			// 检查drone查询是否有错误
			for err := range droneErrChan {
				if err != nil {
					errChan <- err
					return
				}
			}

			// 线程安全地添加到最终结果中
			mu.Lock()
			result = append(result, jobEntity)
			mu.Unlock()
		}(job)
	}

	// 等待所有job处理完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// createWorkerWaylineFile 创建并保存航线文件
// jobID: 任务ID
// jobName: 任务名称
// drone: 无人机信息
// wayline: 航线信息
// params: 航线生成参数
// 返回保存的航线信息和可能的错误
func (j *JobImpl) createWorkerWaylineFile(ctx context.Context, jobID uint, jobName string, drone po.JobDronePO, wayline po.JobWaylinePO, params po.JobWaylineGenerationParams) (*po.Wayline, error) {
	// 获取无人机变体信息
	var variation po.DroneVariation
	variations, err := j.modelRepo.SelectAllDroneVariation(ctx, nil)
	if err != nil {
		j.l.Error("获取无人机变体失败", slog.Any("error", err))
		return nil, err
	}
	for _, v := range variations {
		if v.ID == drone.VariationID {
			variation = v
			break
		}
	}
	j.l.Info("无人机变体", slog.Any("DroneVariation", variation))

	// 生成航线文件
	waylineDoc, err := j.generateWorkerWayline(ctx, drone.PhysicalDroneID, variation, drone.TakeoffPoint, wayline, params)
	if err != nil {
		j.l.Error("生成航线文件失败", slog.Any("error", err))
		return nil, err
	}
	j.l.Info("航线文件已创建", "wayline", waylineDoc)

	// 生成航线模板和具体航线
	template, err := waylineDoc.GenerateXML()
	if err != nil {
		j.l.Error("生成航线模板失败", slog.Any("error", err))
		return nil, err
	}

	waylineDoc.Author = nil
	waylineDoc.CreateTime = nil
	waylineDoc.UpdateTime = nil
	for i := range waylineDoc.Folders {
		f := &waylineDoc.Folders[i]
		f.TemplateType = nil
		f.WaylineID = &i
		hMode := wpml.ExecuteHeightModeRelativeToStartPoint
		f.ExecuteHeightMode = &hMode
		for j := range f.Placemarks {
			p := &f.Placemarks[j]
			// 处理高度
			// if *p.UseGlobalHeight {
			// 	p.ExecuteHeight = f.GlobalHeight
			// } else {
			// 	p.ExecuteHeight = p.EllipsoidHeight
			// }
			// p.EllipsoidHeight = nil
			// p.Height = nil
			// p.UseGlobalHeight = nil
			// 处理速度
			if *p.UseGlobalSpeed {
				p.WaypointSpeed = f.AutoFlightSpeed
			}
			p.UseGlobalSpeed = nil
			// 处理 GlobalHeadingParam
			if *p.UseGlobalHeadingParam {
				p.WaypointHeadingParam = f.GlobalWaypointHeadingParam
			}
			p.UseGlobalHeadingParam = nil
			// 处理 WaypointTurnParam
			if *p.UseGlobalTurnParam {
				//p.WaypointTurnParam.WaypointTurnMode = *f.GlobalWaypointTurnMode
				p.WaypointTurnParam = &wpml.WaypointTurnParam{
					WaypointTurnMode: *f.GlobalWaypointTurnMode,
				}
			}
			p.UseGlobalTurnParam = nil
			// 处理 StraightLine
			if *p.UseStraightLine {
				p.UseStraightLine = f.GlobalUseStraightLine
			}
			p.UseStraightLine = nil
			// 处理 GimbalPitchMode
			if *f.GimbalPitchMode == wpml.GimbalPitchModeManual {
				p.WaypointGimbalHeadingParam = &wpml.WaypointGimbalHeadingParam{
					WaypointGimbalPitchAngle: 0,
					WaypointGimbalYawAngle:   0,
				}
			}
			workType := wpml.WaypointWorkTypeNone
			p.WaypointWorkType = &workType
		}
		// 擦除不需要的字段
		f.TemplateType = nil
		f.WaylineCoordinateSysParam = nil
		f.PayloadParam = nil
		f.GlobalWaypointTurnMode = nil
		f.GlobalUseStraightLine = nil
		f.GimbalPitchMode = nil
		f.GlobalHeight = nil
		f.GlobalWaypointHeadingParam = nil
	}
	waylineXML, err := waylineDoc.GenerateXML()
	if err != nil {
		j.l.Error("生成航线文件失败", slog.Any("error", err))
		return nil, err
	}

	// 创建kmz目录
	if err := os.MkdirAll("kmz", os.ModePerm); err != nil {
		j.l.Error("创建kmz目录失败", slog.Any("error", err))
		return nil, err
	}

	// 生成KMZ文件
	kmzFileName := "kmz/" + "job-" + strconv.Itoa(int(jobID)) + "-" + "drone-key-" + wayline.DroneKey + "-" + "drone-id-" + strconv.Itoa(int(drone.PhysicalDroneID)) + ".kmz"
	err = wpml.GenerateKMZ(kmzFileName, template, waylineXML)
	if err != nil {
		j.l.Error("生成KMZ文件失败", slog.Any("error", err))
		return nil, err
	}
	j.l.Info("KMZ文件已生成", "kmzFileName", kmzFileName)

	// 获取物理无人机信息
	physicalDrone, err := j.droneRepo.SelectByID(ctx, drone.PhysicalDroneID)
	if err != nil && err.Error() != "no realtime data" {
		j.l.Error("获取无人机信息失败", slog.Any("error", err))
		return nil, err
	}

	// 构建航线信息
	droneModelKey := strconv.Itoa(variation.DroneModel.Domain) + "-" + strconv.Itoa(variation.DroneModel.Type) + "-" + strconv.Itoa(variation.DroneModel.SubType)
	payloadModelKey := "1-" + strconv.Itoa(variation.Gimbals[0].Type) + "-" + strconv.Itoa(variation.Gimbals[0].SubType)
	waylinePO := po.Wayline{
		JobID:       jobID,
		JobDroneKey: wayline.DroneKey,
		DroneSN:     physicalDrone.SN,
		WaylineName: jobName + "-" + wayline.DroneKey,
		StartWaylinePoint: datatypes.NewJSONType(po.StartWaylinePoint{
			StartLatitude:  drone.TakeoffPoint.Lat,
			StartLongitude: drone.TakeoffPoint.Lng,
		}),
		DroneModelKey:    droneModelKey,
		PayloadModelKeys: []string{payloadModelKey},
	}

	// 清理临时文件
	defer func() {
		if err := os.Remove(kmzFileName); err != nil {
			j.l.Error("删除kmz临时文件失败", slog.Any("error", err))
		}
	}()

	// 保存航线文件到数据库
	savedWayline, err := j.jobRepo.SaveWayline(ctx, waylinePO, kmzFileName)
	if err != nil {
		j.l.Error("保存航线信息到数据库失败", slog.Any("error", err))
		return nil, err
	}
	j.l.Info("航线信息已保存到数据库", "wayline", wayline.DroneKey)

	return savedWayline, nil
}

func (j *JobImpl) generateWorkerWayline(ctx context.Context, droneID uint, droneVariation po.DroneVariation, takeoffPoint po.JobTakeoffPointPO, wayline po.JobWaylinePO, params po.JobWaylineGenerationParams) (wpml.Document, error) {
	droneModel := droneVariation.DroneModel
	gimbals := droneVariation.Gimbals
	drone, err := j.droneRepo.SelectByID(ctx, droneID)
	if err != nil && err.Error() != "no realtime data" {
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

	trueBool := wpml.BoolAsInt(true)          // 使用变量的地址
	templateType := wpml.TemplateTypeWaypoint // 先创建一个变量
	templateID := 0
	autoFlightSpeed := 1.0
	GimbalPitchMode := wpml.GimbalPitchModeManual
	globalWaypointHeadingParam := wpml.WaypointHeadingParam{
		WaypointHeadingMode: wpml.HeadingFollowWayline, // 航点航线
	}
	globalHeight := float64(params.FlyingHeight)
	executeHeightMode := wpml.ExecuteHeightModeRelativeToStartPoint // 执行高度模式相对于起点
	globalWaypointTurnMode := wpml.ToPointAndStopWithDiscontinuityCurvature
	folder := wpml.Folder{
		TemplateType: &templateType, // 使用变量的地址 // 航点航线
		TemplateID:   &templateID,
		WaylineCoordinateSysParam: &wpml.WaylineCoordinateSysParam{
			CoordinateMode: wpml.CoordinateModeWGS84,            // WGS84坐标系
			HeightMode:     wpml.HeightModeRelativeToStartPoint, // 高度相对于起点
		},
		AutoFlightSpeed:            autoFlightSpeed,             // 自动飞行速度
		GimbalPitchMode:            &GimbalPitchMode,            // 使用点设置
		GlobalHeight:               &globalHeight,               // 全局高度
		ExecuteHeightMode:          &executeHeightMode,          // 执行高度模式
		GlobalWaypointHeadingParam: &globalWaypointHeadingParam, // 航点航线
		GlobalWaypointTurnMode:     &globalWaypointTurnMode,     // 航点转弯模式
		GlobalUseStraightLine:      &trueBool,                   // 使用直线
	}

	for idx, waypoint := range wayline.Waypoints {
		lng, lat := coordinate.GCJ02ToWGS84(waypoint.Lng, waypoint.Lat)
		height := float64(params.FlyingHeight)
		placemark := wpml.Placemark{
			Point:                 wpml.Point{Coordinates: wpml.FormatCoordinates(float64(lng), float64(lat))},
			Index:                 idx,
			UseGlobalHeight:       &trueBool,
			Height:                &height,
			EllipsoidHeight:       nil,
			ExecuteHeight:         &height,
			UseGlobalSpeed:        &trueBool,
			UseGlobalHeadingParam: &trueBool,
			WaypointHeadingParam:  nil,
			UseGlobalTurnParam:    &trueBool,
			WaypointTurnParam:     nil,
			UseStraightLine:       &trueBool,
			GimbalPitchAngle:      -90,
			ActionGroup:           nil,
		}
		if idx == 0 {
			var actions []wpml.Action
			actionIdx := 0
			trueBool := wpml.BoolAsInt(true)
			falseBool := wpml.BoolAsInt(false)

			if params.GimbalPitch != 0 {
				pitchRotateAction := wpml.Action{
					ActionId:   actionIdx,
					ActionType: wpml.ActionGimbalRotate,
					ActionParams: &wpml.GimbalRotateParams{
						PayloadPositionIndex:    gimbals[0].Gimbalindex,
						GimbalHeadingYawBase:    "north",
						GimbalRotateMode:        "absoluteAngle",
						GimbalPitchRotateEnable: trueBool,
						GimbalPitchRotateAngle:  float64(params.GimbalPitch),
						GimbalYawRotateEnable:   falseBool,
						GimbalYawRotateAngle:    0,
						GimbalRollRotateEnable:  falseBool,
						GimbalRollRotateAngle:   0,
						GimbalRotateTimeEnable:  falseBool,
						GimbalRotateTime:        0,
					},
				}
				actionIdx++
				actions = append(actions, pitchRotateAction)
			}
			if params.GimbalZoom != 1 {
				zoomAction := wpml.Action{
					ActionType: wpml.ActionZoom,
					ActionParams: &wpml.ZoomParams{
						PayloadPositionIndex: gimbals[0].Gimbalindex,
						// TODO: 这里需要根据实际情况设置焦距
						FocalLength: float64(params.GimbalZoom) * 70,
					},
				}
				actionIdx++
				actions = append(actions, zoomAction)
			}

			if len(actions) > 0 {
				for idx, action := range actions {
					action.ActionId = idx
				}
				actionGroup := &wpml.ActionGroup{
					ActionGroupId:         0,
					ActionGroupStartIndex: 0,
					ActionGroupEndIndex:   0,
					ActionGroupMode:       wpml.ActionGroupModeSequence,
					ActionTrigger:         wpml.ActionTrigger{TriggerType: wpml.TriggerReachPoint},
				}
				actionGroup.Actions = actions
				placemark.ActionGroup = actionGroup
			}
		}

		folder.Placemarks = append(folder.Placemarks, placemark)
	}

	doc.Folders = append(doc.Folders, folder)

	return doc, nil
}

func (j *JobImpl) createCommandWaylineFile(ctx context.Context, jobID uint, jobName string, drone po.JobDronePO, commandDrone po.JobCommandDronePO) (*po.Wayline, error) {
	// 获取无人机变体信息
	var variation po.DroneVariation
	variations, err := j.modelRepo.SelectAllDroneVariation(ctx, nil)
	if err != nil {
		j.l.Error("获取无人机变体失败", slog.Any("error", err))
		return nil, err
	}
	for _, v := range variations {
		if v.ID == drone.VariationID {
			variation = v
			break
		}
	}
	j.l.Info("无人机变体", slog.Any("DroneVariation", variation))

	// 生成航线文件
	waylineDoc, err := j.generateCommandWayline(ctx, drone.PhysicalDroneID, variation, drone.TakeoffPoint, commandDrone)
	if err != nil {
		j.l.Error("生成航线文件失败", slog.Any("error", err))
		return nil, err
	}
	j.l.Info("航线文件已创建", "wayline", waylineDoc)

	// 生成航线模板和具体航线
	template, err := waylineDoc.GenerateXML()
	if err != nil {
		j.l.Error("生成航线模板失败", slog.Any("error", err))
		return nil, err
	}

	waylineDoc.Author = nil
	waylineDoc.CreateTime = nil
	waylineDoc.UpdateTime = nil
	for i := range waylineDoc.Folders {
		f := &waylineDoc.Folders[i]
		f.TemplateType = nil
		f.WaylineID = &i
		hMode := wpml.ExecuteHeightModeRelativeToStartPoint
		f.ExecuteHeightMode = &hMode
		for j := range f.Placemarks {
			p := &f.Placemarks[j]
			// 处理高度
			// if *p.UseGlobalHeight {
			// 	p.ExecuteHeight = f.GlobalHeight
			// } else {
			// 	p.ExecuteHeight = p.EllipsoidHeight
			// }
			// p.EllipsoidHeight = nil
			// p.Height = nil
			// p.UseGlobalHeight = nil
			// 处理速度
			if *p.UseGlobalSpeed {
				p.WaypointSpeed = f.AutoFlightSpeed
			}
			p.UseGlobalSpeed = nil
			// 处理 GlobalHeadingParam
			if *p.UseGlobalHeadingParam {
				p.WaypointHeadingParam = f.GlobalWaypointHeadingParam
			}
			p.UseGlobalHeadingParam = nil
			// 处理 WaypointTurnParam
			if *p.UseGlobalTurnParam {
				//p.WaypointTurnParam.WaypointTurnMode = *f.GlobalWaypointTurnMode
				p.WaypointTurnParam = &wpml.WaypointTurnParam{
					WaypointTurnMode: *f.GlobalWaypointTurnMode,
				}
			}
			p.UseGlobalTurnParam = nil
			// 处理 StraightLine
			if *p.UseStraightLine {
				p.UseStraightLine = f.GlobalUseStraightLine
			}
			p.UseStraightLine = nil
			// 处理 GimbalPitchMode
			if *f.GimbalPitchMode == wpml.GimbalPitchModeManual {
				p.WaypointGimbalHeadingParam = &wpml.WaypointGimbalHeadingParam{
					WaypointGimbalPitchAngle: 0,
					WaypointGimbalYawAngle:   0,
				}
			}
			workType := wpml.WaypointWorkTypeNone
			p.WaypointWorkType = &workType
		}
		// 擦除不需要的字段
		f.TemplateType = nil
		f.WaylineCoordinateSysParam = nil
		f.PayloadParam = nil
		f.GlobalWaypointTurnMode = nil
		f.GlobalUseStraightLine = nil
		f.GimbalPitchMode = nil
		f.GlobalHeight = nil
		f.GlobalWaypointHeadingParam = nil
	}
	waylineXML, err := waylineDoc.GenerateXML()
	if err != nil {
		j.l.Error("生成航线文件失败", slog.Any("error", err))
		return nil, err
	}

	// 创建kmz目录
	if err := os.MkdirAll("kmz", os.ModePerm); err != nil {
		j.l.Error("创建kmz目录失败", slog.Any("error", err))
		return nil, err
	}

	// 生成KMZ文件
	kmzFileName := "kmz/" + "job-" + strconv.Itoa(int(jobID)) + "-" + "drone-key-" + commandDrone.DroneKey + "-" + "drone-id-" + strconv.Itoa(int(drone.PhysicalDroneID)) + ".kmz"
	err = wpml.GenerateKMZ(kmzFileName, template, waylineXML)
	if err != nil {
		j.l.Error("生成KMZ文件失败", slog.Any("error", err))
		return nil, err
	}
	j.l.Info("KMZ文件已生成", "kmzFileName", kmzFileName)

	// 获取物理无人机信息
	physicalDrone, err := j.droneRepo.SelectByID(ctx, drone.PhysicalDroneID)
	if err != nil && err.Error() != "no realtime data" {
		j.l.Error("获取无人机信息失败", slog.Any("error", err))
		return nil, err
	}

	// 构建航线信息
	droneModelKey := strconv.Itoa(variation.DroneModel.Domain) + "-" + strconv.Itoa(variation.DroneModel.Type) + "-" + strconv.Itoa(variation.DroneModel.SubType)
	payloadModelKey := "1-" + strconv.Itoa(variation.Gimbals[0].Type) + "-" + strconv.Itoa(variation.Gimbals[0].SubType)
	waylinePO := po.Wayline{
		JobID:       jobID,
		JobDroneKey: commandDrone.DroneKey,
		DroneSN:     physicalDrone.SN,
		WaylineName: jobName + "-" + commandDrone.DroneKey,
		StartWaylinePoint: datatypes.NewJSONType(po.StartWaylinePoint{
			StartLatitude:  drone.TakeoffPoint.Lat,
			StartLongitude: drone.TakeoffPoint.Lng,
		}),
		DroneModelKey:    droneModelKey,
		PayloadModelKeys: []string{payloadModelKey},
	}

	// 清理临时文件
	defer func() {
		if err := os.Remove(kmzFileName); err != nil {
			j.l.Error("删除kmz临时文件失败", slog.Any("error", err))
		}
	}()

	// 保存航线文件到数据库
	savedWayline, err := j.jobRepo.SaveWayline(ctx, waylinePO, kmzFileName)
	if err != nil {
		j.l.Error("保存航线信息到数据库失败", slog.Any("error", err))
		return nil, err
	}
	j.l.Info("航线信息已保存到数据库", "wayline", commandDrone.DroneKey)

	return savedWayline, nil
}

func (j *JobImpl) generateCommandWayline(ctx context.Context, droneID uint, droneVariation po.DroneVariation, takeoffPoint po.JobTakeoffPointPO, commandDrone po.JobCommandDronePO) (wpml.Document, error) {
	droneModel := droneVariation.DroneModel
	gimbals := droneVariation.Gimbals
	drone, err := j.droneRepo.SelectByID(ctx, droneID)
	if err != nil && err.Error() != "no realtime data" {
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

	trueBool := wpml.BoolAsInt(true)          // 使用变量的地址
	templateType := wpml.TemplateTypeWaypoint // 先创建一个变量
	templateID := 0
	autoFlightSpeed := 1.0
	GimbalPitchMode := wpml.GimbalPitchModeManual
	globalWaypointHeadingParam := wpml.WaypointHeadingParam{
		WaypointHeadingMode: wpml.HeadingFollowWayline, // 航点航线
	}
	globalHeight := float64(commandDrone.Position.Altitude)
	executeHeightMode := wpml.ExecuteHeightModeRelativeToStartPoint // 执行高度模式相对于起点
	globalWaypointTurnMode := wpml.ToPointAndStopWithDiscontinuityCurvature
	folder := wpml.Folder{
		TemplateType: &templateType, // 使用变量的地址 // 航点航线
		TemplateID:   &templateID,
		WaylineCoordinateSysParam: &wpml.WaylineCoordinateSysParam{
			CoordinateMode: wpml.CoordinateModeWGS84,            // WGS84坐标系
			HeightMode:     wpml.HeightModeRelativeToStartPoint, // 高度相对于起点
		},
		AutoFlightSpeed:            autoFlightSpeed,             // 自动飞行速度
		GimbalPitchMode:            &GimbalPitchMode,            // 使用点设置
		GlobalHeight:               &globalHeight,               // 全局高度
		ExecuteHeightMode:          &executeHeightMode,          // 执行高度模式
		GlobalWaypointHeadingParam: &globalWaypointHeadingParam, // 航点航线
		GlobalWaypointTurnMode:     &globalWaypointTurnMode,     // 航点转弯模式
		GlobalUseStraightLine:      &trueBool,                   // 使用直线
	}

	lng, lat := coordinate.GCJ02ToWGS84(takeoffPoint.Lng, takeoffPoint.Lat)
	takeOffPlacemark := wpml.Placemark{
		Point:                 wpml.Point{Coordinates: wpml.FormatCoordinates(float64(lng), float64(lat))},
		Index:                 0,
		UseGlobalHeight:       &trueBool,
		Height:                &takeoffPoint.Altitude,
		EllipsoidHeight:       nil,
		ExecuteHeight:         &takeoffPoint.Altitude,
		UseGlobalSpeed:        &trueBool,
		UseGlobalHeadingParam: &trueBool,
		WaypointHeadingParam:  nil,
		UseGlobalTurnParam:    &trueBool,
		WaypointTurnParam:     nil,
		UseStraightLine:       &trueBool,
		GimbalPitchAngle:      -90,
		ActionGroup:           nil,
	}
	folder.Placemarks = append(folder.Placemarks, takeOffPlacemark)

	lng, lat = coordinate.GCJ02ToWGS84(commandDrone.Position.Lng, commandDrone.Position.Lat)
	commandPlacemark := wpml.Placemark{
		Point:                 wpml.Point{Coordinates: wpml.FormatCoordinates(float64(lng), float64(lat))},
		Index:                 0,
		UseGlobalHeight:       &trueBool,
		Height:                &commandDrone.Position.Altitude,
		EllipsoidHeight:       nil,
		ExecuteHeight:         &commandDrone.Position.Altitude,
		UseGlobalSpeed:        &trueBool,
		UseGlobalHeadingParam: &trueBool,
		WaypointHeadingParam:  nil,
		UseGlobalTurnParam:    &trueBool,
		WaypointTurnParam:     nil,
		UseStraightLine:       &trueBool,
		GimbalPitchAngle:      -90,
		ActionGroup:           nil,
	}
	hoverAction := wpml.Action{
		ActionType: wpml.ActionHover,
		ActionParams: &wpml.HoverParams{
			HoverTime: 3 * 60,
		},
	}
	actions := []wpml.Action{hoverAction}
	ag := &wpml.ActionGroup{
		ActionGroupId:         0,
		ActionGroupStartIndex: 0,
		ActionGroupEndIndex:   0,
		ActionGroupMode:       wpml.ActionGroupModeSequence,
		ActionTrigger:         wpml.ActionTrigger{TriggerType: wpml.TriggerReachPoint},
		Actions:               actions,
	}
	commandPlacemark.ActionGroup = ag
	folder.Placemarks = append(folder.Placemarks, commandPlacemark)

	doc.Folders = append(doc.Folders, folder)

	return doc, nil
}
