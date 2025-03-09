package service

import (
	"context"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/ro"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
)

type (
	DroneSvc interface {
		Repo() DroneRepo
		SaveDroneTopo(ctx context.Context, update dto.UpdateTopoPayload) error
		FetchDeviceTopo(ctx context.Context, workspace string) ([]entity.Drone, []entity.RC, error)
		UpdateStateBySN(ctx context.Context, sn string, msg dto.DroneMessageProperty) error
	}

	DroneRepo interface {
		SelectAll(ctx context.Context) ([]entity.Drone, error)
		Save(ctx context.Context, d entity.Drone) error
		SelectBySN(ctx context.Context, sn string) (entity.Drone, error)
		FetchStateBySN(ctx context.Context, sn string) (ro.Drone, error)
		SaveState(ctx context.Context, state ro.Drone) error
		SelectAllByID(ctx context.Context, ids []uint) ([]entity.Drone, error)
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
