package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/pkg/coordinate"
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

func clusterFormatCoordinate(coord float64) float64 {
	return math.Round(coord*100000000) / 100000000 // 保留8位小数
}

// clusterResults 对结果进行空间聚类，保留置信度最高的点，并直接返回 dto.ResultItemDTO 列表
func (s *ResultImpl) clusterResults(items []dto.ResultItemDTO, clusterRadius float64) []dto.ResultItemDTO {
	if len(items) == 0 {
		return []dto.ResultItemDTO{}
	}

	var clustered []dto.ResultItemDTO
	// 记录原始项是否已被聚类，使用其原始 ID
	isClustered := make(map[uint]bool)

	for _, item := range items {
		if isClustered[item.ID] {
			continue // 已经处理过的点跳过
		}

		// 初始化当前簇的代表点，以当前点作为置信度最高的代表
		// 注意：这里创建一个副本，以避免直接修改迭代中的 `item`
		clusterRepresentative := item
		clusterRepresentative.Count = 1 // 初始化计数为1

		isClustered[item.ID] = true // 标记当前点已被处理

		// 遍历其他未聚类的点，寻找在半径内的点
		for _, otherItem := range items {
			if item.ID == otherItem.ID || isClustered[otherItem.ID] {
				continue // 跳过自身或已聚类的点
			}

			// 计算当前簇的代表点与 otherItem 之间的距离
			// 重要：这里距离计算的基准点是 `item`，即当前簇的"起始"点
			distance := coordinate.HaversineDistance(item.Lat, item.Lng, otherItem.Lat, otherItem.Lng)

			if distance <= clusterRadius {
				clusterRepresentative.Count++ // 增加计数
				// 比较创建时间，选择最新的
				if otherItem.Confidence >= clusterRepresentative.Confidence {
					clusterRepresentative.ID = otherItem.ID // 更新为最新的原始ID
					clusterRepresentative.JobName = otherItem.JobName
					clusterRepresentative.DroneCallsign = otherItem.DroneCallsign
					clusterRepresentative.TargetLabel = otherItem.TargetLabel
					clusterRepresentative.Lng = otherItem.Lng
					clusterRepresentative.Lat = otherItem.Lat
					clusterRepresentative.ImageUrl = otherItem.ImageUrl
					clusterRepresentative.CreatedAt = otherItem.CreatedAt
					clusterRepresentative.Confidence = otherItem.Confidence
				}

				isClustered[otherItem.ID] = true // 标记 otherItem 已被处理
			}
		}
		// 格式化最终的经纬度 (即置信度最高的点的经纬度)
		clusterRepresentative.Lng = clusterFormatCoordinate(clusterRepresentative.Lng)
		clusterRepresentative.Lat = clusterFormatCoordinate(clusterRepresentative.Lat)
		clustered = append(clustered, clusterRepresentative)
	}

	return clustered
}
func (s *ResultImpl) List(ctx context.Context, query dto.ResultQuery) ([]dto.ResultItemDTO, int64, error) {
	isJobSearch := query.Page <= 0 || query.PageSize <= 0
	s.l.Info("查询参数", slog.Any("query", query), slog.Bool("isJobSearch", isJobSearch))
	// 如果是任务搜索，查询所有结果并进行聚类
	// 任务搜索时，不进行分页，默认返回1000条
	// 普通搜索时，进行分页
	if query.Page <= 0 {
		query.Page = 1
		query.PageSize = 1000
	}
	results, total, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	s.l.Info("查询结果数", slog.Int("count", len(results)))

	// 使用 map 缓存任务和无人机信息，避免重复查询
	jobMap := make(map[uint]*po.Job)
	droneMap := make(map[uint]*po.Drone)

	var items []dto.ResultItemDTO
	for _, r := range results {
		job, ok := jobMap[r.JobID]
		if !ok {
			job, err = s.jobRepo.FetchPOByID(ctx, r.JobID)
			if err != nil {
				s.l.Error("获取任务信息失败", slog.Any("err", err))
				continue
			}
			jobMap[r.JobID] = job
		}

		// 获取无人机信息
		drone, ok := droneMap[r.DroneID]
		if !ok {
			drone, _ = s.droneRepo.SelectByIDV2(ctx, r.DroneID)
			droneMap[r.DroneID] = drone
		}

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
			TargetLabel:   r.DetectObjectType.Label, // 从关联的 DetectObjectType 表中获取物体标签
			Confidence:    r.ObjectConfidence,       // 从结果中获取置信度
			Lng:           coordinate.Lng,           // 从解析后的结构体获取经度并格式化
			Lat:           coordinate.Lat,           // 从解析后的结构体获取纬度并格式化
			ImageUrl:      r.ImageUrl,
			CreatedAt:     r.CreatedTime.Format("2006-01-02 15:04:05"), // 添加创建时间
		})
	}
	s.l.Info("结果项数", slog.Int("count", len(items)))
	if isJobSearch {
		// 定义矩形区域的四个顶点坐标（按顺序：左下、右下、右上、左上）
		// 注意：这些坐标需要您根据实际需求修改
		var boundaryPoints = []struct {
			Lng float64 `json:"lng"`
			Lat float64 `json:"lat"`
		}{
			{Lng: 117.04429209209722, Lat: 36.596760806499184}, // 左下角
			{Lng: 117.04501484159785, Lat: 36.596766320635865}, // 右下角
			{Lng: 117.04501074218794, Lat: 36.59774877796754},  // 右上角
			{Lng: 117.04428599840564, Lat: 36.59774926469286},  // 左上角
		}

		// 进行空间聚类
		clusteredItems := s.clusterResults(items, 6.0) // 5米半径
		var filteredItems []dto.ResultItemDTO
		// 遍历，手工剔除不符合要求（地理坐标以外、类型不符的）的项
		for i := range clusteredItems {
			s.l.Debug("聚类后结果项", slog.Any("item", clusteredItems[i]))
			//if clusteredItems[i].Lng < -180 || clusteredItems[i].Lng > 180 ||
			//	clusteredItems[i].Lat < -90 || clusteredItems[i].Lat > 90 {
			//	s.l.Warn("剔除坐标异常的结果项", slog.Any("item", clusteredItems[i]))
			//	clusteredItems = append(clusteredItems[:i], clusteredItems[i+1:]...)
			//	i--
			//	continue
			//}
			validTargetLabels := map[string]bool{
				"黄色坦克": true,
				"红色卡车": true,
			}
			if validTargetLabels[clusteredItems[i].TargetLabel] {
				// 检查是否在指定矩形区域内
				if s.isPointInRectangle(clusteredItems[i].Lng, clusteredItems[i].Lat, boundaryPoints) {
					s.l.Warn("符合要求的结果项", slog.Any("item", clusteredItems[i]))
					filteredItems = append(filteredItems, clusteredItems[i])
				} else {
					s.l.Info("目标不在指定区域内，已过滤", slog.Any("item", clusteredItems[i]))
				}
				continue
			}
		}
		if len(filteredItems) == 0 {
			return []dto.ResultItemDTO{}, 0, nil
		}
		s.l.Info("聚类后结果项数", slog.Int("count", len(filteredItems)))

		return filteredItems, int64(len(filteredItems)), nil
	} else {
		// 普通搜索，直接返回分页结果
		return items, total, nil
	}
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

