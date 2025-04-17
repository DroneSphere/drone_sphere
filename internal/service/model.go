package service

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
)

type (
	// ModelSvc 型号服务接口
	// 根据要求，如果Service接口的函数只调用Repo的函数，可以不在Service中再次定义函数
	ModelSvc interface {
		Repo() ModelRepo
		// 批量创建型号
		BatchCreateModels(ctx context.Context, droneModels []po.DroneModel, gimbalModels []po.GimbalModel,
			gatewayModels []po.GatewayModel, payloadModels []po.PayloadModel) error
	}

	// ModelRepo 型号仓库接口
	ModelRepo interface {
		// 查询相关方法
		SelectAllDroneVariation(ctx context.Context, query map[string]string) ([]po.DroneVariation, error)
		SelectAllDroneModel(ctx context.Context, name string) ([]entity.DroneModel, error)
		SelectDroneModels(ctx context.Context, query map[string]interface{}) ([]entity.DroneModel, error)
		SelectDroneModelByID(ctx context.Context, id uint) (*entity.DroneModel, error)

		SelectAllGimbals(ctx context.Context) ([]po.GimbalModel, error)
		SelectGimbalModels(ctx context.Context, query map[string]interface{}) ([]po.GimbalModel, error)
		SelectGimbalModelByID(ctx context.Context, id uint) (*po.GimbalModel, error)
		SelectGimbalsByIDs(ctx context.Context, ids []uint) ([]po.GimbalModel, error)

		SelectAllGatewayModel(ctx context.Context, name string) ([]po.GatewayModel, error)
		SelectGatewayModels(ctx context.Context, query map[string]interface{}) ([]po.GatewayModel, error)
		SelectGatewayModelByID(ctx context.Context, id uint) (*po.GatewayModel, error)

		SelectAllPayloadModel(ctx context.Context) ([]po.PayloadModel, error)

		// 创建型号方法
		CreateDroneModel(ctx context.Context, model *po.DroneModel) error
		CreateGimbalModel(ctx context.Context, model *po.GimbalModel) error
		CreateGatewayModel(ctx context.Context, model *po.GatewayModel) error
		CreatePayloadModel(ctx context.Context, model *po.PayloadModel) error

		// 更新型号方法
		UpdateDroneModel(ctx context.Context, id uint, model *po.DroneModel) error
		UpdateGimbalModel(ctx context.Context, id uint, model *po.GimbalModel) error
		UpdateGatewayModel(ctx context.Context, id uint, model *po.GatewayModel) error

		// 通用更新型号方法（可用于软删除）
		UpdateDroneModelFields(ctx context.Context, id uint, updates map[string]interface{}) error
		UpdateGimbalModelFields(ctx context.Context, id uint, updates map[string]interface{}) error
		UpdateGatewayModelFields(ctx context.Context, id uint, updates map[string]interface{}) error

		// 生成无人机变体
		GenerateDroneVariations(ctx context.Context, droneModelID uint) ([]po.DroneVariation, error)
	}
)

// ModelImpl 型号服务实现
type ModelImpl struct {
	repo ModelRepo
	l    *slog.Logger
}

// NewModelImpl 创建型号服务实例
func NewModelImpl(repo ModelRepo, l *slog.Logger) ModelSvc {
	return &ModelImpl{
		repo: repo,
		l:    l,
	}
}

// Repo 获取仓库实例
func (m *ModelImpl) Repo() ModelRepo {
	return m.repo
}

// BatchCreateModels 批量创建型号方法
// 这个方法包含业务逻辑，不仅仅是调用Repo，所以保留在Service中
func (m *ModelImpl) BatchCreateModels(ctx context.Context, droneModels []po.DroneModel, gimbalModels []po.GimbalModel,
	gatewayModels []po.GatewayModel, payloadModels []po.PayloadModel) error {

	// 先创建网关型号，因为无人机型号依赖于网关型号
	for i := range gatewayModels {
		if err := m.repo.CreateGatewayModel(ctx, &gatewayModels[i]); err != nil {
			m.l.Error("批量创建网关型号失败", "error", err)
			return err
		}
	}

	// 创建无人机型号
	for i := range droneModels {
		if err := m.repo.CreateDroneModel(ctx, &droneModels[i]); err != nil {
			m.l.Error("批量创建无人机型号失败", "error", err)
			return err
		}

		// 创建完无人机型号后，自动生成变体
		if _, err := m.repo.GenerateDroneVariations(ctx, droneModels[i].ID); err != nil {
			m.l.Error("生成无人机变体失败", "drone_model_id", droneModels[i].ID, "error", err)
			return err
		}
	}

	// 创建云台型号
	for i := range gimbalModels {
		if err := m.repo.CreateGimbalModel(ctx, &gimbalModels[i]); err != nil {
			m.l.Error("批量创建云台型号失败", "error", err)
			return err
		}
	}

	// 创建负载型号
	for i := range payloadModels {
		if err := m.repo.CreatePayloadModel(ctx, &payloadModels[i]); err != nil {
			m.l.Error("批量创建负载型号失败", "error", err)
			return err
		}
	}

	return nil
}
