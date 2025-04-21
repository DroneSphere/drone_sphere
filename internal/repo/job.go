package repo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/pkg/coordinate"
	"github.com/dronesphere/pkg/wpml"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type JobDefaultRepo struct {
	tx  *gorm.DB
	s3  *minio.Client
	rds *redis.Client
	l   *slog.Logger
}

func NewJobDefaultRepo(db *gorm.DB, s3 *minio.Client, rds *redis.Client, l *slog.Logger) *JobDefaultRepo {
	// if err := db.AutoMigrate(&po.Job{}); err != nil {
	// 	l.Error("Failed to auto migrate Drone", slog.Any("err", err))
	// 	panic(err)
	// }
	return &JobDefaultRepo{
		tx:  db,
		s3:  s3,
		rds: rds,
		l:   l,
	}
}

func (j *JobDefaultRepo) DeleteByID(ctx context.Context, id uint) error {
	if err := j.tx.Model(&po.Job{}).
		Where("job_id = ?", id).
		Update("state", -1).Error; err != nil {
		j.l.Error("Failed to delete job", slog.Any("err", err))
		return err
	}
	return nil
}

func (j *JobDefaultRepo) Save(ctx context.Context, job *po.Job) error {
	if err := j.tx.Save(job).Error; err != nil {
		j.l.Error("Failed to create job", slog.Any("err", err))
		return err
	}
	return nil
}

func (j *JobDefaultRepo) FetchPOByID(ctx context.Context, id uint) (*po.Job, error) {
	var job po.Job
	if err := j.tx.
		WithContext(ctx).
		Where("job_id = ?", id).
		First(&job).Error; err != nil {
		j.l.Error("Failed to fetch job by id", slog.Any("err", err))
		return nil, err
	}
	return &job, nil
}

func (j *JobDefaultRepo) FetchByID(ctx context.Context, id uint) (*entity.Job, error) {
	var jsonStr string
	if err := j.tx.Raw(`
			SELECT
				JSON_OBJECT(
					'id', j.job_id,
					'name', j.job_name,
					'description', j.job_description,
					'schedule_time', DATE_FORMAT(j.schedule_time, '%Y-%m-%dT%H:%i:%sZ'),
					'area', JSON_OBJECT('id', a.area_id, 'name', a.area_name, 'description', a.area_description, 'points', a.area_points),
					'drones', JSON_ARRAYAGG(
						JSON_OBJECT(
							'id',  dm.drone_model_id,
							'key', JSON_EXTRACT(j.drones, '$[0].key'),
							'name', dm.drone_model_name,
							'description', dm.drone_model_description,
											'color', JSON_EXTRACT(j.drones, '$[0].color'),
							'index', JSON_EXTRACT(j.drones, '$[0].index'),                
							'variantion', JSON_EXTRACT(j.drones, '$[0].variantion'),
							'description', JSON_EXTRACT(j.drones, '$[0].description'),
							'variantion_id', JSON_EXTRACT(j.drones, '$[0].variantion_id')
						)
					),
					'waylines', j.waylines,
					'mappings', JSON_ARRAYAGG(
						JSON_OBJECT(
							'physical_drone_id', d.drone_id,
							'physical_drone_sn', d.sn,
							'selected_drone_key', JSON_EXTRACT(j.mappings, '$[0].selected_drone_key'),
							'physical_drone_callsign', d.callsign,
							'physical_drone_description', d.drone_description,
							'physical_drone_model_id',d.drone_model_id
						)
					)
				)
			FROM
				tb_jobs j
			LEFT JOIN
				tb_areas a ON j.area_id = a.area_id
			LEFT JOIN
				tb_drone_models dm ON JSON_EXTRACT(j.drones, '$[0].model_id') = dm.drone_model_id
			LEFT JOIN
				tb_drones d ON JSON_EXTRACT(j.mappings, '$[0].physical_drone_sn') = d.sn
			WHERE
				j.job_id = ?
			AND j.state = 0
			GROUP BY j.job_id`, id).Scan(&jsonStr).Error; err != nil {
		j.l.Error("获取任务失败", slog.Any("err", err))
		return nil, err
	}

	var job entity.Job
	if err := json.Unmarshal([]byte(jsonStr), &job); err != nil {
		j.l.Error("解析任务失败", slog.Any("err", err))
		return nil, err
	}

	return &job, nil
}

