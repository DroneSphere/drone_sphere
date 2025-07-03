package repo

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

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
// 参数说明:
// sn: 无人机序列号，用于精确匹配，可为空
// callsign: 无人机呼号，用于精确匹配，可为空
// modelID: 无人机型号ID，用于精确匹配，可为0
func (r *DroneDefaultRepo) SelectAll(ctx context.Context, sn string, callsign string, modelID uint, page, pageSize int) ([]entity.Drone, int64, error) {
	var ds []entity.Drone
	var ps []po.Drone

	// 构建查询条件
	query := r.tx.WithContext(ctx).
		Preload("DroneModel.Gimbals").
		Preload("DroneModel").
		Where("state = ?", 0).
		Order("created_time DESC")
	// 添加可选的筛选条件
	if sn != "" {
		// 将输入的 sn 转换为大写
		// 以便与数据库中的 sn 进行匹配
		query = query.Where("sn LIKE ?", "%"+strings.ToUpper(sn)+"%")
	}
	if callsign != "" {
		query = query.Where("callsign LIKE ?", "%"+callsign+"%")
	}
	if modelID > 0 {
		query = query.Where("drone_model_id = ?", modelID)
	}

	// 查询总数
	var total int64
	if err := query.Model(&po.Drone{}).Count(&total).Error; err != nil {
		r.l.Error("查询无人机总数失败", slog.Any("error", err))
		return nil, 0, err
	}
	r.l.Info("查询无人机总数成功", slog.Int64("total", total))
	if total == 0 {
		r.l.Info("无人机列表为空")
		return ds, 0, nil
	}

	// 添加分页条件
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// 执行查询
	if err := query.Find(&ps).Error; err != nil {
		r.l.Error("查询无人机列表失败", slog.Any("error", err))
		return ds, 0, err
	}
	r.l.Info("获取无人机持久化数据成功", slog.Any("po", ps))

	// 使用 goroutine 并行处理无人机数据
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, p := range ps {
		wg.Add(1)
		go func(drone po.Drone) {
			defer wg.Done()

			// 获取无人机实时状态
			var rt ro.Drone
			rt, err := r.FetchStateBySN(ctx, drone.SN)
			if err != nil {
				r.l.Info("实时数据获取失败", slog.Any("sn", drone.SN), slog.Any("err", err))
			} else {
				r.l.Info("实时状态获取成功", slog.Any("sn", drone.SN), slog.Any("rt", rt))
			}

			// 装配无人机实体
			e := entity.NewDrone(&drone, &rt)

			// 使用 mutex 保护共享资源
			mu.Lock()
			ds = append(ds, *e)
			mu.Unlock()
		}(p)
	}

	// 等待所有 goroutine 完成
	wg.Wait()
	r.l.Info("获取无人机列表成功", slog.Any("entity", ds))
	return ds, total, nil
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

// UpdateLiveInfoBySN 更新无人机直播相关信息，包括推拉流地址和当前视频流ID
func (r *DroneDefaultRepo) UpdateLiveInfoBySN(ctx context.Context, sn, pushRTMPUrl, pullRTMPUrl, videoID string) error {
	updates := map[string]interface{}{
		"live_push_rtmp_url":   pushRTMPUrl,
		"live_pull_webrtc_url": pullRTMPUrl,
		"current_video_id":     videoID, // 新增 current_video_id 的更新
	}
	return r.tx.WithContext(ctx).Model(&po.Drone{}).Where("state = 0 AND sn = ?", sn).Updates(updates).Error
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

// SelectByID 根据ID获取无人机实体
func (r *DroneDefaultRepo) SelectByID(ctx context.Context, id uint) (entity.Drone, error) {
	var pp po.Drone
	var rr ro.Drone
	if err := r.tx.WithContext(ctx).
		Where("drone_id = ?", id).First(&pp).Error; err != nil {
		r.l.Error("持久化数据获取失败", slog.Any("id", id), slog.Any("err", err))
		return entity.Drone{}, err
	}
	rr, err := r.FetchStateBySN(ctx, pp.SN)
	if err != nil {
		r.l.Error("实时数据获取失败", slog.Any("sn", pp.SN), slog.Any("err", err))
	}
	return *entity.NewDrone(&pp, &rr), err
}

func (r *DroneDefaultRepo) SelectByIDV2(ctx context.Context, id uint) (*po.Drone, error) {
	var po po.Drone
	if err := r.tx.WithContext(ctx).
		Where("drone_id = ?", id).First(&po).Error; err != nil {
		r.l.Error("无人机数据获取失败", slog.Any("id", id), slog.Any("err", err))
		return &po, err
	}
	return &po, nil
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

// UpdateDroneInfo 根据 SN 更新无人机基本信息
func (r *DroneDefaultRepo) UpdateDroneInfo(ctx context.Context, sn string, updates map[string]interface{}) error {
	if err := r.tx.Model(&po.Drone{}).Where("sn = ?", sn).Updates(updates).Error; err != nil {
		r.l.Error("更新无人机信息失败", slog.Any("sn", sn), slog.Any("updates", updates), slog.Any("err", err))
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
		FROM tb_drone_models dm 
		WHERE dm.state = 0
		ORDER BY dm.drone_model_name
	`).Scan(&models).Error; err != nil {
		r.l.Error("获取无人机型号列表失败", slog.Any("error", err))
		return nil, err
	}

	return models, nil
}

// FetchGatewaySNByDroneSN 从 Redis 获取无人机关联的网关SN
// 注意：这个实现假设网关SN直接作为字符串存储在Redis中，键为 "gateway_of_drone:{droneSN}"
// 你需要根据实际情况调整键名和数据解析逻辑
func (r *DroneDefaultRepo) FetchGatewaySNByDroneSN(ctx context.Context, droneSN string) (string, error) {
	key := "topology:" + droneSN // 示例键名，请根据实际情况修改
	gwSN, err := r.rds.Get(ctx, key).Result()
	if err == redis.Nil {
		r.l.Warn("Redis中未找到无人机的网关SN", slog.String("drone_sn", droneSN), slog.String("redis_key", key))
		return "", errors.New("gateway SN not found for drone " + droneSN)
	} else if err != nil {
		r.l.Error("从Redis获取网关SN失败", slog.String("drone_sn", droneSN), slog.String("redis_key", key), slog.Any("error", err))
		return "", err
	}
	r.l.Info("从Redis获取网关SN成功", slog.String("drone_sn", droneSN), slog.String("gw_sn", gwSN))
	return gwSN, nil
}
