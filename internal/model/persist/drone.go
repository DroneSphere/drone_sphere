package persist

import (
	"github.com/dronesphere/internal/model/entity"
	"gorm.io/gorm"
)

type Drone struct {
	gorm.Model
	entity.Drone
}
