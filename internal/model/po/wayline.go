package po

import (
	"time"

	"gorm.io/datatypes"
)

type Wayline struct {
	ID                uint                                  `json:"id" gorm:"primaryKey;column:wayline_id"`
	CreatedTime       time.Time                             `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime       time.Time                             `json:"updated_time" gorm:"autoUpdateTime"`
	State             int                                   `json:"state" gorm:"default:0"` // -1: deleted, 0: active
	JobID             uint                                  `json:"job_id" gorm:"column:job_id"`
	JobDroneKey       string                                `json:"job_drone_key" gorm:"column:job_drone_key"`
	DroneSN           string                                `json:"drone_sn" gorm:"column:drone_sn"`
	WaylineName       string                                `json:"wayline_name" gorm:"column:wayline_name"`
	StartWaylinePoint datatypes.JSONType[StartWaylinePoint] `json:"start_wayline_point" gorm:"column:start_wayline_point;type:json"`
	DroneModelKey     string                                `json:"drone_model_key" gorm:"column:drone_model_key"`
	PayloadModelKeys  datatypes.JSONSlice[string]           `json:"payload_model_keys" gorm:"column:payload_model_keys;type:json"`
	TemplateTypes     datatypes.JSONSlice[int]              `json:"template_types" gorm:"column:template_types;type:json"`
	S3Key             string                                `json:"s3_key" gorm:"column:s3_key"`
}

type StartWaylinePoint struct {
	StartLatitude  float64 `json:"start_latitude" gorm:"column:start_latitude"`
	StartLongitude float64 `json:"start_longitude" gorm:"column:start_longitude"`
}

func (w Wayline) TableName() string {
	return "tb_waylines"
}