func (j *JobDefaultRepo) SelectAll(ctx context.Context, jobName, areaName string, scheduleTimeStart, scheduleTimeEnd string) ([]*entity.Job, error) {
	j.l.Info("查询所有任务",
		slog.Any("jobName", jobName),
		slog.Any("areaName", areaName),
		slog.Any("scheduleTimeStart", scheduleTimeStart),
		slog.Any("scheduleTimeEnd", scheduleTimeEnd))

	var jobs []struct {
		po.Job
		AreaName string `json:"area_name"`
	}

	// 基础SQL查询
	query := `SELECT j.*,a.area_name,a.area_description,a.center_lat,a.center_lng,a.area_points FROM tb_jobs j LEFT JOIN tb_areas a ON j.area_id=a.area_id WHERE j.state=0`

	// 参数列表
	params := []interface{}{}

	if jobName != "" {
		query += " AND j.job_name LIKE ?"
		params = append(params, "%"+jobName+"%")
	}

	if areaName != "" {
		query += " AND a.area_name LIKE ?"
		params = append(params, "%"+areaName+"%")
	}

	// 添加时间筛选条件
	if scheduleTimeStart != "" {
		// 将开始日期补充为当天的00:00
		query += " AND j.schedule_time >= STR_TO_DATE(?, '%Y-%m-%d')"
		params = append(params, scheduleTimeStart)
	}

	if scheduleTimeEnd != "" {
		// 将结束日期补充为当天的23:59
		query += " AND j.schedule_time <= STR_TO_DATE(?, '%Y-%m-%d') + INTERVAL 1 DAY"
		params = append(params, scheduleTimeEnd)
	}

	// 添加排序
	query += " ORDER BY j.job_id DESC"

	// 执行查询
	if err := j.tx.Raw(query, params...).Scan(&jobs).Error; err != nil {
		j.l.Error("获取所有任务失败", slog.Any("err", err))
		return nil, err
	}

	// 创建任务实体列表
	var jobEntities []*entity.Job
	for _, job := range jobs {
		jobEntity := entity.NewJob(&job.Job)
		jobEntity.Area.Name = job.AreaName
		jobEntities = append(jobEntities, jobEntity)
	}

	// 如果有任务，则收集所有物理无人机ID
	if len(jobEntities) > 0 {
		// 收集所有mappings中的无人机ID
		var allDroneIDs []uint
		droneIDMap := make(map[uint]bool)

		for _, job := range jobEntities {
			for _, mapping := range job.Mappings {
				if !droneIDMap[mapping.PhysicalDroneID] {
					allDroneIDs = append(allDroneIDs, mapping.PhysicalDroneID)
					droneIDMap[mapping.PhysicalDroneID] = true
				}
			}
		}

		// 如果存在物理无人机ID，则批量查询无人机信息
		if len(allDroneIDs) > 0 {
			// 定义无人机信息结构体，确保与后面使用的结构体类型一致
			type DroneInfo struct {
				DroneID          uint   `gorm:"column:drone_id"`
				SN               string `gorm:"column:sn"`
				Callsign         string `gorm:"column:callsign"`
				DroneDescription string `gorm:"column:drone_description"`
				DroneModelID     uint   `gorm:"column:drone_model_id"`
				DroneModelName   string `gorm:"column:drone_model_name"`
				ModelDescription string `gorm:"column:drone_model_description"`
				ModelDomain      int    `gorm:"column:drone_model_domain"`
				ModelType        int    `gorm:"column:drone_model_type"`
				ModelSubType     int    `gorm:"column:drone_model_sub_type"`
				IsRTKAvailable   int    `gorm:"column:is_rtk_available"`
			}

			var droneInfos []DroneInfo

			if err := j.tx.Raw(
				`
				SELECT 
					d.drone_id, 
					d.sn, 
					d.callsign, 
					d.drone_description,
					d.drone_model_id,
					dm.drone_model_name, 
					dm.drone_model_description, 
					dm.drone_model_domain, 
					dm.drone_model_type, 
					dm.drone_model_sub_type, 
					dm.is_rtk_available
				FROM 
					tb_drones d
				LEFT JOIN 
					tb_drone_models dm ON d.drone_model_id = dm.drone_model_id
				WHERE 
					d.drone_id IN (?)
				`, allDroneIDs).Scan(&droneInfos).Error; err != nil {
				j.l.Error("批量获取无人机详情失败", slog.Any("err", err))
				j.l.Warn("无法获取无人机详细信息，只返回基本任务数据")
			} else {
				// 将查询到的无人机详细信息建立映射关系
				droneInfoMap := make(map[uint]DroneInfo)

				for _, di := range droneInfos {
					droneInfoMap[di.DroneID] = di
				}

				// 遍历所有任务，更新mappings中的信息
				for jobIdx, job := range jobEntities {
					for mappingIdx, mapping := range job.Mappings {
						if info, exists := droneInfoMap[mapping.PhysicalDroneID]; exists {
							// 更新mappings中的信息
							jobEntities[jobIdx].Mappings[mappingIdx].PhysicalDroneSN = info.SN
							jobEntities[jobIdx].Mappings[mappingIdx].PhysicalDroneCallsign = info.Callsign
						}
					}
				}

				j.l.Info("已成功关联所有任务的无人机详细信息",
					slog.Any("任务数", len(jobEntities)),
					slog.Any("无人机数", len(droneInfos)))
			}
		}
	}

	return jobEntities, nil
}

