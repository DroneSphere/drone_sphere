package po

import (
	"time"

	"gorm.io/datatypes"
)

type Wayline struct {
	ID          uint      `json:"wayline_id" gorm:"primaryKey;column:wayline_id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime"`
	State       int       `json:"state" gorm:"default:0"` // -1: deleted, 0: active
	Name        string    `json:"wayline_name" gorm:"column:wayline_name"`
	UUID        string    `json:"uuid" gorm:"column:uuid"`
	// Username          string                                `json:"username" gorm:"column:create_user_id"`
	DroneModelKey     string                                `json:"drone_model_key" gorm:"column:drone_model_key"`
	PayloadModelKeys  datatypes.JSONSlice[string]           `json:"payload_model_keys" gorm:"column:payload_model_keys;type:json"`
	Favorited         bool                                  `json:"favorited" gorm:"column:is_favorited"`
	TemplateTypes     datatypes.JSONSlice[int]              `json:"template_types" gorm:"column:template_types;type:json"`
	ActionType        int                                   `json:"action_type" gorm:"column:action_type"`
	S3Key             string                                `json:"s3_key" gorm:"column:s3_key"`
	StartWaylinePoint datatypes.JSONType[StartWaylinePoint] `json:"start_wayline_point" gorm:"column:start_wayline_point;type:json"`
}

type StartWaylinePoint struct {
	StartLatitude  float64 `json:"start_latitude" gorm:"column:start_latitude"`
	StartLontitude float64 `json:"start_lontitude" gorm:"column:start_lontitude"`
}

// TableName 指定 Wayline 表名为 tb_waylines
func (w Wayline) TableName() string {
	return "tb_waylines"
}
