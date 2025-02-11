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
	SN           string `json:"sn" redis:"sn"`
	OnlineStatus bool   `json:"online_status" redis:"online_status"`
	RCSN         string `json:"rcsn" redis:"rcsn"`
}