func (j *JobDefaultRepo) SelectPhysicalDrones(ctx context.Context) ([]dto.PhysicalDrone, error) {
	var jsonStr string
	if err := j.tx.Raw(`
			SELECT
				JSON_ARRAYAGG(
					JSON_OBJECT(
						'id',
						d.drone_id,
						'sn',
						d.sn,
						'callsign',
						d.callsign,
						'model',
						JSON_OBJECT('id', dm.drone_model_id, 'name', dm.drone_model_name),
						'gimbals',
						dg.gimbals
					)
				) AS drone_data
			FROM
				drone.tb_drones d
				LEFT JOIN drone.tb_drone_models dm ON d.drone_model_id = dm.drone_model_id
				LEFT JOIN (
					SELECT
						dg.drone_model_id AS drone_model_id,
						JSON_ARRAYAGG(JSON_OBJECT('id', gm.gimbal_model_id, 'name', gm.gimbal_model_name)) AS gimbals
					FROM
						drone.tb_drone_gimbal dg
						LEFT JOIN drone.tb_gimbal_models gm ON dg.gimbal_model_id = gm.gimbal_model_id
					GROUP BY
						dg.drone_model_id
				) dg ON d.drone_model_id = dg.drone_model_id;
		`).Scan(&jsonStr).Error; err != nil {
		j.l.Error("Failed to fetch physical drones", slog.Any("err", err))
		return nil, err
	}
	j.l.Info("Fetched physical drones JSON", slog.Any("json", jsonStr))

	var drones []dto.PhysicalDrone
	err := sonic.Unmarshal([]byte(jsonStr), &drones)
	if err != nil {
		j.l.Error("Failed to unmarshal physical drones", slog.Any("err", err))
		return nil, err
	}
	j.l.Info("Fetched physical drones", slog.Any("drones", drones))

	return drones, nil
}

const contentType = "application/zip"

