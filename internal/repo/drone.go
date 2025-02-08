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

func (r *DroneGormRepo) buildRCKeyPrefix() string {
	return r.rdsPrefix + "rc:"
}

func (r *DroneGormRepo) buildDroneKeyPrefix() string {
	return r.rdsPrefix + ":drone:"
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
	key := r.buildDroneKeyPrefix() + "*" // 获取所有无人机状态
	status, err := r.rds.HGetAll(nil, key).Result()
	if err != nil {
		r.l.Error("Failed to get drone status from redis", slog.Any("err", err))
		return nil, err
	}
	slog.Info("Drone status from redis", slog.Any("status", status))
	for i, d := range ds {
		if e, ok := status[d.SN]; ok {
			ds[i].Status = entity.DroneStatusOnline
			r.l.Info("Drone online", slog.Any("sn", d.SN), slog.Any("status", e))
		} else {
			ds[i].Status = entity.DroneStatusOffline
		}
	}
	return ds, nil
}

func (r *DroneGormRepo) RemoveDroneBySN(ctx context.Context, rc string) error {
	key := r.buildRCKeyPrefix() + rc
	err := r.rds.Del(ctx, key).Err()
	if err != nil {
		r.l.Error("remove drone status from redis failed", slog.Any("err", err))
		return err
	}
	return nil
}

func (r *DroneGormRepo) Save(ctx context.Context, d *entity.Drone, rc string) error {
	// 保存到数据库
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
	slog.Info("Update realtime status", slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	rcKey := r.buildRCKeyPrefix() + rc
	err := r.rds.HSet(ctx, rcKey, d.SN, entity.DroneStatusOnline).Err()
	if err != nil {
		r.l.Error("Save rc side status to redis failed", slog.Any("err", err))
		return err
	}
	slog.Info("Update RC Side status success", slog.Any("rc", rc), slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	droneKey := r.buildDroneKeyPrefix() + d.SN
	err = r.rds.HSet(ctx, droneKey, "status", entity.DroneStatusOnline).Err()
	if err != nil {
		r.l.Error("Save drone side status to redis failed", slog.Any("err", err))
		return err
	}
	slog.Info("Update Drone Side status success", slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	return nil
}
