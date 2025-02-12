package repo

import (
	"context"
	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/jinzhu/copier"
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
		r.l.Error("ListAll failed", slog.Any("err", err))
		return nil, err
	}
	r.l.Info("Drone ListAll", slog.Any("drones", ps))

	for _, p := range ps {
		r.l.Debug("包装结构体", slog.Any("SN", p.SN))
		// 依据 p.SN 获取实时数据
		var rt po.RTDrone
		rt, err := r.FetchRealtimeDrone(ctx, p.SN)
		if err != nil {
			r.l.Error("ListAll failed", slog.Any("err", err))
			continue
		}
		r.l.Info("实时状态获取成功", slog.Any("SN", p.SN), slog.Any("rt", rt))
		// 拷贝静态数据和实时数据
		var e entity.Drone
		if err := copier.Copy(&e, &p); err != nil {
			r.l.Error("数据拷贝失败", slog.Any("SN", p.SN), slog.Any("err", err))
			continue
		}
		if err := copier.Copy(&e.RTDrone, &rt); err != nil {
			r.l.Error("数据拷贝失败", slog.Any("SN", p.SN), slog.Any("err", err))
			continue
		}
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
		var p po.ORMDrone
		if err := copier.Copy(&p, d); err != nil {
			r.l.Error("Save copier failed", slog.Any("err", err))
			return err
		}
		if err := r.tx.Save(&p).Error; err != nil {
			return err
		}
		slog.Info("save drone to db success", slog.Any("drone", d))
	}

	if err := r.SaveRealtimeDrone(ctx, po.RTDrone{
		SN:           d.SN,
		OnlineStatus: true,
		RCSN:         rcsn,
	}); err != nil {
		r.l.Error("Save realtime drone failed", slog.Any("err", err))
	}

	if err := r.SaveRealtimeRC(ctx, po.RTRC{
		SN:           rcsn,
		OnlineStatus: true,
	}); err != nil {
		r.l.Error("Save realtime rc failed", slog.Any("err", err))
	}
	return nil
}

// FetchRealtimeDrone 获取无人机实时状态
func (r *DroneGormRepo) FetchRealtimeDrone(ctx context.Context, sn string) (po.RTDrone, error) {
	var dr po.RTDrone
	t, err := r.rds.JSONGet(ctx, r.buildDroneKeyPrefix()+sn, ".").Result()
	if err != nil {
		r.l.Error("获取 Redis 数据失败", slog.Any("SN", sn), slog.Any("err", err))
	}
	r.l.Debug("实时状态获取成功", slog.Any("SN", sn), slog.Any("t", t))
	if err := sonic.UnmarshalString(t, &dr); err != nil {
		r.l.Error("反序列化失败", slog.Any("SN", sn), slog.Any("err", err))
		return dr, err
	}
	r.l.Debug("反序列化成功", slog.Any("SN", sn), slog.Any("dr", dr))

	return dr, nil
}

// SaveRealtimeDrone 保存无人机实时状态
func (r *DroneGormRepo) SaveRealtimeDrone(ctx context.Context, data po.RTDrone) error {
	droneKey := r.buildDroneKeyPrefix() + data.SN
	r.l.Debug("SaveRealtimeDrone", slog.Any("data", data), slog.Any("droneKey", droneKey))
	if err := r.rds.JSONSet(ctx, droneKey, ".", data).Err(); err != nil {
		r.l.Error("Save drone status to redis failed", slog.Any("err", err))
		return err
	}

	return nil
}

// SaveRealtimeRC 保存遥控器实时状态
func (r *DroneGormRepo) SaveRealtimeRC(ctx context.Context, data po.RTRC) error {
	rcKey := r.buildRCKeyPrefix() + data.SN
	if err := r.rds.HSet(ctx, rcKey, data).Err(); err != nil {
		r.l.Error("Save rc status to redis failed", slog.Any("err", err))
		return err
	}
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
