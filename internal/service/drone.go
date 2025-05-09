package service

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/ro"
	"github.com/dronesphere/internal/repo"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type (
	DroneSvc interface {
		Repo() DroneRepo
		SaveDroneTopo(ctx context.Context, update dto.UpdateTopoPayload) error
		FetchDeviceTopo(ctx context.Context, workspace string) ([]entity.Drone, []entity.RC, error)
		UpdateStateBySN(ctx context.Context, sn string, msg dto.DroneMessageProperty) error
		// 新增：从消息创建无人机实体
		CreateDroneFromMsg(ctx context.Context, sn string, msg dto.ProductTopo, modelRepo *repo.ModelDefaultRepo) (*entity.Drone, error)
	}

	DroneRepo interface {
		SelectAll(ctx context.Context, sn string, callsign string, modelID uint) ([]entity.Drone, error)
		Save(ctx context.Context, d entity.Drone) error
		SelectBySN(ctx context.Context, sn string) (entity.Drone, error)
		SelectByID(ctx context.Context, id uint) (entity.Drone, error)
		SelectByIDV2(ctx context.Context, id uint) (*po.Drone, error)
		FetchStateBySN(ctx context.Context, sn string) (ro.Drone, error)
		SaveState(ctx context.Context, state ro.Drone) error
		SelectAllByID(ctx context.Context, ids []uint) ([]entity.Drone, error)
		UpdateDroneInfo(ctx context.Context, sn string, updates map[string]interface{}) error
		FetchDroneModelOptions(ctx context.Context) ([]dto.DroneModelOption, error) // 获取无人机型号选项列表
	}
)

type DroneImpl struct {
	r    DroneRepo
	l    *slog.Logger
	mqtt mqtt.Client
}

func NewDroneImpl(r DroneRepo, l *slog.Logger, mqtt mqtt.Client) DroneSvc {
	return &DroneImpl{
		r:    r,
		l:    l,
		mqtt: mqtt,
	}
}

func (s *DroneImpl) Repo() DroneRepo {
	return s.r
}

func (s *DroneImpl) SaveDroneTopo(ctx context.Context, data dto.UpdateTopoPayload) error {
	rc := ctx.Value(dto.SNKey).(string)
	s.l.Info("SaveDroneTopo", slog.Any("data", data), slog.Any("rc", rc))
	// 如果没有子设备，按照遥控器SN删除无人机
	if len(data.SubDevices) == 0 {
		s.l.Info("SubDevices is empty, remove drone", slog.Any("rc", rc))
		return nil
	}

	// 保存无人机信息
	subDevice := data.SubDevices[0]
	drone := entity.Drone{
		SN:      subDevice.SN,
		Type:    subDevice.Type,
		SubType: subDevice.SubType,
	}
	s.l.Info("SaveDroneTopo", slog.Any("data", data))
	if err := s.r.Save(ctx, drone); err != nil {
		s.l.Error("SaveDroneTopo failed", slog.Any("err", err))
		return err
	}
	s.l.Info("SaveDroneTopo success", slog.Any("drone", drone))

	return nil
}

func (s *DroneImpl) FetchDeviceTopo(ctx context.Context, workspace string) ([]entity.Drone, []entity.RC, error) {
	var ds []entity.Drone
	var rcs []entity.RC
	//dds, rccs, err := s.r.FetchDeviceTopoByWorkspace(ctx, workspace)
	//if err != nil {
	//	return nil, nil, err
	//}
	//for _, d := range dds {
	//	var e entity.Drone
	//	if err := copier.Copy(&e, &d); err != nil {
	//		s.l.Error("SelectAll copier failed", slog.Any("err", err))
	//		return nil, nil, err
	//	}
	//	ds = append(ds, e)
	//}
	//for _, rc := range rccs {
	//	var e entity.RC
	//	if err := copier.Copy(&e, &rc); err != nil {
	//		s.l.Error("SelectAll copier failed", slog.Any("err", err))
	//		return nil, nil, err
	//	}
	//	rcs = append(rcs, e)
	//}
	return ds, rcs, nil
}

