package po

import "gorm.io/gorm"

type Job struct {
	gorm.Model
	Name        string   `json:"name"`
	Description string   `json:"description"`
	AreaID      uint     `json:"area_id"`
	AlgoID      uint     `json:"algo_id"`
	DroneSNList []string `json:"drone_sn_list"`
}
