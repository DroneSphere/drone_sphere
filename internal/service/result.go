package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
)

type ResultSvc interface {
	// Repo 返回结果仓库
	Repo() ResultRepo
	// Create 创建检测结果
	Create(ctx context.Context, result dto.CreateResultDTO) (uint, error)
	// CreateBatch 批量创建检测结果
	CreateBatch(ctx context.Context, results []dto.CreateResultDTO) ([]uint, error)
	// GetByID 获取单个检测结果
	GetByID(ctx context.Context, id uint) (*dto.ResultDetailDTO, error)
	// List 列出检测结果
	List(ctx context.Context, query dto.ResultQuery) ([]dto.ResultItemDTO, int64, error)
	// GetJobOptions 获取任务选项
	GetJobOptions(ctx context.Context) ([]dto.JobOption, error)
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
	// GetObjectTypeOptions 获取物体类型选项
	GetObjectTypeOptions(ctx context.Context) ([]dto.ObjectTypeOption, error)
	// DeleteByID 根据ID删除结果
	DeleteByID(ctx context.Context, id uint) error
	GetObjectTypeIDByType(ctx context.Context, objectType string) (uint, error)
}

type ResultImpl struct {
	repo      ResultRepo
	jobRepo   JobRepo
	droneRepo DroneRepo
	l         *slog.Logger
}

func NewResultImpl(repo ResultRepo, jobRepo JobRepo, droneRepo DroneRepo, l *slog.Logger) ResultSvc {
	return &ResultImpl{
		repo:      repo,
		jobRepo:   jobRepo,
		droneRepo: droneRepo,
		l:         l,
	}
}

// Repo 返回结果仓库
func (s *ResultImpl) Repo() ResultRepo {
	return s.repo
}

func (s *ResultImpl) Create(ctx context.Context, result dto.CreateResultDTO) (uint, error) {
	r := &po.Result{
		JobID:     result.JobID,
		WaylineID: result.WaylineID,
		// DroneID:            result.DroneID,
		// DetectObjectTypeID: result.ObjectTypeID, // 使用物体类型ID关联到物体类型表
		ObjectConfidence: result.ObjectConfidence,
		ObjectPosition:   result.Position,
		ObjectCoordinate: result.Coordinate,
		ImageUrl:         result.ImageUrl,
	}

	if result.DroneID != 0 && result.ObjectTypeID != 0 {
		// 关联无人机ID和物体类型ID
		r.DroneID = result.DroneID
		r.DetectObjectTypeID = result.ObjectTypeID
	} else if result.DroneSN != "" && result.ObjectType != "" {
		// 通过无人机序列号和物体类型名称获取ID
		drone, err := s.droneRepo.SelectBySN(ctx, result.DroneSN)
		if err != nil {
			s.l.Error("获取无人机信息失败", slog.Any("err", err))
			return 0, err
		}
		r.DroneID = drone.ID
		r.DetectObjectTypeID, err = s.repo.GetObjectTypeIDByType(ctx, result.ObjectType)
		if err != nil {
			s.l.Error("获取物体类型ID失败", slog.Any("err", err))
			return 0, err
		}
	} else {
		s.l.Error("缺少必要的参数", slog.Any("result", result))
		return 0, fmt.Errorf("缺少必要的参数")
	}

	if err := s.repo.Create(ctx, r); err != nil {
		return 0, err
	}

	return r.ID, nil
}

func (s *ResultImpl) CreateBatch(ctx context.Context, results []dto.CreateResultDTO) ([]uint, error) {
	var ids []uint
	for _, result := range results {
		id, err := s.Create(ctx, result)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
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

	drone, _ := s.droneRepo.SelectByID(ctx, r.DroneID)

	// 使用从关联的 DetectObjectType 中获取的物体类型信息
	return &dto.ResultDetailDTO{
		ID:               r.ID,
		JobID:            r.JobID,
		JobName:          job.Name,
		WaylineID:        r.WaylineID,
		DroneID:          r.DroneID,
		DroneCallsign:    drone.Callsign,
		DroneSN:          drone.SN,
		ObjectType:       r.DetectObjectType.Type,  // 从关联表中获取物体类型
		ObjectLabel:      r.DetectObjectType.Label, // 从关联表中获取物体标签
		ObjectConfidence: r.ObjectConfidence,
		Position:         r.ObjectPosition,
		Coordinate:       r.ObjectCoordinate,
		ImageUrl:         r.ImageUrl,
		CreatedAt:        r.CreatedTime.Format("2006-01-02 15:04:05"),
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
		// 获取无人机信息
		drone, _ := s.droneRepo.SelectByID(ctx, r.DroneID)

		// 从 ObjectPosition JSON字段中提取经纬度信息
		var coordinate struct {
			Lng float64 `json:"lng"` // 使用float64类型接收数值类型的经度
			Lat float64 `json:"lat"` // 使用float64类型接收数值类型的纬度
		}
		// 尝试解析JSON数据
		if err := json.Unmarshal(r.ObjectCoordinate, &coordinate); err != nil {
			s.l.Warn("解析位置信息失败", slog.Any("err", err), slog.Any("result_id", r.ID))
		}

		items = append(items, dto.ResultItemDTO{
			ID:            r.ID,
			JobName:       job.Name,
			DroneCallsign: drone.Callsign,
			TargetLabel:   r.DetectObjectType.Label,         // 从关联的 DetectObjectType 表中获取物体标签
			Lng:           formatCoordinate(coordinate.Lng), // 从解析后的结构体获取经度并格式化
			Lat:           formatCoordinate(coordinate.Lat), // 从解析后的结构体获取纬度并格式化
			ImageUrl:      r.ImageUrl,
			CreatedAt:     r.CreatedTime.Format("2006-01-02 15:04:05"), // 添加创建时间
		})
	}

	return items, total, nil
}

func (s *ResultImpl) GetJobOptions(ctx context.Context) ([]dto.JobOption, error) {
	return s.repo.GetJobOptions(ctx)
}

// formatTime 格式化时间戳
func (s *ResultImpl) formatTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
}

// formatCoordinate 将坐标浮点数转换为字符串
func formatCoordinate(value float64) string {
	return fmt.Sprintf("%.6f", value) // 保留6位小数的坐标字符串
}
