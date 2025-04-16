package dto

import (
	"gorm.io/datatypes"
)

// ResultQuery 查询结果的参数
type ResultQuery struct {
	JobName      string `query:"job_name"`       // 任务名称,可选
	ObjectTypeID uint   `query:"object_type_id"` // 物体类型ID,可选
	Page         int    `query:"page"`           // 页码
	PageSize     int    `query:"page_size"`      // 每页数量
}

// ResultItemDTO 结果列表项
type ResultItemDTO struct {
	ID          uint   `json:"id"`           // 结果ID
	JobName     string `json:"job_name"`     // 任务名称
	TargetLabel string `json:"target_label"` // 检测目标分类
	Lng         string `json:"lng"`          // 经度
	Lat         string `json:"lat"`          // 纬度
	CreatedAt   string `json:"created_at"`   // 检测时间
	ImageUrl    string `json:"image_url"`    // 图片URL
}

// ResultDetailDTO 结果详情
type ResultDetailDTO struct {
	ID               uint           `json:"id"`                // 结果ID
	JobID            uint           `json:"job_id"`            // 任务ID
	JobName          string         `json:"job_name"`          // 任务名称
	WaylineID        uint           `json:"wayline_id"`        // 航线ID
	DroneID          uint           `json:"drone_id"`          // 无人机ID
	ObjectType       int            `json:"object_type"`       // 物体类型
	ObjectLabel      string         `json:"object_label"`      // 物体标签
	ObjectConfidence float32        `json:"object_confidence"` // 物体置信度
	Position         datatypes.JSON `json:"position"`          // 位置信息(经纬度)
	Coordinate       datatypes.JSON `json:"coordinate"`        // 坐标信息
	CreatedAt        string         `json:"created_at"`        // 创建时间
	ImageUrl         string         `json:"image_url"`         // 图片URL
}

// CreateResultDTO 创建结果的请求
type CreateResultDTO struct {
	JobID            uint           `json:"job_id"`            // 任务ID
	WaylineID        uint           `json:"wayline_id"`        // 航线ID
	DroneID          uint           `json:"drone_id"`          // 无人机ID
	ObjectTypeID     uint           `json:"object_type_id"`    // 物体类型ID
	ObjectConfidence float32        `json:"object_confidence"` // 物体置信度
	Position         datatypes.JSON `json:"position"`          // 位置信息
	Coordinate       datatypes.JSON `json:"coordinate"`        // 坐标信息
	ImageUrl         string         `json:"image_url"`         // 图片URL
}

// ObjectTypeOption 物体类型选项
type ObjectTypeOption struct {
	ID    uint   `json:"id"`    // 类型值
	Type  int    `json:"type"`  // 类型ID
	Label string `json:"label"` // 类型标签
}

// JobOption 任务选项
type JobOption struct {
	ID   uint   `json:"id"`   // 任务ID
	Name string `json:"name"` // 任务名称
}
