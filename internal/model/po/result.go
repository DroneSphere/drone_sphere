package po

import (
	"time"

	"gorm.io/datatypes"
)

type Result struct {
	ID                 uint             `json:"result_id" gorm:"primaryKey;column:result_id"`
	CreatedTime        time.Time        `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime        time.Time        `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State              int              `json:"state" gorm:"default:0"`                                              // -1: deleted, 0: active
	JobID              uint             `json:"job_id"`                                                              // 任务ID
	WaylineID          uint             `json:"wayline_id"`                                                          // 航线ID
	DroneID            uint             `json:"drone_id"`                                                            // 无人机ID
	DetectObjectType   DetectObjectType `json:"object_type_info" gorm:"foreignKey:DetectObjectTypeID;references:ID"` // 物体类型信息，使用DetectObjectTypeID关联到DetectObjectType表的ID字段
	DetectObjectTypeID uint             `json:"object_type_id" gorm:"column:object_type_id"`                         // 物体类型ID
	ObjectConfidence   float32          `json:"object_confidence" `                                                  // 物体置信度
	ObjectPosition     datatypes.JSON   `json:"object_position"`                                                     // 物体位置
	ObjectCoordinate   datatypes.JSON   `json:"object_coordinate"`                                                   // 物体坐标
	ImageUrl           string           `json:"image_url"`                                                           // 图片URL
}

// TableName 指定 Result 表名为 tb_results
func (r Result) TableName() string {
	return "tb_results"
}

// DetectObjectType 物体检测类型
type DetectObjectType struct {
	ID             uint      `json:"object_id" gorm:"primaryKey;column:object_id"`
	CreatedTime    time.Time `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime    time.Time `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State          int       `json:"state" gorm:"default:0"` // -1: deleted, 0: active
	Type           string    `json:"type"`                   // 物体类型
	Label          string    `json:"label"`                  // 物体标签
	AdditionalInfo string    `json:"additional_info"`        // 物体附加信息
}

func (d DetectObjectType) TableName() string {
	return "list_detect_object_type"
}
