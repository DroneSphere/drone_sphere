package entity

import "github.com/dronesphere/internal/model/po"

type Job struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Area        SearchArea `json:"area"`
	Algo        DetectAlgo `json:"algo"`
	Drones      []Drone    `json:"drones"`
}

func NewJob(j *po.Job) *Job {
	var ds []Drone
	for _, id := range j.DroneIDs {
		ds = append(ds, Drone{ID: id})
	}
	return &Job{
		ID:          j.ID,
		Name:        j.Name,
		Description: j.Description,
		Area:        SearchArea{ID: j.AreaID},
		Drones:      ds,
	}
}
