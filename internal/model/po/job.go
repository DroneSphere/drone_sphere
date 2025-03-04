package po

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	AreaID      uint                        `json:"area_id"`
	AlgoID      uint                        `json:"algo_id"`
	DroneSNList datatypes.JSONSlice[string] `json:"drone_sn_list"`
	DroneIDs    datatypes.JSONSlice[uint]   `json:"drone_ids"`
}

func (j Job) TableName() string {
	return "jobs"
}
