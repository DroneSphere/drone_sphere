package repo

import (
	"context"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
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
	return &DroneGormRepo{
		tx:        db,
		rds:       rds,
		l:         l,
		rdsPrefix: "drone:",
	}
}

// buildRCKeyPrefix 构建遥控器状态的redis key前缀
func (r *DroneGormRepo) buildRCKeyPrefix() string {
	return r.rdsPrefix + "rc:"
}

// buildDroneKeyPrefix 构建无人机状态的redis key前缀
func (r *DroneGormRepo) buildDroneKeyPrefix() string {
	return r.rdsPrefix + "drone:"
}

// ListAll 列出所有无人机
func (r *DroneGormRepo) ListAll(ctx context.Context) ([]entity.Drone, error) {
	var ps []persist.Drone
	if err := r.tx.Find(&ps).Error; err != nil {
		return nil, err
	}

	var ds []entity.Drone
	for _, p := range ps {
		ds = append(ds, p.Drone)
	}
	r.l.Info("List all drones", slog.Any("drones", ds))

	keys, err := r.rds.Keys(ctx, r.buildDroneKeyPrefix()+"*").Result()
	if err != nil {
		r.l.Error("Failed to get drone onlineSet from redis", slog.Any("err", err))
		return nil, err
	}
	r.l.Info("Drone keys", slog.Any("keys", keys))

	onlineSet := mapset.NewSet[string]()
	for _, k := range keys {
		onlineSet.Add(k[len(r.buildDroneKeyPrefix()):])
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

// RemoveDroneBySN 根据遥控器SN删除无人机
func (r *DroneGormRepo) RemoveDroneBySN(ctx context.Context, rc string) error {
	if err := r.rds.Del(ctx, r.buildRCKeyPrefix()+rc).Err(); err != nil {
		r.l.Error("remove drone status from redis failed", slog.Any("err", err))
		return err
	}
	return nil
}

// Save 保存无人机信息
func (r *DroneGormRepo) Save(ctx context.Context, d *entity.Drone, rc string) error {
	if err := r.tx.Where("sn = ?", d.SN).First(&persist.Drone{}).Error; err == nil {
		slog.Info("drone already exist", slog.Any("drone", d))
	} else {
		if err := r.tx.Save(&persist.Drone{Drone: *d}).Error; err != nil {
			return err
		}
		slog.Info("save drone to db success", slog.Any("drone", d))
	}

	slog.Info("Update realtime status", slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	rcKey := r.buildRCKeyPrefix() + rc
	if err := r.rds.HSet(ctx, rcKey, d.SN, entity.DroneStatusOnline).Err(); err != nil {
		r.l.Error("Save rc side status to redis failed", slog.Any("err", err))
		return err
	}
	r.rds.Expire(ctx, rcKey, 60*time.Second)
	slog.Info("Update RC Side status success", slog.Any("rc", rc), slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))

	droneKey := r.buildDroneKeyPrefix() + d.SN
	if err := r.rds.HSet(ctx, droneKey, "status", entity.DroneStatusOnline, "rc", rc).Err(); err != nil {
		r.l.Error("Save drone side status to redis failed", slog.Any("err", err))
		return err
	}
	r.rds.Expire(ctx, droneKey, 60*time.Second)
	slog.Info("Update Drone Side status success", slog.Any("sn", d.SN), slog.Any("status", entity.DroneStatusOnline))
	return nil
}

// SaveRealtimeStatus 保存无人机实时状态
func (r *DroneGormRepo) SaveRealtimeStatus(ctx context.Context, sn string, isOnline bool) error {
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

	if isOnline {
		r.rds.Expire(ctx, droneKey, 60*time.Second)
	} else {
		r.rds.Del(ctx, droneKey)
	}

	rcKey := r.buildRCKeyPrefix() + drone["rc"]
	if err := r.rds.HSet(ctx, rcKey, sn, isOnline).Err(); err != nil {
		r.l.Error("Save rc side status to redis failed", slog.Any("err", err))
		return err
	}
	r.rds.Expire(ctx, rcKey, 60*time.Second)

	return nil
}
