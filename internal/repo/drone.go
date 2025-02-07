package repo

import (
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/persist"
	"gorm.io/gorm"
	"log/slog"
)

type DroneGormRepo struct {
	tx *gorm.DB
	l  *slog.Logger
}

func NewDroneGormRepo(db *gorm.DB, l *slog.Logger) *DroneGormRepo {
	//err := db.AutoMigrate(&persist.Drone{})
	//if err != nil {
	//	l.Error("auto migrate drone table failed", slog.Any("err", err))
	//	panic(err)
	//}
	return &DroneGormRepo{
		tx: db,
		l:  l,
	}
}

func (r *DroneGormRepo) ListAll() ([]entity.Drone, error) {
	var ps []persist.Drone
	err := r.tx.Find(&ps).Error
	if err != nil {
		return nil, err
	}
	var ds []entity.Drone
	for _, p := range ps {
		d := p.Drone
		d.Status = entity.DroneStatusOffline
		ds = append(ds, d)
	}
	return ds, nil
}

func (r *DroneGormRepo) Save(d *entity.Drone, rc string) error {
	exist := r.tx.Where("sn = ?", d.SN).First(&persist.Drone{}).Error
	if exist == nil {
		return r.tx.Model(&persist.Drone{}).Where("sn = ?", d.SN).Updates(&persist.Drone{
			Drone: *d,
		}).Error
	}
	p := persist.Drone{
		Drone: *d,
	}
	return r.tx.Save(&p).Error
}
