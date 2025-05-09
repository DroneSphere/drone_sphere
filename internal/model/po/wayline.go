package po

import (
	"time"

	"gorm.io/datatypes"
)

type Wayline struct {
	ID                uint                                  `gorm:"primaryKey;column:wayline_id"`
	CreatedTime       time.Time                             `gorm:"autoCreateTime"`
	UpdatedTime       time.Time                             `gorm:"autoUpdateTime"`
	State             int                                   `gorm:"default:0"` // -1: deleted, 0: active
	JobID             uint                                  `gorm:"column:job_id"`
	JobDroneKey       string                                `gorm:"column:job_drone_key"`
	DroneSN           string                                `gorm:"column:drone_sn"`
	WaylineName       string                                `gorm:"column:wayline_name"`
	StartWaylinePoint datatypes.JSONType[StartWaylinePoint] `gorm:"column:start_wayline_point;type:json"`
	DroneModelKey     string                                `gorm:"column:drone_model_key"`
	PayloadModelKeys  datatypes.JSONSlice[string]           `gorm:"column:payload_model_keys;type:json"`
	TemplateTypes     datatypes.JSONSlice[int]              `gorm:"column:template_types;type:json"`
	S3Key             string                                `gorm:"column:s3_key"`
}

type StartWaylinePoint struct {
	StartLatitude  float64 `json:"start_latitude" gorm:"column:start_latitude"`
	StartLongitude float64 `json:"start_longitude" gorm:"column:start_longitude"`
}

func (w Wayline) TableName() string {
	return "tb_waylines"
}
