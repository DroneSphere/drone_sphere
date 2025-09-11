package repo

import (
	"context"
	"log/slog"

	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type JobDefaultRepo struct {
	tx  *gorm.DB
	s3  *minio.Client
	rds *redis.Client
	l   *slog.Logger
}

func NewJobDefaultRepo(db *gorm.DB, s3 *minio.Client, rds *redis.Client, l *slog.Logger) *JobDefaultRepo {
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

func (j *JobDefaultRepo) SelectByID(ctx context.Context, id uint) (*po.Job, error) {
	var jobPO po.Job
	if err := j.tx.Where(
		"job_id = ?",
		id,
	).Find(&jobPO).Error; err != nil {
		j.l.Error("获取任务失败", slog.Any("err", err))
		return nil, err
	}
	return &jobPO, nil
}

func (j *JobDefaultRepo) SelectAll(ctx context.Context, jobName, areaName string, scheduleTimeStart, scheduleTimeEnd string) ([]po.Job, error) {
	j.l.Info("查询所有任务",
		slog.Any("jobName", jobName),
		slog.Any("areaName", areaName),
		slog.Any("scheduleTimeStart", scheduleTimeStart),
		slog.Any("scheduleTimeEnd", scheduleTimeEnd))

	query := j.tx.WithContext(ctx).Where("state = 0")

	if jobName != "" {
		query = query.Where("job_name LIKE ?", "%"+jobName+"%")
	}

	if areaName != "" {
		query = query.Where("area_name LIKE ?", "%"+areaName+"%")
	}

	// 添加时间筛选条件
	if scheduleTimeStart != "" {
		// 将开始日期补充为当天的00:00
		query = query.Where("schedule_time >= STR_TO_DATE(?, '%Y-%m-%d')", scheduleTimeStart)
	}

	if scheduleTimeEnd != "" {
		// 将结束日期补充为当天的23:59
		query = query.Where("schedule_time <= STR_TO_DATE(?, '%Y-%m-%d') + INTERVAL 1 DAY", scheduleTimeEnd)
	}

	var jobs []po.Job
	if err := query.Find(&jobs).Error; err != nil {
		j.l.Error("Failed to fetch all jobs", slog.Any("err", err))
		return nil, err
	}
	j.l.Info("Fetched all jobs", slog.Any("jobs", jobs))
	return jobs, nil
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
				) dg ON d.drone_model_id = dg.drone_model_id
			WHERE
				d.state = 0;
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

func (j *JobDefaultRepo) SaveWayline(ctx context.Context, wayline po.Wayline, kmzFile string) (*po.Wayline, error) {
	// 使用uuid生成唯一的 S3Key
	wayline.S3Key = uuid.New().String() + ".kmz"
	// 上传到 S3
	inf, err := j.s3.FPutObject(ctx, "kmz", wayline.S3Key, kmzFile, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		j.l.Error("Failed to upload wayline to S3", slog.Any("err", err))
		return nil, err
	}
	j.l.Info("Uploaded wayline to S3", slog.Any("info", inf))

	// 将 kmz 文件的 S3Key 保存到数据库
	j.l.Info("Saving wayline to database", slog.Any("wayline", wayline))
	if err := j.tx.WithContext(ctx).Save(&wayline).Error; err != nil {
		j.l.Error("Failed to save wayline to database", slog.Any("err", err))
		return nil, err
	}
	j.l.Info("Saved wayline to database", slog.Any("wayline", wayline))
	return &wayline, nil
}

func (j *JobDefaultRepo) SaveWaylineAndKmzKey(ctx context.Context, wayline po.Wayline, kmzKey string) (*po.Wayline, error) {
	wayline.S3Key = kmzKey

	// 将 kmz 文件的 S3Key 保存到数据库
	j.l.Info("Saving wayline to database", slog.Any("wayline", wayline))
	if err := j.tx.WithContext(ctx).Save(&wayline).Error; err != nil {
		j.l.Error("Failed to save wayline to database", slog.Any("err", err))
		return nil, err
	}
	j.l.Info("Saved wayline to database", slog.Any("wayline", wayline))
	return &wayline, nil
}
