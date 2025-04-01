package po

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Wayline struct {
	gorm.Model
	Name              string                                `json:"name" gorm:"column:name"`
	Username          string                                `json:"username" gorm:"column:username"`
	DroneModelKey     string                                `json:"drone_model_key" gorm:"column:drone_model_key"`
	PayloadModelKeys  datatypes.JSONSlice[string]           `json:"payload_model_keys" gorm:"column:payload_model_keys;type:json"`
	Favorited         bool                                  `json:"favorited" gorm:"column:favorited"`
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
