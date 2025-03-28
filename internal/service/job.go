package service

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
)

type (
	JobSvc interface {
		Repo() JobRepo
		FetchByID(ctx context.Context, id uint) (*entity.Job, error)
		FetchAvailableAreas(ctx context.Context) ([]*entity.Area, error)
		FetchAvailableDrones(ctx context.Context) ([]entity.Drone, error)
		CreateJob(ctx context.Context, name, description string, areaID uint, drones []dto.JobCreationDrone, waylines []dto.JobCreationWayline, mappings []dto.JobCreationMapping) (uint, error)
		ModifyJob(ctx context.Context, id uint, name, description string, droneIDs []uint) (*entity.Job, error)
		FetchAll(ctx context.Context) ([]*entity.Job, error)
	}

	JobRepo interface {
		Save(ctx context.Context, job *po.Job) error
		FetchPOByID(ctx context.Context, id uint) (*po.Job, error)
		FetchByID(ctx context.Context, id uint) (*entity.Job, error)
		SelectAll(ctx context.Context) ([]*entity.Job, error)
		SelectPhysicalDrones(ctx context.Context) ([]dto.PhysicalDrone, error)
		CreateWaylineFile(ctx context.Context, name string, drone dto.JobCreationDrone, wayline dto.JobCreationWayline) (string, error)
	}
)

type JobImpl struct {
	jobRepo   JobRepo
	areaRepo  AreaRepo
	droneRepo DroneRepo
	l         *slog.Logger
}

func NewJobImpl(jobRepo JobRepo, areaRepo AreaRepo, droneRepo DroneRepo, l *slog.Logger) *JobImpl {
	return &JobImpl{
		jobRepo:   jobRepo,
		areaRepo:  areaRepo,
		droneRepo: droneRepo,
		l:         l,
	}
}

func (j *JobImpl) Repo() JobRepo {
	return j.jobRepo
}

func (j *JobImpl) FetchAvailableAreas(ctx context.Context) ([]*entity.Area, error) {
	areas, err := j.areaRepo.FetchAll(ctx, "")
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

	area, err := j.areaRepo.SelectByID(ctx, job.Area.ID)
	if err != nil {
		return nil, err
	}
	job.Area = *area

	return job, nil
}

func (j *JobImpl) CreateJob(ctx context.Context, name, description string, areaID uint, drones []dto.JobCreationDrone, waylines []dto.JobCreationWayline, mappings []dto.JobCreationMapping) (uint, error) {
	job := &po.Job{
		Name:        name,
		Description: description,
		AreaID:      areaID,
		Drones:      drones,
		Waylines:    waylines,
		Mappings:    mappings,
	}
	if err := j.jobRepo.Save(ctx, job); err != nil {
		return 0, err
	}
	j.l.Info("Job created", "job", job)
	// 逐个创建航线文件
	for _, w := range job.Waylines {
		var dr dto.JobCreationDrone
		for _, d := range job.Drones {
			if d.Key == w.DroneKey {
				dr = d
				break
			}
		}
		waylineKey, err := j.jobRepo.CreateWaylineFile(ctx, job.Name, dr, w)
		if err != nil {
			j.l.Error("CreateWaylineFile Error: ", slog.Any("error", err))
			return 0, err
		}
		j.l.Info("Wayline file created", "waylineKey", waylineKey)
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
		area, err := j.areaRepo.SelectByID(ctx, e.Area.ID)
		if err != nil {
			return nil, err
		}
		e.Area = *area
	}
	return job, nil
}