func (j *JobDefaultRepo) CreateWaylineFile(ctx context.Context, name string, drone dto.JobCreationDrone, wayline dto.JobCreationWayline) (string, error) {
	//  查询数据库获取无人机信息
	droneInfo := wpml.DroneInfo{
		DroneEnumValue:    wpml.DroneM3Series,
		DroneSubEnumValue: wpml.SubM3E,
	}
	payload := wpml.PayloadInfo{
		PayloadEnumValue:     wpml.PayloadM3E,
		PayloadSubEnumValue:  wpml.PayloadSubM3E,
		PayloadPositionIndex: 0,
	}

	builder := wpml.NewBuilder().Init("system").SetDefaultMissionConfig(droneInfo, payload)
	fBuilder := builder.Template.CreateFolder(wpml.TemplateTypeWaypoint, 0)
	for _, mark := range wayline.Points {
		fBuilder.AppendDefaultPlacemark(coordinate.GCJ02ToWGS84(mark.Lng, mark.Lat))
	}

	// 生成航线文件
	templateXML, err := builder.Template.GenerateXML()
	if err != nil {
		j.l.Error("Failed to generate template XML", slog.Any("err", err))
		return "", err
	}
	j.l.Info("Generated template XML", slog.Any("templateXML", templateXML))

	// 生成航线文件
	builder.GenerateWayline()
	waylineXML, err := builder.Wayline.GenerateXML()
	if err != nil {
		j.l.Error("Failed to generate wayline XML", slog.Any("err", err))
		return "", err
	}
	j.l.Info("Generated wayline XML", slog.Any("waylineXML", waylineXML))

	// 生成KMZ文件
	id := uuid.New()
	uuidBytes, _ := id.MarshalBinary()
	compressedUUID := make([]byte, 32) // buffer for base64 encoding
	base64.RawURLEncoding.Encode(compressedUUID, uuidBytes)
	filename := string(compressedUUID[:22]) + ".kmz"
	if err := wpml.GenerateKMZ(filename, templateXML, waylineXML); err != nil {
		j.l.Error("Failed to generate KMZ file", slog.Any("err", err))
		return "", err
	}
	j.l.Info("Generated KMZ file", slog.Any("filename", filename))

	// 删除本地文件
	defer func() {
		if err := os.Remove(filename); err != nil {
			j.l.Error("Failed to remove local KMZ file", slog.Any("err", err))
		} else {
			j.l.Info("Removed local KMZ file", slog.Any("filename", filename))
		}
	}()

	// 保存到 s3
	info, err := j.s3.FPutObject(ctx, "kmz", filename, filename, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		j.l.Error("Failed to save KMZ file to S3", slog.Any("err", err))
		return "", err
	}
	j.l.Info("Saved KMZ file to S3", slog.Any("info", info))

	po := &po.Wayline{
		Name: name,
		// Username:         "admin",
		DroneModelKey:    "0-" + strconv.Itoa(int(droneInfo.DroneEnumValue)) + "-" + strconv.Itoa(int(droneInfo.DroneSubEnumValue)),
		PayloadModelKeys: []string{"0-" + strconv.Itoa(int(payload.PayloadEnumValue)) + "-" + strconv.Itoa(int(payload.PayloadSubEnumValue))},
		Favorited:        false,
		TemplateTypes:    []int{0},
		ActionType:       0,
		S3Key:            filename,
		StartWaylinePoint: datatypes.NewJSONType(po.StartWaylinePoint{
			StartLatitude:  wayline.Points[0].Lat,
			StartLontitude: wayline.Points[0].Lng,
		}),
	}
	j.l.Info("Creating wayline PO", slog.Any("waylinePO", po))
	if err := j.tx.Create(po).Error; err != nil {
		j.l.Error("Failed to save wayline to database", slog.Any("err", err))
		return "", err
	}
	j.l.Info("Saved wayline to database", slog.Any("wayline", po))

	return filename, nil
}
