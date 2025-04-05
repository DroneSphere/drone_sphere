package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
)

// 物体类型常量定义
const (
	ObjectTypeSoldier = 1 // 士兵
	ObjectTypeTank    = 2 // 坦克
	ObjectTypeCar     = 3 // 车辆
)

type ResultSvc interface {
	// Create 创建检测结果
	Create(ctx context.Context, result dto.CreateResultDTO) (uint, error)
	// GetByID 获取单个检测结果
	GetByID(ctx context.Context, id uint) (*dto.ResultDetailDTO, error)
	// List 列出检测结果
	List(ctx context.Context, query dto.ResultQuery) ([]dto.ResultItemDTO, int64, error)
	// GetJobOptions 获取任务选项
	GetJobOptions(ctx context.Context) ([]dto.JobOption, error)
	// GetObjectTypeOptions 获取物体类型选项
	GetObjectTypeOptions(ctx context.Context) []dto.ObjectTypeOption
}

type ResultRepo interface {
	// Create 创建检测结果
	Create(ctx context.Context, result *po.Result) error
	// GetByID 根据ID获取结果
	GetByID(ctx context.Context, id uint) (*po.Result, error)
	// List 列出结果
	List(ctx context.Context, query dto.ResultQuery) ([]po.Result, int64, error)
	// GetJobOptions 获取任务选项
	GetJobOptions(ctx context.Context) ([]dto.JobOption, error)
}

type ResultImpl struct {
	repo    ResultRepo
	jobRepo JobRepo
	l       *slog.Logger
}

func NewResultImpl(repo ResultRepo, jobRepo JobRepo, l *slog.Logger) ResultSvc {
	return &ResultImpl{
		repo:    repo,
		jobRepo: jobRepo,
		l:       l,
	}
}

func (s *ResultImpl) Create(ctx context.Context, result dto.CreateResultDTO) (uint, error) {
	r := &po.Result{
		JobID:            result.JobID,
		WaylineID:        result.WaylineID,
		DroneID:          result.DroneID,
		ObjectType:       result.ObjectType,
		ObjectLabel:      result.ObjectLabel,
		ObjectConfidence: result.ObjectConfidence,
		ObjectPosition:   result.Position,
		ObjectCoordinate: result.Coordinate,
	}

	if err := s.repo.Create(ctx, r); err != nil {
		return 0, err
	}

	return r.ID, nil
}

func (s *ResultImpl) GetByID(ctx context.Context, id uint) (*dto.ResultDetailDTO, error) {
	r, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 获取任务信息
	job, err := s.jobRepo.FetchPOByID(ctx, r.JobID)
	if err != nil {
		s.l.Error("获取任务信息失败", slog.Any("err", err))
		return nil, err
	}

	return &dto.ResultDetailDTO{
		ID:               r.ID,
		JobID:            r.JobID,
		JobName:          job.Name,
		WaylineID:        r.WaylineID,
		DroneID:          r.DroneID,
		ObjectType:       r.ObjectType,
		ObjectLabel:      r.ObjectLabel,
		ObjectConfidence: r.ObjectConfidence,
		Position:         r.ObjectPosition,
		Coordinate:       r.ObjectCoordinate,
	}, nil
}

func (s *ResultImpl) List(ctx context.Context, query dto.ResultQuery) ([]dto.ResultItemDTO, int64, error) {
	results, total, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	var items []dto.ResultItemDTO
	for _, r := range results {
		// 获取任务信息
		job, err := s.jobRepo.FetchPOByID(ctx, r.JobID)
		if err != nil {
			s.l.Error("获取任务信息失败", slog.Any("err", err))
			continue
		}

		items = append(items, dto.ResultItemDTO{
			ID:          r.ID,
			JobName:     job.Name,
			TargetLabel: r.ObjectLabel,
			// Lng:         r.ObjectPosition.,
			// Lat:         pos.Lat,
		})
	}

	return items, total, nil
}

func (s *ResultImpl) GetJobOptions(ctx context.Context) ([]dto.JobOption, error) {
	return s.repo.GetJobOptions(ctx)
}

func (s *ResultImpl) GetObjectTypeOptions(ctx context.Context) []dto.ObjectTypeOption {
	return []dto.ObjectTypeOption{
		{Value: ObjectTypeSoldier, Label: "士兵"},
		{Value: ObjectTypeTank, Label: "坦克"},
		{Value: ObjectTypeCar, Label: "车辆"},
	}
}

// formatTime 格式化时间戳
func (s *ResultImpl) formatTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}
