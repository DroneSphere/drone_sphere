package po

import (
	"github.com/dronesphere/internal/model/dto"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Name        string                                      `json:"name"`
	Description string                                      `json:"description"`
	AreaID      uint                                        `json:"area_id"`
	AlgoID      uint                                        `json:"algo_id"`       // Deprecated
	DroneSNList datatypes.JSONSlice[string]                 `json:"drone_sn_list"` // Deprecated
	DroneIDs    datatypes.JSONSlice[uint]                   `json:"drone_ids"`     // Deprecated
	Drones      datatypes.JSONSlice[dto.JobCreationDrone]   `json:"drones"`
	Waylines    datatypes.JSONSlice[dto.JobCreationWayline] `json:"waylines"`
	Mappings    datatypes.JSONSlice[dto.JobCreationMapping] `json:"mappings"`
}

func (j Job) TableName() string {
	return "jobs"
}
