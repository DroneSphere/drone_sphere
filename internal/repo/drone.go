package repo

import (
	"github.com/dronesphere/internal/model/entity"
	"gorm.io/gorm"
	"log/slog"
)

type DroneGormRepo struct {
	tx *gorm.DB
	l  *slog.Logger
}

func NewDroneGormRepo(db *gorm.DB, l *slog.Logger) *DroneGormRepo {
	return &DroneGormRepo{
		tx: db,
		l:  l,
	}
}

func (r *DroneGormRepo) ListAll() ([]entity.Drone, error) {
	return nil, nil
}
