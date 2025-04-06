package entity

import (
	"time"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
)

type Job struct {
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	Area         Area                     `json:"area"`
	ScheduleTime time.Time                `json:"schedule_time"` // 任务计划执行时间
	Drones       []dto.JobCreationDrone   `json:"drones"`
	Waylines     []dto.JobCreationWayline `json:"waylines"`
	Mappings     []dto.JobCreationMapping `json:"mappings"`
}

func NewJob(j *po.Job) *Job {
	return &Job{
		ID:           j.ID,
		Name:         j.Name,
		Description:  j.Description,
		Area:         Area{ID: j.AreaID},
		ScheduleTime: j.ScheduleTime,
		Drones:       j.Drones,
		Waylines:     j.Waylines,
		Mappings:     j.Mappings,
	}
}
