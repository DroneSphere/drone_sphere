package repo

import (
	"context"
	"encoding/base64"
	"log/slog"
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

func (j *JobDefaultRepo) Save(ctx context.Context, job *po.Job) error {
	if err := j.tx.Save(job).Error; err != nil {
		j.l.Error("Failed to create job", slog.Any("err", err))
		return err
	}
	return nil
}

func (j *JobDefaultRepo) FetchPOByID(ctx context.Context, id uint) (*po.Job, error) {
	var job po.Job
	if err := j.tx.First(&job, id).Error; err != nil {
		j.l.Error("Failed to fetch job by id", slog.Any("err", err))
		return nil, err
	}
	return &job, nil
}

func (j *JobDefaultRepo) FetchByID(ctx context.Context, id uint) (*entity.Job, error) {
	job, err := j.FetchPOByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return entity.NewJob(job), nil
}

func (j *JobDefaultRepo) SelectAll(ctx context.Context) ([]*entity.Job, error) {
	var jobs []*po.Job
	if err := j.tx.Find(&jobs).Error; err != nil {
		j.l.Error("Failed to fetch all jobs", slog.Any("err", err))
		return nil, err
	}

	var res []*entity.Job
	for _, job := range jobs {
		res = append(res, entity.NewJob(job))
	}

	return res, nil
}

func (j *JobDefaultRepo) SelectPhysicalDrones(ctx context.Context) ([]dto.PhysicalDrone, error) {
	var jsonStr string
	if err := j.tx.Raw(`
			SELECT JSON_ARRAYAGG(
				JSON_OBJECT(
					'id', d.drone_id,
					'sn', d.sn,
					'callsign', d.callsign,
					'model', JSON_OBJECT('id', dm.drone_model_id, 'name', dm.drone_model_name),
					'gimbals', dg.gimbals
				)
			) AS drone_data
			FROM drone.tb_drones d
			LEFT JOIN drone.tb_drone_models dm ON d.drone_model_id = dm.gateway_model_id
			LEFT JOIN (
				SELECT
					dg.drone_model_id AS drone_model_id,
					JSON_ARRAYAGG(JSON_OBJECT('id', gm.gimbal_model_id, 'name', gm.gimbal_model_name)) AS gimbals
				FROM drone.tb_drone_gimbal dg
				LEFT JOIN drone.tb_gimbal_models gm ON dg.gimbal_model_id = gm.gimbal_model_id
				GROUP BY dg.drone_model_id
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
		// if err := os.Remove(filename); err != nil {
		// 	j.l.Error("Failed to remove local KMZ file", slog.Any("err", err))
		// } else {
		// 	j.l.Info("Removed local KMZ file", slog.Any("filename", filename))
		// }
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
