package repo

import (
	"context"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
)

type JobDefaultRepo struct {
	tx  *gorm.DB
	rds *redis.Client
	l   *slog.Logger
}

func NewJobDefaultRepo(db *gorm.DB, rds *redis.Client, l *slog.Logger) *JobDefaultRepo {
	if err := db.AutoMigrate(&po.Job{}); err != nil {
		l.Error("Failed to auto migrate ORMDrone", slog.Any("err", err))
		panic(err)
	}
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
