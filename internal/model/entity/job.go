package entity

import (
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
)

type Job struct {
	ID          uint                     `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Area        Area                     `json:"area"`
	Algo        DetectAlgo               `json:"algo"`
	Drones      []dto.JobCreationDrone   `json:"drones"`
	Waylines    []dto.JobCreationWayline `json:"waylines"`
	Mappings    []dto.JobCreationMapping `json:"mappings"`
}

func NewJob(j *po.Job) *Job {
	return &Job{
		ID:          j.ID,
		Name:        j.Name,
		Description: j.Description,
		Area:        Area{ID: j.AreaID},
		Drones:      j.Drones,
		Waylines:    j.Waylines,
		Mappings:    j.Mappings,
	}
}