// UpdateStateBySN 更新无人机实时数据状态
func (s *DroneImpl) UpdateStateBySN(ctx context.Context, sn string, msg dto.DroneMessageProperty) error {
	var state = ro.Drone{
		SN:                   sn,
		Status:               ro.DroneStatusOnline,
		DroneMessageProperty: msg,
	}
	if err := s.r.SaveState(ctx, state); err != nil {
		s.l.Error("Save realtime drone failed", slog.Any("err", err))
		return err
	}
	return nil
}

// CreateDroneFromMsg 从消息创建无人机实体
// 该方法将原本位于 entity.Drone 中的 NewDroneFromMsg 功能提升到服务层
// 避免了 entity 包对 repo 包的循环引用问题
func (s *DroneImpl) CreateDroneFromMsg(ctx context.Context, sn string, msg dto.ProductTopo, modelRepo *repo.ModelDefaultRepo) (*entity.Drone, error) {
	// 创建基本的无人机实体
	d := &entity.Drone{
		SN:       sn,
		Status:   ro.DroneStatusOnline, // 新接入的无人机默认为在线状态
		Callsign: sn[:8],               // 使用SN前8位作为默认呼号
	}

	// 使用 modelRepo 查询匹配的无人机型号
	if modelRepo != nil {
		// 找到匹配的无人机型号
		droneModel, err := findDroneModelByDomainTypeSubType(ctx, modelRepo, "0", msg.Type, msg.SubType)
		if err == nil && droneModel != nil {
			// 找到匹配的无人机型号
			d.DroneModelID = droneModel.ID
			d.DroneModel = *droneModel
			s.l.Info("无人机型号匹配成功",
				slog.String("sn", sn),
				slog.String("model_name", droneModel.Name),
				slog.Uint64("model_id", uint64(droneModel.ID)))

			// 查询该型号的默认变体
			variation, err := findDefaultDroneVariation(ctx, modelRepo, droneModel.ID)
			if err == nil && variation != nil {
				d.VariationID = variation.ID
				d.Variation = *variation
				s.l.Info("无人机默认变体匹配成功",
					slog.String("sn", sn),
					slog.String("model_name", droneModel.Name),
					slog.String("variation_name", variation.Name))
			} else {
				s.l.Warn("无人机默认变体匹配失败",
					slog.String("sn", sn),
					slog.Uint64("model_id", uint64(droneModel.ID)),
					slog.Any("error", err))
			}
		} else {
			// 未找到匹配的无人机型号
			s.l.Warn("无人机型号匹配失败",
				slog.String("sn", sn),
				slog.String("domain", msg.Domain),
				slog.Int64("type", int64(msg.Type)),
				slog.Int64("sub_type", int64(msg.SubType)),
				slog.Any("error", err))
		}
	}

	return d, nil
}

// findDroneModelByDomainTypeSubType 是内部辅助方法，根据 Domain、Type、SubType 查找无人机型号
// 将此功能从 modelRepo 中提取到服务层，作为内部实现
func findDroneModelByDomainTypeSubType(ctx context.Context, modelRepo *repo.ModelDefaultRepo, domain string, deviceType int, subType int) (*po.DroneModel, error) {
	// 将 domain 字符串转换为整数
	// var domainInt int
	// var err error

	// // 兼容不同类型的 domain 格式
	// if domain != "" {
	// 	domainInt, err = strconv.Atoi(domain)
	// 	if err != nil {
	// 		// 默认为 0，表示未知领域
	// 		domainInt = 0
	// 	}
	// }

	// 在数据库中查询匹配的无人机型号
	return modelRepo.FindDroneModelByDomainTypeSubType(ctx, domain, deviceType, subType)
}

// findDefaultDroneVariation 是内部辅助方法，查询指定型号的默认变体
func findDefaultDroneVariation(ctx context.Context, modelRepo *repo.ModelDefaultRepo, droneModelID uint) (*po.DroneVariation, error) {
	return modelRepo.FindDefaultDroneVariation(ctx, droneModelID)
}
