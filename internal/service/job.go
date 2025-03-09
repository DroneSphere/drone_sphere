package service

import (
	"context"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"log/slog"
)

type (
	JobSvc interface {
		FetchByID(ctx context.Context, id uint) (*entity.Job, error)
		FetchAvailableAreas(ctx context.Context) ([]*entity.SearchArea, error)
		FetchAvailableDrones(ctx context.Context) ([]entity.Drone, error)
		CreateJob(ctx context.Context, name, description string, areaID uint) (uint, error)
		ModifyJob(ctx context.Context, id uint, name, description string, droneIDs []uint) (*entity.Job, error)
		FetchAll(ctx context.Context) ([]*entity.Job, error)
	}

	JobRepo interface {
		Save(ctx context.Context, job *po.Job) error
		FetchPOByID(ctx context.Context, id uint) (*po.Job, error)
		FetchByID(ctx context.Context, id uint) (*entity.Job, error)
		SelectAll(ctx context.Context) ([]*entity.Job, error)
	}
)

type JobImpl struct {
	jobRepo   JobRepo
	areaRepo  SearchAreaRepo
	droneRepo DroneRepo
	l         *slog.Logger
}

func NewJobImpl(jobRepo JobRepo, areaRepo SearchAreaRepo, droneRepo DroneRepo, l *slog.Logger) *JobImpl {
	return &JobImpl{
		jobRepo:   jobRepo,
		areaRepo:  areaRepo,
		droneRepo: droneRepo,
		l:         l,
	}
}

func (j *JobImpl) FetchAvailableAreas(ctx context.Context) ([]*entity.SearchArea, error) {
	areas, err := j.areaRepo.FetchAll(ctx)
	if err != nil {
		return nil, err
	}

	return areas, nil
}

func (j *JobImpl) FetchAvailableDrones(ctx context.Context) ([]entity.Drone, error) {
	drones, err := j.droneRepo.SelectAll(ctx)
	if err != nil {
		return nil, err
	}
	return drones, nil
}

func (j *JobImpl) FetchByID(ctx context.Context, id uint) (*entity.Job, error) {
	job, err := j.jobRepo.FetchByID(ctx, id)
	if err != nil {
		return nil, err
	}

	area, err := j.areaRepo.FetchByID(ctx, job.Area.ID)
	if err != nil {
		return nil, err
	}
	job.Area = *area

	var ids []uint
	for _, d := range job.Drones {
		ids = append(ids, d.ID)
	}
	drones, err := j.droneRepo.SelectAllByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	job.Drones = drones

	return job, nil
}

func (j *JobImpl) CreateJob(ctx context.Context, name, description string, areaID uint) (uint, error) {
	job := &po.Job{
		Name:        name,
		Description: description,
		AreaID:      areaID,
	}
	if err := j.jobRepo.Save(ctx, job); err != nil {
		return 0, err
	}
	return job.ID, nil
}

func (j *JobImpl) ModifyJob(ctx context.Context, id uint, name, description string, droneIDs []uint) (*entity.Job, error) {
	p, err := j.jobRepo.FetchPOByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		p.Name = name
	}
	if description != "" {
		p.Description = description
	}
	p.DroneIDs = droneIDs
	if err := j.jobRepo.Save(ctx, p); err != nil {
		return nil, err
	}
	return j.FetchByID(ctx, id)
}

func (j *JobImpl) FetchAll(ctx context.Context) ([]*entity.Job, error) {
	job, err := j.jobRepo.SelectAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, e := range job {
		area, err := j.areaRepo.FetchByID(ctx, e.Area.ID)
		if err != nil {
			return nil, err
		}
		e.Area = *area

		var ids []uint
		for _, d := range e.Drones {
			ids = append(ids, d.ID)
		}
		drones, err := j.droneRepo.SelectAllByID(ctx, ids)
		if err != nil {
			return nil, err
		}
		e.Drones = drones
	}
	return job, nil
}
