package po

import (
	"github.com/dronesphere/internal/pkg/misc"
	"gorm.io/datatypes"
)

type Result struct {
	misc.BaseModel
	JobID            uint           `json:"job_id"`             // 任务ID
	WaylineID        uint           `json:"wayline_id"`         // 航线ID
	DroneID          uint           `json:"drone_id"`           // 无人机ID
	ObjectType       int            `json:"object_type" `       // 物体类型
	ObjectLabel      int            `json:"object_label" `      // 物体标签
	ObjectConfidence float32        `json:"object_confidence" ` // 物体置信度
	ObjectPosition   datatypes.JSON `json:"object_position"`    // 物体位置
	ObjectCoordinate datatypes.JSON `json:"object_coordinate"`  // 物体坐标
}

// TableName 指定 Result 表名为 tb_results
func (r Result) TableName() string {
	return "tb_results"
}