// isPointInRectangle 检查点是否在矩形区域内
// 使用射线法判断点是否在任意四边形内
func (s *ResultImpl) isPointInRectangle(lng, lat float64, boundaryPoints []struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}) bool {
	if len(boundaryPoints) != 4 {
		s.l.Error("边界点数量不正确，需要4个点", slog.Int("count", len(boundaryPoints)))
		return false
	}

	// 使用射线法判断点是否在多边形内
	intersections := 0
	n := len(boundaryPoints)

	for i := 0; i < n; i++ {
		j := (i + 1) % n

		// 检查射线是否与边相交
		if ((boundaryPoints[i].Lat > lat) != (boundaryPoints[j].Lat > lat)) &&
			(lng < (boundaryPoints[j].Lng-boundaryPoints[i].Lng)*(lat-boundaryPoints[i].Lat)/(boundaryPoints[j].Lat-boundaryPoints[i].Lat)+boundaryPoints[i].Lng) {
			intersections++
		}
	}

	// 如果交点数为奇数，则点在多边形内
	isInside := intersections%2 == 1

	s.l.Debug("点在矩形区域内判断",
		slog.Float64("lng", lng),
		slog.Float64("lat", lat),
		slog.Bool("is_inside", isInside),
		slog.Int("intersections", intersections))

	return isInside
}
