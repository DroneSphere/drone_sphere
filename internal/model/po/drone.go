package po

import (
	"github.com/dronesphere/internal/model/entity"
	"gorm.io/gorm"
)

type ORMDrone struct {
	gorm.Model
	entity.Drone
}

func (ORMDrone) TableName() string {
	return "drones"
}

type RTDrone struct {
	SN   string `json:"sn" redis:"sn"`
	RCSN string `json:"rcsn" redis:"rcsn"`
}
