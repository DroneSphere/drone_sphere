package repo

import (
	"context"
	"github.com/jinzhu/copier"
	"time"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
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
	//if err := db.AutoMigrate(&po.ORMDrone{}); err != nil {
	//	l.Error("Failed to auto migrate ORMDrone", slog.Any("err", err))
	//	panic(err)
	//}
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
	var ds []entity.Drone
	var ps []po.ORMDrone
	if err := r.tx.Find(&ps).Error; err != nil {
		return nil, err
	}

	for i, d := range ds {
		key := r.buildDroneKeyPrefix() + d.SN
		var rt po.RTDrone
		if err := r.rds.HGetAll(ctx, key).Scan(&rt); err != nil {
			r.l.Error("Get drone from redis failed", slog.Any("err", err))
			continue
		}
		var e entity.Drone
		if err := copier.Copy(&e, &ps[i]); err != nil {
			r.l.Error("ListAll copier failed", slog.Any("err", err))
			return nil, err
		}
		e.OnlineStatus = rt.OnlineStatus
		ds = append(ds, e)
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
func (r *DroneGormRepo) Save(ctx context.Context, d *entity.Drone, rcsn string) error {
	if err := r.tx.Where("sn = ?", d.SN).First(&po.ORMDrone{}).Error; err == nil {
		slog.Info("drone already exist", slog.Any("drone", d))
	} else {
		if err := r.tx.Save(&po.ORMDrone{Drone: *d}).Error; err != nil {
			return err
		}
		slog.Info("save drone to db success", slog.Any("drone", d))
	}

	if err := r.SaveRealtimeDrone(ctx, po.RTDrone{
		SN:           d.SN,
		OnlineStatus: false,
		RCSN:         rcsn,
	}); err != nil {
		r.l.Error("Save realtime drone failed", slog.Any("err", err))
	}

	if err := r.SaveRealtimeRC(ctx, po.RTRC{
		SN:           rcsn,
		OnlineStatus: false,
	}); err != nil {
		r.l.Error("Save realtime rc failed", slog.Any("err", err))
	}
	return nil
}

// FetchRealtimeDrone 获取无人机实时状态
func (r *DroneGormRepo) FetchRealtimeDrone(ctx context.Context, sn string) (po.RTDrone, error) {
	var dr po.RTDrone
	if err := r.rds.HGetAll(ctx, r.buildDroneKeyPrefix()+sn).Scan(&dr); err != nil {
		r.l.Error("Get drone from redis failed", slog.Any("err", err))
	}
	return dr, nil
}

// SaveRealtimeDrone 保存无人机实时状态
func (r *DroneGormRepo) SaveRealtimeDrone(ctx context.Context, data po.RTDrone) error {
	droneKey := r.buildDroneKeyPrefix() + data.SN
	r.l.Info("SaveRealtimeDrone", slog.Any("data", data), slog.Any("droneKey", droneKey))
	if err := r.rds.HSet(ctx, droneKey, data).Err(); err != nil {
		r.l.Error("Save drone status to redis failed", slog.Any("err", err))
		return err
	}
	r.rds.Expire(ctx, droneKey, 40*time.Second)
	r.l.Info("Save drone status to redis success", slog.Any("data", data))

	return nil
}

// SaveRealtimeRC 保存遥控器实时状态
func (r *DroneGormRepo) SaveRealtimeRC(ctx context.Context, data po.RTRC) error {
	rcKey := r.buildRCKeyPrefix() + data.SN
	if err := r.rds.HSet(ctx, rcKey, data).Err(); err != nil {
		r.l.Error("Save rc status to redis failed", slog.Any("err", err))
		return err
	}
	r.rds.Expire(ctx, rcKey, 40*time.Second)
	r.l.Info("Save rc status to redis success", slog.Any("data", data))
	return nil
}

func (r *DroneGormRepo) FetchDeviceTopoByWorkspace(ctx context.Context, workspace string) ([]entity.Drone, []entity.RC, error) {
	r.l.Info("FetchDeviceTopoByWorkspace", slog.Any("workspace", workspace))
	// 获取所有无人机和遥控器的实时状态
	allDroneKey := r.buildDroneKeyPrefix() + "*"
	allRCKey := r.buildRCKeyPrefix() + "*"
	var drones []po.RTDrone
	var rcs []po.RTRC
	keys, err := r.rds.Keys(ctx, allDroneKey).Result()
	if err != nil {
		r.l.Error("Failed to get drone onlineSet from redis", slog.Any("err", err))
		return nil, nil, err
	}
	r.l.Info("RTDrone keys", slog.Any("keys", keys))
	for _, k := range keys {
		var d po.RTDrone
		if err := r.rds.HGetAll(ctx, k).Scan(&d); err != nil {
			r.l.Error("Get drone from redis failed", slog.Any("err", err))
		}
		drones = append(drones, d)
	}
	keys, err = r.rds.Keys(ctx, allRCKey).Result()
	if err != nil {
		r.l.Error("Failed to get rc onlineSet from redis", slog.Any("err", err))
		return nil, nil, err

	}
	r.l.Info("RTRC keys", slog.Any("keys", keys))
	for _, k := range keys {
		var rc po.RTRC
		if err := r.rds.HGetAll(ctx, k).Scan(&rc); err != nil {
			r.l.Error("Get rc from redis failed", slog.Any("err", err))
		}
		rcs = append(rcs, rc)
	}

	// 拷贝数据
	var ds []entity.Drone
	for _, d := range drones {
		var e entity.Drone
		if err := copier.Copy(&e, &d); err != nil {
			r.l.Error("ListAll copier failed", slog.Any("err", err))
			return nil, nil, err
		}
		ds = append(ds, e)
	}
	var rs []entity.RC
	for _, rc := range rcs {
		var e entity.RC
		if err := copier.Copy(&e, &rc); err != nil {
			r.l.Error("ListAll copier failed", slog.Any("err", err))
			return nil, nil, err
		}
		rs = append(rs, e)
	}

	return ds, rs, nil
}
