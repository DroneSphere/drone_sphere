package repo

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/persist"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
	"time"
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
		rdsPrefix: "drone:",
	}
}

func (r *DroneGormRepo) buildRCKeyPrefix() string {
	return r.rdsPrefix + "rc:"
}

func (r *DroneGormRepo) buildDroneKeyPrefix() string {
	return r.rdsPrefix + "drone:"
}

func (r *DroneGormRepo) ListAll(ctx context.Context) ([]entity.Drone, error) {
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
	r.l.Info("List all drones", slog.Any("drones", ds))
	// 获取在线状态
	keys, err := r.rds.Keys(ctx, r.buildDroneKeyPrefix()+"*").Result()
	if err != nil {
		r.l.Error("Failed to get drone onlineSet from redis", slog.Any("err", err))
		return nil, err
	}
	r.l.Info("Drone keys", slog.Any("keys", keys))
	onlineSet := mapset.NewSet[string]()
	for _, k := range keys {
		sn := k[len(r.buildDroneKeyPrefix()):]
		onlineSet.Add(sn)
	}
	slog.Info("Drone onlineSet from redis", slog.Any("onlineSet", onlineSet))
	for i, d := range ds {
		if onlineSet.Contains(d.SN) {
			ds[i].Status = entity.DroneStatusOnline
			r.l.Info("Drone online", slog.Any("sn", d.SN))
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
	r.rds.Expire(ctx, rcKey, 60*time.Second)
	slog.Info("Update RC Side status success", slog.Any("rc", rc), slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	droneKey := r.buildDroneKeyPrefix() + d.SN
	err = r.rds.HSet(ctx, droneKey, "status", entity.DroneStatusOnline, "rc", rc).Err()
	if err != nil {
		r.l.Error("Save drone side status to redis failed", slog.Any("err", err))
		return err
	}
	r.rds.Expire(ctx, droneKey, 60*time.Second)
	slog.Info("Update Drone Side status success", slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	return nil
}

func (r *DroneGormRepo) SaveRealtimeStatus(ctx context.Context, sn string, isOnline bool) error {
	// 获取无人机信息
	droneKey := r.buildDroneKeyPrefix() + sn
	drone, err := r.rds.HGetAll(ctx, droneKey).Result()
	if err != nil {
		r.l.Error("Failed to get drone status from redis", slog.Any("err", err))
		return err
	}
	if len(drone) == 0 {
		r.l.Error("Drone not found", slog.Any("sn", sn))
		return nil
	}

	// 更新状态
	if isOnline {
		r.rds.Expire(ctx, droneKey, 60*time.Second)
	} else {
		r.rds.Del(ctx, droneKey)
	}

	// 更新遥控器状态
	rc := drone["rc"]
	rcKey := r.buildRCKeyPrefix() + rc
	err = r.rds.HSet(ctx, rcKey, sn, isOnline).Err()
	if err != nil {
		r.l.Error("Save rc side status to redis failed", slog.Any("err", err))
		return err
	}
	r.rds.Expire(ctx, rcKey, 60*time.Second)

	return nil
}
