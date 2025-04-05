package repo

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"gorm.io/gorm"
)

type (
	// GatewaySvc 网关设备仓储接口定义
	GatewayRepo interface {
		// 基础的CRUD操作
		Create(ctx context.Context, gateway *po.Gateway) error
		UpdateCallsign(ctx context.Context, sn, callsign string) error
		UpdateStatus(ctx context.Context, sn string, status int) error
		UpdateOnlineStatus(ctx context.Context, sn string, isOnline bool) error
		DeleteBySN(ctx context.Context, sn string) error

		// 查询相关方法
		SelectAll(ctx context.Context) ([]*entity.Gateway, error)
		SelectBySN(ctx context.Context, sn string) (*entity.Gateway, error)
		SelectByUserID(ctx context.Context, userID uint) ([]*entity.Gateway, error)

		// 关联关系管理方法
		AddDroneRelation(ctx context.Context, gatewaySN, droneSN string) error
		RemoveDroneRelation(ctx context.Context, gatewaySN, droneSN string) error
		GetConnectedDrones(ctx context.Context, gatewaySN string) ([]po.Drone, error)
	}

	// GatewayDefaultRepo 网关设备仓储默认实现
	GatewayDefaultRepo struct {
		tx *gorm.DB
		l  *slog.Logger
	}
)

// NewGatewayRepo 创建网关设备仓储实例
func NewGatewayRepo(tx *gorm.DB, l *slog.Logger) GatewayRepo {
	return &GatewayDefaultRepo{
		tx: tx,
		l:  l,
	}
}

func (r *GatewayDefaultRepo) Create(ctx context.Context, gateway *po.Gateway) error {
	return r.tx.WithContext(ctx).Create(gateway).Error
}

func (r *GatewayDefaultRepo) UpdateCallsign(ctx context.Context, sn, callsign string) error {
	return r.tx.WithContext(ctx).Model(&po.Gateway{}).Where("sn = ?", sn).Update("callsign", callsign).Error
}

func (r *GatewayDefaultRepo) UpdateStatus(ctx context.Context, sn string, status int) error {
	return r.tx.WithContext(ctx).Model(&po.Gateway{}).Where("sn = ?", sn).Update("status", status).Error
}

func (r *GatewayDefaultRepo) UpdateOnlineStatus(ctx context.Context, sn string, isOnline bool) error {
	status := 0
	if isOnline {
		status = 1
	}
	return r.tx.WithContext(ctx).Model(&po.Gateway{}).Where("sn = ?", sn).
		Updates(map[string]interface{}{
			"status":         status,
			"last_online_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

func (r *GatewayDefaultRepo) DeleteBySN(ctx context.Context, sn string) error {
	return r.tx.WithContext(ctx).Where("sn = ?", sn).Delete(&po.Gateway{}).Error
}

func (r *GatewayDefaultRepo) SelectAll(ctx context.Context) ([]*entity.Gateway, error) {
	var gateways []po.Gateway
	if err := r.tx.WithContext(ctx).
		Preload("GatewayModel").
		Find(&gateways).Error; err != nil {
		return nil, err
	}

	var result []*entity.Gateway
	for i := range gateways {
		if e := entity.NewGatewayFromPO(&gateways[i]); e != nil {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *GatewayDefaultRepo) SelectBySN(ctx context.Context, sn string) (*entity.Gateway, error) {
	var gateway po.Gateway
	if err := r.tx.WithContext(ctx).
		Preload("GatewayModel").
		Where("sn = ?", sn).
		First(&gateway).Error; err != nil {
		return nil, err
	}
	return entity.NewGatewayFromPO(&gateway), nil
}

func (r *GatewayDefaultRepo) SelectByUserID(ctx context.Context, userID uint) ([]*entity.Gateway, error) {
	var gateways []po.Gateway
	if err := r.tx.WithContext(ctx).
		Preload("GatewayModel").
		Where("user_id = ?", userID).
		Find(&gateways).Error; err != nil {
		return nil, err
	}

	var result []*entity.Gateway
	for i := range gateways {
		if e := entity.NewGatewayFromPO(&gateways[i]); e != nil {
			result = append(result, e)
		}
	}
	return result, nil
}

// AddDroneRelation 添加网关与无人机的关联关系
func (r *GatewayDefaultRepo) AddDroneRelation(ctx context.Context, gatewaySN, droneSN string) error {
	var gateway po.Gateway
	var drone po.Drone

	// 查询网关和无人机
	err := r.tx.WithContext(ctx).Where("sn = ?", gatewaySN).First(&gateway).Error
	if err != nil {
		return err
	}

	err = r.tx.WithContext(ctx).Where("sn = ?", droneSN).First(&drone).Error
	if err != nil {
		return err
	}

	// 创建关联记录
	relation := &po.GatewayDroneRelation{
		GatewayID: gateway.ID,
		GatewaySN: gateway.SN,
		DroneID:   drone.ID,
		DroneSN:   drone.SN,
	}

	return r.tx.WithContext(ctx).Create(relation).Error
}

// RemoveDroneRelation 移除网关与无人机的关联关系
func (r *GatewayDefaultRepo) RemoveDroneRelation(ctx context.Context, gatewaySN, droneSN string) error {
	return r.tx.WithContext(ctx).
		Where("gateway_sn = ? AND drone_sn = ?", gatewaySN, droneSN).
		Delete(&po.GatewayDroneRelation{}).Error
}

// GetConnectedDrones 获取连接到指定网关的所有无人机
func (r *GatewayDefaultRepo) GetConnectedDrones(ctx context.Context, gatewaySN string) ([]po.Drone, error) {
	var drones []po.Drone
	err := r.tx.WithContext(ctx).
		Preload("DroneModel"). // 预加载无人机型号信息
		Joins("JOIN tb_gateway_drone_relations ON tb_gateway_drone_relations.drone_id = tb_drones.id").
		Where("tb_gateway_drone_relations.gateway_sn = ? AND tb_gateway_drone_relations.state = 0", gatewaySN).
		Find(&drones).Error
	if err != nil {
		r.l.Error("获取网关关联的无人机失败", "error", err, "gateway_sn", gatewaySN)
		return nil, err
	}
	return drones, nil
}
