package repo

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"gorm.io/gorm"
)

type ModelDefaultRepo struct {
	tx *gorm.DB
	l  *slog.Logger
}

func NewModelDefaultRepo(tx *gorm.DB, l *slog.Logger) *ModelDefaultRepo {
	_ = tx.AutoMigrate(&po.GatewayModel{}, &po.DroneModel{}, &po.GimbalModel{}, &po.PayloadModel{})
	return &ModelDefaultRepo{
		tx: tx,
		l:  l,
	}
}

func (r *ModelDefaultRepo) SelectAllDroneModel(ctx context.Context) ([]entity.DroneModel, error) {
	var drones []po.DroneModel
	if err := r.tx.WithContext(ctx).Preload("Gimbals").Find(&drones).Error; err != nil {
		r.l.Error("select all drone models", "error", err)
		return nil, err
	}
	var res []entity.DroneModel
	for _, drone := range drones {
		var gateway po.GatewayModel
		if err := r.tx.WithContext(ctx).Where("id = ?", drone.GatewayID).First(&gateway).Error; err != nil {
			r.l.Error("select drone model's gateway", "error", err)
			return nil, err
		}
		res = append(res, *entity.NewDroneModelFromPO(drone, gateway))
	}
	return res, nil
}
func (r *ModelDefaultRepo) SelectAllGimbals(ctx context.Context) ([]po.GimbalModel, error) {
	var gimbals []po.GimbalModel
	if err := r.tx.WithContext(ctx).Find(&gimbals).Error; err != nil {
		r.l.Error("select all gimbals", "error", err)
		return nil, err
	}
	return gimbals, nil
}
func (r *ModelDefaultRepo) SelectGimbalsByIDs(ctx context.Context, ids []uint) ([]po.GimbalModel, error) {
	var gimbals []po.GimbalModel
	if err := r.tx.WithContext(ctx).Where("id IN ?", ids).Find(&gimbals).Error; err != nil {
		r.l.Error("select gimbals by ids", "error", err)
		return nil, err
	}
	return gimbals, nil
}

func (r *ModelDefaultRepo) SelectAllGatewayModel(ctx context.Context) ([]po.GatewayModel, error) {
	var gateways []po.GatewayModel
	if err := r.tx.WithContext(ctx).Find(&gateways).Error; err != nil {
		r.l.Error("select all gateway models", "error", err)
		return nil, err
	}
	return gateways, nil
}
func (r *ModelDefaultRepo) SelectAllPayloadModel(ctx context.Context) ([]po.PayloadModel, error) {
	var payloads []po.PayloadModel
	if err := r.tx.WithContext(ctx).Find(&payloads).Error; err != nil {
		r.l.Error("select all payload models", "error", err)
		return nil, err
	}
	return payloads, nil
}
