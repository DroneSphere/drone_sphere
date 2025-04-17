package repo

import (
	"context"
	"log/slog"
	"strings"

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

// SelectGatewayModels 根据条件查询网关型号
func (r *ModelDefaultRepo) SelectGatewayModels(ctx context.Context, query map[string]interface{}) ([]po.GatewayModel, error) {
	var gateways []po.GatewayModel
	db := r.tx.WithContext(ctx)

	// 处理查询条件
	if query != nil {
		for k, v := range query {
			if strings.Contains(k, "?") {
				db = db.Where(k, v) // 对于包含占位符的条件
			} else {
				db = db.Where(k+" = ?", v) // 对于不包含占位符的条件
			}
		}
	}

	if err := db.Find(&gateways).Error; err != nil {
		r.l.Error("查询网关型号列表失败", "error", err)
		return nil, err
	}
	return gateways, nil
}

// SelectGatewayModelByID 根据ID查询单个网关型号
func (r *ModelDefaultRepo) SelectGatewayModelByID(ctx context.Context, id uint) (*po.GatewayModel, error) {
	var gateway po.GatewayModel
	if err := r.tx.WithContext(ctx).Where("gateway_model_id = ?", id).First(&gateway).Error; err != nil {
		r.l.Error("根据ID查询网关型号失败", "id", id, "error", err)
		return nil, err
	}
	return &gateway, nil
}

// SelectDroneModels 根据条件查询无人机型号
func (r *ModelDefaultRepo) SelectDroneModels(ctx context.Context, query map[string]interface{}) ([]entity.DroneModel, error) {
	var drones []po.DroneModel
	db := r.tx.WithContext(ctx).Preload("Gimbals")

	// 处理查询条件
	if query != nil {
		for k, v := range query {
			if strings.Contains(k, "?") {
				db = db.Where(k, v) // 对于包含占位符的条件
			} else {
				db = db.Where(k+" = ?", v) // 对于不包含占位符的条件
			}
		}
	}

	if err := db.Find(&drones).Error; err != nil {
		r.l.Error("查询无人机型号列表失败", "error", err)
		return nil, err
	}

	// 转换为实体模型
	var res []entity.DroneModel
	for _, drone := range drones {
		var gateway po.GatewayModel
		if err := r.tx.WithContext(ctx).Where("gateway_model_id = ?", drone.GatewayID).First(&gateway).Error; err != nil {
			r.l.Error("查询无人机型号的网关失败", "error", err)
			return nil, err
		}
		res = append(res, *entity.NewDroneModelFromPO(drone, gateway))
	}
	return res, nil
}

// SelectDroneModelByID 根据ID查询单个无人机型号
func (r *ModelDefaultRepo) SelectDroneModelByID(ctx context.Context, id uint) (*entity.DroneModel, error) {
	var drone po.DroneModel
	if err := r.tx.WithContext(ctx).Preload("Gimbals").Where("drone_model_id = ?", id).First(&drone).Error; err != nil {
		r.l.Error("根据ID查询无人机型号失败", "id", id, "error", err)
		return nil, err
	}

	// 加载关联的网关信息
	var gateway po.GatewayModel
	if err := r.tx.WithContext(ctx).Where("gateway_model_id = ?", drone.GatewayID).First(&gateway).Error; err != nil {
		r.l.Error("查询无人机型号的网关失败", "gateway_id", drone.GatewayID, "error", err)
		return nil, err
	}

	return entity.NewDroneModelFromPO(drone, gateway), nil
}

// SelectGimbalModels 根据条件查询云台型号
func (r *ModelDefaultRepo) SelectGimbalModels(ctx context.Context, query map[string]interface{}) ([]po.GimbalModel, error) {
	var gimbals []po.GimbalModel
	db := r.tx.WithContext(ctx)

	// 处理查询条件
	if query != nil {
		for k, v := range query {
			if strings.Contains(k, "?") {
				db = db.Where(k, v) // 对于包含占位符的条件
			} else {
				db = db.Where(k+" = ?", v) // 对于不包含占位符的条件
			}
		}
	}

	if err := db.Find(&gimbals).Error; err != nil {
		r.l.Error("查询云台型号列表失败", "error", err)
		return nil, err
	}
	return gimbals, nil
}

// SelectGimbalModelByID 根据ID查询单个云台型号
func (r *ModelDefaultRepo) SelectGimbalModelByID(ctx context.Context, id uint) (*po.GimbalModel, error) {
	var gimbal po.GimbalModel
	if err := r.tx.WithContext(ctx).Where("gimbal_model_id = ?", id).First(&gimbal).Error; err != nil {
		r.l.Error("根据ID查询云台型号失败", "id", id, "error", err)
		return nil, err
	}
	return &gimbal, nil
}

// UpdateDroneModel 更新无人机型号
func (r *ModelDefaultRepo) UpdateDroneModel(ctx context.Context, id uint, model *po.DroneModel) error {
	// 先检查记录是否存在
	var existing po.DroneModel
	if err := r.tx.WithContext(ctx).Where("drone_model_id = ?", id).First(&existing).Error; err != nil {
		r.l.Error("更新无人机型号失败，记录不存在", "id", id, "error", err)
		return err
	}

	// 更新记录
	if err := r.tx.WithContext(ctx).Model(&po.DroneModel{}).Where("drone_model_id = ?", id).Updates(map[string]interface{}{
		"drone_model_name":        model.Name,
		"drone_model_description": model.Description,
		"drone_model_domain":      model.Domain,
		"drone_model_type":        model.Type,
		"drone_model_sub_type":    model.SubType,
		"gateway_model_id":        model.GatewayID,
		"is_rtk_available":        model.IsRTKAvailable,
	}).Error; err != nil {
		r.l.Error("更新无人机型号失败", "id", id, "error", err)
		return err
	}

	// 如果有云台关联信息，需要更新关联
	if len(model.Gimbals) > 0 {
		// 清除原有关联
		if err := r.tx.WithContext(ctx).Model(&existing).Association("Gimbals").Clear(); err != nil {
			r.l.Error("清除无人机型号云台关联失败", "id", id, "error", err)
			return err
		}

		// 添加新关联
		if err := r.tx.WithContext(ctx).Model(&existing).Association("Gimbals").Replace(&model.Gimbals); err != nil {
			r.l.Error("更新无人机型号云台关联失败", "id", id, "error", err)
			return err
		}
	}

	// 如果有负载关联信息，需要更新关联
	if len(model.Payloads) > 0 {
		// 清除原有关联
		if err := r.tx.WithContext(ctx).Model(&existing).Association("Payloads").Clear(); err != nil {
			r.l.Error("清除无人机型号负载关联失败", "id", id, "error", err)
			return err
		}

		// 添加新关联
		if err := r.tx.WithContext(ctx).Model(&existing).Association("Payloads").Replace(&model.Payloads); err != nil {
			r.l.Error("更新无人机型号负载关联失败", "id", id, "error", err)
			return err
		}
	}

	return nil
}

// UpdateGimbalModel 更新云台型号
func (r *ModelDefaultRepo) UpdateGimbalModel(ctx context.Context, id uint, model *po.GimbalModel) error {
	// 先检查记录是否存在
	var existing po.GimbalModel
	if err := r.tx.WithContext(ctx).Where("gimbal_model_id = ?", id).First(&existing).Error; err != nil {
		r.l.Error("更新云台型号失败，记录不存在", "id", id, "error", err)
		return err
	}

	// 更新记录
	if err := r.tx.WithContext(ctx).Model(&po.GimbalModel{}).Where("gimbal_model_id = ?", id).Updates(map[string]interface{}{
		"gimbal_model_name":        model.Name,
		"gimbal_model_description": model.Description,
		"gimbal_model_product":     model.Product,
		"gimbal_model_domain":      model.Domain,
		"gimbal_model_type":        model.Type,
		"gimbal_model_sub_type":    model.SubType,
		"gimbalindex":              model.Gimbalindex,
		"is_thermal_available":     model.IsThermalAvailable,
	}).Error; err != nil {
		r.l.Error("更新云台型号失败", "id", id, "error", err)
		return err
	}

	return nil
}

// UpdateGatewayModel 更新网关型号
func (r *ModelDefaultRepo) UpdateGatewayModel(ctx context.Context, id uint, model *po.GatewayModel) error {
	// 先检查记录是否存在
	var existing po.GatewayModel
	if err := r.tx.WithContext(ctx).Where("gateway_model_id = ?", id).First(&existing).Error; err != nil {
		r.l.Error("更新网关型号失败，记录不存在", "id", id, "error", err)
		return err
	}

	// 更新记录
	if err := r.tx.WithContext(ctx).Model(&po.GatewayModel{}).Where("gateway_model_id = ?", id).Updates(map[string]interface{}{
		"gateway_model_name":        model.Name,
		"gateway_model_description": model.Description,
		"gateway_model_domain":      model.Domain,
		"gateway_model_type":        model.Type,
		"gateway_model_sub_type":    model.SubType,
	}).Error; err != nil {
		r.l.Error("更新网关型号失败", "id", id, "error", err)
		return err
	}

	return nil
}
