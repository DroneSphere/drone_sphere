package po

import (
	"time"

	"github.com/dronesphere/internal/model/dto"
	"gorm.io/datatypes"
)

type Job struct {
	ID           uint                                        `json:"job_id" gorm:"primaryKey;column:job_id"`
	CreatedTime  time.Time                                   `json:"created_time" gorm:"autoCreateTime;column:created_time"`
	UpdatedTime  time.Time                                   `json:"updated_time" gorm:"autoUpdateTime;column:updated_time"`
	State        int                                         `json:"state" gorm:"default:0;column:state"` // -1: deleted, 0: active
	Name         string                                      `json:"job_name" gorm:"column:job_name"`
	Description  string                                      `json:"job_description" gorm:"column:job_description"`
	AreaID       uint                                        `json:"area_id" gorm:"column:area_id"`
	ScheduleTime time.Time                                   `json:"schedule_time" gorm:"column:schedule_time"` // 任务计划执行时间
	Drones       datatypes.JSONSlice[dto.JobCreationDrone]   `json:"drones" gorm:"column:drones"`
	Waylines     datatypes.JSONSlice[dto.JobCreationWayline] `json:"waylines" gorm:"column:waylines"`
	Mappings     datatypes.JSONSlice[dto.JobCreationMapping] `json:"mappings" gorm:"column:mappings"`
}

func (j Job) TableName() string {
	return "tb_jobs" // 添加 tb_ 前缀到表名
}
