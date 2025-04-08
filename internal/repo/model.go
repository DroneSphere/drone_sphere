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
	return &ModelDefaultRepo{
		tx: tx,
		l:  l,
	}
}

func (r *ModelDefaultRepo) SelectAllDroneVariation(ctx context.Context, query map[string]string) ([]po.DroneVariation, error) {
	var variations []po.DroneVariation
	if err := r.tx.
		WithContext(ctx).
		Preload("DroneModel").
		Preload("Gimbals").
		Preload("Payloads").
		Where(query).
		Find(&variations).Error; err != nil {
		r.l.Error("select all drone variations", "error", err)
		return nil, err
	}
	return variations, nil
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
		if err := r.tx.WithContext(ctx).Where("gateway_model_id = ?", drone.GatewayID).First(&gateway).Error; err != nil {
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

// 创建无人机型号
func (r *ModelDefaultRepo) CreateDroneModel(ctx context.Context, model *po.DroneModel) error {
	if err := r.tx.WithContext(ctx).Create(model).Error; err != nil {
		r.l.Error("创建无人机型号失败", "error", err)
		return err
	}
	return nil
}

// 创建云台型号
func (r *ModelDefaultRepo) CreateGimbalModel(ctx context.Context, model *po.GimbalModel) error {
	if err := r.tx.WithContext(ctx).Create(model).Error; err != nil {
		r.l.Error("创建云台型号失败", "error", err)
		return err
	}
	return nil
}

// 创建网关型号
func (r *ModelDefaultRepo) CreateGatewayModel(ctx context.Context, model *po.GatewayModel) error {
	if err := r.tx.WithContext(ctx).Create(model).Error; err != nil {
		r.l.Error("创建网关型号失败", "error", err)
		return err
	}
	return nil
}

// 创建负载型号
func (r *ModelDefaultRepo) CreatePayloadModel(ctx context.Context, model *po.PayloadModel) error {
	if err := r.tx.WithContext(ctx).Create(model).Error; err != nil {
		r.l.Error("创建负载型号失败", "error", err)
		return err
	}
	return nil
}

// 生成无人机变体
func (r *ModelDefaultRepo) GenerateDroneVariations(ctx context.Context, droneModelID uint) ([]po.DroneVariation, error) {
	// 使用po包中已定义的生成变体方法
	variations, err := po.GenerateDroneVariations(r.tx.WithContext(ctx), droneModelID)
	if err != nil {
		r.l.Error("生成无人机变体失败", "drone_model_id", droneModelID, "error", err)
		return nil, err
	}
	return variations, nil
}

// FindDroneModelByDomainTypeSubType 根据 Domain、Type、SubType 查找无人机型号
// 参数说明：
// domain: 设备领域，字符串类型，需要转换为整数
// deviceType: 设备类型
// subType: 设备子类型
func (r *ModelDefaultRepo) FindDroneModelByDomainTypeSubType(ctx context.Context, domain string, deviceType int, subType int) (*po.DroneModel, error) {
	r.l.Debug("查询无人机型号",
		"type", deviceType,
		"sub_type", subType)
	r.l.Warn("domain 参数已弃，使用 deviceType 和 subType 进行查询")

	// 在数据库中查询匹配的无人机型号
	var droneModel po.DroneModel
	if err := r.tx.WithContext(ctx).
		Where("drone_model_type = ? AND drone_model_sub_type = ?", deviceType, subType).
		First(&droneModel).Error; err != nil {
		r.l.Error("根据标识查找无人机型号失败",
			"type", deviceType,
			"sub_type", subType,
			"error", err)
		return nil, err
	}

	// 加载关联的云台和负载信息
	if err := r.tx.WithContext(ctx).
		Preload("Gimbals").
		Preload("Payloads").
		First(&droneModel, droneModel.ID).Error; err != nil {
		r.l.Warn("加载无人机型号的关联数据失败", "model_id", droneModel.ID, "error", err)
		// 继续使用不完整的型号信息
	}

	return &droneModel, nil
}

// FindDefaultDroneVariation 根据 DroneModel 查找默认变体
func (r *ModelDefaultRepo) FindDefaultDroneVariation(ctx context.Context, droneModelID uint) (*po.DroneVariation, error) {
	// 查询指定型号的默认变体配置
	var variation po.DroneVariation
	if err := r.tx.WithContext(ctx).
		Where("drone_model_id = ?", droneModelID).
		Preload("DroneModel").
		Preload("Gimbals").
		Preload("Payloads").
		First(&variation).Error; err != nil {

		// 如果没有默认变体，则尝试获取该型号的任意变体
		if err := r.tx.WithContext(ctx).
			Where("drone_model_id = ?", droneModelID).
			Preload("DroneModel").
			Preload("Gimbals").
			Preload("Payloads").
			First(&variation).Error; err != nil {
			r.l.Error("查找无人机默认变体失败", "model_id", droneModelID, "error", err)
			return nil, err
		}
	}

	return &variation, nil
}
