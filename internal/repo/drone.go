package repo

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/ro"
	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type DroneDefaultRepo struct {
	tx        *gorm.DB
	rds       *redis.Client
	l         *slog.Logger
	rdsPrefix string
}

func NewDroneGormRepo(db *gorm.DB, rds *redis.Client, l *slog.Logger) *DroneDefaultRepo {
	return &DroneDefaultRepo{
		tx:        db,
		rds:       rds,
		l:         l,
		rdsPrefix: "drone:",
	}
}

// GetDB 获取数据库连接
func (r *DroneDefaultRepo) GetDB() *gorm.DB {
	return r.tx
}

// SelectAll 列出所有无人机
func (r *DroneDefaultRepo) SelectAll(ctx context.Context) ([]entity.Drone, error) {
	var ds []entity.Drone
	var ps []po.Drone
	if err := r.tx.WithContext(ctx).
		Preload("DroneModel.Gimbals").
		Preload("DroneModel").
		Where("state = ?", 0).
		Find(&ps).Error; err != nil {
		r.l.Error(err.Error())
		panic(err)
	}
	r.l.Info("获取无人机持久化数据成功", slog.Any("po", ps))

	for _, p := range ps {
		// 获取无人机实时状态
		var rt ro.Drone
		rt, err := r.FetchStateBySN(ctx, p.SN)
		if err != nil {
			r.l.Info("实时数据获取失败", slog.Any("sn", p.SN), slog.Any("err", err))
		} else {
			r.l.Info("实时状态获取成功", slog.Any("sn", p.SN), slog.Any("rt", rt))
		}
		// 装配无人机实体
		e := entity.NewDrone(&p, &rt)
		ds = append(ds, *e)
	}
	r.l.Info("获取无人机列表成功", slog.Any("entity", ds))
	return ds, nil
}

// Save 保存无人机信息
func (r *DroneDefaultRepo) Save(ctx context.Context, d entity.Drone) error {
	err := r.tx.Where("sn = ?", d.SN).First(&po.Drone{}).Error
	if err == nil {
		slog.Info("记录已存在", slog.Any("drone", d))
	} else {
		var p po.Drone
		_ = copier.Copy(&p, d)
		if err := r.tx.Save(&p).Error; err != nil {
			r.l.Error("记录保存失败", slog.Any("drone", d), slog.Any("err", err))
			return err
		}
		slog.Info("记录保存成功", slog.Any("drone", d))
	}

	return nil
}

// SelectBySN 根据SN获取无人机实体
func (r *DroneDefaultRepo) SelectBySN(ctx context.Context, sn string) (entity.Drone, error) {
	var pp po.Drone
	var rr ro.Drone
	if err := r.tx.WithContext(ctx).
		Preload("DroneModel").
		Where("sn = ?", sn).First(&pp).Error; err != nil {
		r.l.Error("持久化数据获取失败", slog.Any("sn", sn), slog.Any("err", err))
		return entity.Drone{}, err
	}
	rr, err := r.FetchStateBySN(ctx, sn)
	if err != nil {
		r.l.Error("实时数据获取失败", slog.Any("sn", sn), slog.Any("err", err))
	}
	return *entity.NewDrone(&pp, &rr), err
}

const ErrNoRTData = "no realtime data"

// FetchStateBySN 根据SN获取无人机实时状态
func (r *DroneDefaultRepo) FetchStateBySN(ctx context.Context, sn string) (ro.Drone, error) {
	var rd ro.Drone
	r.l.Debug("获取实时数据", slog.Any("sn", sn), slog.Any("rdsPrefix", r.rdsPrefix))
	t, err := r.rds.JSONGet(ctx, r.rdsPrefix+sn, ".").Result()
	if err != nil {
		r.l.Error("实时数据获取失败", slog.Any("sn", sn), slog.Any("err", err))
		return ro.Drone{}, err
	}
	if t == "" {
		r.l.Error("实时数据为空", slog.Any("sn", sn))
		return rd, errors.New(ErrNoRTData)
	}
	_ = sonic.UnmarshalString(t, &rd)
	r.l.Debug("实时数据获取成功", slog.Any("sn", sn), slog.Any("rd", rd))

	return rd, nil
}

// SaveState 保存无人机实时状态
func (r *DroneDefaultRepo) SaveState(ctx context.Context, state ro.Drone) error {
	droneKey := r.rdsPrefix + state.SN
	r.l.Debug("保存实时状态", slog.Any("droneKey", droneKey), slog.Any("state", state))
	if err := r.rds.JSONSet(ctx, droneKey, ".", state).Err(); err != nil {
		r.l.Error("保存实时状态失败", slog.Any("err", err))
		return err
	}
	return nil
}

// SelectAllByID 根据 ID 列出所有无人机
func (r *DroneDefaultRepo) SelectAllByID(ctx context.Context, ids []uint) ([]entity.Drone, error) {
	var drones []entity.Drone
	for _, id := range ids {
		var pp po.Drone
		var rr ro.Drone
		if err := r.tx.Where("id = ?", id).First(&pp).Error; err != nil {
			r.l.Error("获取无人机持久化数据失败", slog.Any("id", id), slog.Any("err", err))
			continue
		}
		rr, err := r.FetchStateBySN(ctx, pp.SN)
		if err != nil {
			r.l.Error("获取无人机实时数据失败", slog.Any("id", id), slog.Any("err", err))
			continue
		}
		drones = append(drones, *entity.NewDrone(&pp, &rr))
	}
	return drones, nil
}

// UpdateCallsign 根据 ID 更新无人机信息
func (r *DroneDefaultRepo) UpdateCallsign(ctx context.Context, sn, callsign string) error {
	if err := r.tx.Model(&po.Drone{}).Where("sn = ?", sn).Update("callsign", callsign).Error; err != nil {
		r.l.Error("更新无人机呼号失败", slog.Any("sn", sn), slog.Any("callsign", callsign), slog.Any("err", err))
		return err
	}
	return nil
}

// FetchDroneModelOptions 查询无人机型号选项列表
// 此方法执行原生SQL查询，获取现有无人机对应的型号列表（去重）
// 主要用于前端下拉选择框的数据源
func (r *DroneDefaultRepo) FetchDroneModelOptions(ctx context.Context) ([]dto.DroneModelOption, error) {
	// 查询有效无人机的所有型号
	var models []dto.DroneModelOption
	if err := r.tx.WithContext(ctx).Raw(`
		SELECT DISTINCT dm.drone_model_id as id, dm.drone_model_name as name
		FROM tb_drones d
		JOIN tb_drone_models dm ON d.drone_model_id = dm.drone_model_id
		WHERE d.state = 0
		ORDER BY dm.drone_model_name
	`).Scan(&models).Error; err != nil {
		r.l.Error("获取无人机型号列表失败", slog.Any("error", err))
		return nil, err
	}

	return models, nil
}
