package repo

import (
	"context"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/persist"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
)

type DroneGormRepo struct {
	tx        *gorm.DB
	rds       *redis.Client
	l         *slog.Logger
	rdsPrefix string
}

func NewDroneGormRepo(db *gorm.DB, rds *redis.Client, l *slog.Logger) *DroneGormRepo {
	//err := db.AutoMigrate(&persist.Drone{})
	//if err != nil {
	//	l.Error("auto migrate drone table failed", slog.Any("err", err))
	//	panic(err)
	//}
	return &DroneGormRepo{
		tx:        db,
		rds:       rds,
		l:         l,
		rdsPrefix: "drone",
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
		ds = append(ds, d)
	}
	// 获取在线状态
	hkey := r.rdsPrefix + "status"
	status, err := r.rds.HGetAll(nil, hkey).Result()
	if err != nil {
		r.l.Error("get drone status from redis failed", slog.Any("err", err))
		return nil, err
	}
	slog.Info("status", slog.Any("status", status))
	for i, d := range ds {
		if _, ok := status[d.SN]; ok {
			ds[i].Status = entity.DroneStatusOnline
		} else {
			ds[i].Status = entity.DroneStatusOffline
		}
	}
	return ds, nil
}

func (r *DroneGormRepo) RemoveDroneBySN(ctx context.Context, rc string) error {
	key := r.rdsPrefix + ":" + rc
	err := r.rds.Del(ctx, key).Err()
	if err != nil {
		r.l.Error("remove drone status from redis failed", slog.Any("err", err))
		return err
	}
	return nil
}

func (r *DroneGormRepo) Save(ctx context.Context, d *entity.Drone, rc string) error {
	exist := r.tx.Where("sn = ?", d.SN).First(&persist.Drone{}).Error
	if exist == nil {
		slog.Info("drone already exist", slog.Any("drone", d))
	} else {
		p := persist.Drone{
			Drone: *d,
		}
		err := r.tx.Save(&p).Error
		if err != nil {
			return err
		}
		slog.Info("save drone to db success", slog.Any("drone", d))
	}

	// 保存状态到redis
	slog.Info("save drone status to redis", slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	key := r.rdsPrefix + ":" + rc
	err := r.rds.HSet(ctx, key, d.SN, entity.DroneStatusOnline).Err()
	if err != nil {
		r.l.Error("save drone status to redis failed", slog.Any("err", err))
		return err
	}
	slog.Info("save drone status to redis success", slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	return nil
}
