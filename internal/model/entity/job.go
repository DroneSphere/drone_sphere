package entity

import (
	"time"

	"github.com/dronesphere/internal/model/po"
)

type Job struct {
	ID                      uint                          `json:"id"`
	Name                    string                        `json:"name"`
	Description             string                        `json:"description"`
	Area                    Area                          `json:"area"`
	ScheduleTime            time.Time                     `json:"schedule_time"` // 任务计划执行时间
	Drones                  []JobDrone                    `json:"drones"`
	Waylines                []po.JobWaylinePO             `json:"waylines"`
	CommandDrones           []po.JobCommandDronePO        `json:"command_drones"`
	WaylineGenerationParams po.JobWaylineGenerationParams `json:"wayline_generation_params"`
}

type JobDrone struct {
	po.JobDronePO
	DroneModel    DroneModel     `json:"drone_model"`
	GimbalModel   po.GimbalModel `json:"gimbal_model"`
	PhysicalDrone po.Drone       `json:"physical_drone"`
	Wayline       Wayline        `json:"wayline"`
}
