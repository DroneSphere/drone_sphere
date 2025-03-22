package repo

import (
	"context"
	"log/slog"

	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type JobDefaultRepo struct {
	tx  *gorm.DB
	rds *redis.Client
	l   *slog.Logger
}

func NewJobDefaultRepo(db *gorm.DB, rds *redis.Client, l *slog.Logger) *JobDefaultRepo {
	// if err := db.AutoMigrate(&po.Job{}); err != nil {
	// 	l.Error("Failed to auto migrate Drone", slog.Any("err", err))
	// 	panic(err)
	// }
	return &JobDefaultRepo{
		tx:  db,
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
		WITH t_gimbal AS (
			SELECT 
				drone_models.id AS drone_id, 
				gimbal_models.id AS gimbal_id, 
				gimbal_models.name AS gimbal_name
			FROM drone_gimbal
			LEFT JOIN drone_models ON drone_models.id = drone_gimbal.drone_model_id
			LEFT JOIN gimbal_models ON drone_gimbal.gimbal_model_id = gimbal_models.id
		),
		drone_data AS (
			SELECT 
				drones.id,
				drones.sn,
				drones.callsign,
				json_build_object(
					'id', drone_models.id,
					'name', drone_models.name
				) AS model,
				CASE
					WHEN count(t_gimbal) > 0 THEN
						json_agg(
							json_build_object(
								'id', t_gimbal.gimbal_id,
								'name', t_gimbal.gimbal_name
							)
						)
					ELSE NULL
				END AS gimbal
			FROM drones
			LEFT JOIN drone_models ON drones.model_id = drone_models.id
			LEFT JOIN t_gimbal ON t_gimbal.drone_id = drones.id
			GROUP BY drones.id, drones.sn, drones.callsign, drone_models.id, drone_models.name
			ORDER BY drones.callsign
		)

		SELECT json_agg(drone_data) FROM drone_data
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
