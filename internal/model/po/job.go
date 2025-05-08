package po

import (
	"time"

	"github.com/dronesphere/internal/model/vo"
	"gorm.io/datatypes"
)

type Job struct {
	ID            uint                                   `json:"job_id" gorm:"primaryKey;column:job_id"`
	CreatedTime   time.Time                              `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime   time.Time                              `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State         int                                    `json:"state" gorm:"default:0;column:state"` // -1: deleted, 0: active
	Name          string                                 `json:"job_name" gorm:"column:job_name"`
	Description   string                                 `json:"job_description" gorm:"column:job_description"`
	AreaID        uint                                   `json:"area_id" gorm:"column:area_id"`
	ScheduleTime  time.Time                              `json:"schedule_time" gorm:"column:schedule_time"` // 任务计划执行时间
	Drones        datatypes.JSONSlice[JobDronePO]        `json:"drones" gorm:"column:drones"`
	Waylines      datatypes.JSONSlice[JobWaylinePO]      `json:"waylines" gorm:"column:waylines"`
	CommandDrones datatypes.JSONSlice[JobCommandDronePO] `json:"command_drones" gorm:"column:command_drones"`
}

func (j Job) TableName() string {
	return "tb_jobs" // 添加 tb_ 前缀到表名
}

type JobTakeoffPointPO struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Altitude float64 `json:"altitude"`
}

type JobDronePO struct {
	// 列表中的唯一标识，格式为 {index}-{model_id}-{variation_id}
	Key         string `json:"key"`
	Index       int    `json:"index"`
	ModelID     uint   `json:"model_id"`
	VariationID uint   `json:"variation_id"`
	// 物理无人机ID
	PhysicalDroneID uint              `json:"physical_drone_id"`
	Color           string            `json:"color"`
	LensType        string            `json:"lens_type"`
	TakeoffPoint    JobTakeoffPointPO `json:"takeoff_point"`
}

type JobWaylinePO struct {
	DroneKey    string        `json:"drone_key"`
	Color       string        `json:"color"`
	Altitude    float64       `json:"altitude"`
	GimbalPitch float64       `json:"gimbal_pitch"`
	GimbalZoom  float64       `json:"gimbal_zoom"`
	Path        []vo.GeoPoint `json:"path"`
	Waypoints   []vo.GeoPoint `json:"waypoints"`
}

type JobCommandDronePO struct {
	DroneKey string      `json:"drone_key"`
	Position vo.GeoPoint `json:"position"`
}
