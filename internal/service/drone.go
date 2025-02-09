package service

import (
	"context"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jinzhu/copier"
	"log/slog"
)

type (
	DroneSvc interface {
		SaveDroneTopo(ctx context.Context, update dto.UpdateTopoPayload) error
		ListAll(ctx context.Context) ([]api.DroneItemResult, error)
		UpdateOnline(ctx context.Context, sn string) error
		UpdateOffline(ctx context.Context, sn string) error
	}

	DroneRepo interface {
		ListAll(ctx context.Context) ([]entity.Drone, error)
		RemoveDroneBySN(ctx context.Context, rc string) error
		Save(ctx context.Context, d *entity.Drone, rc string) error
		SaveRealtimeStatus(ctx context.Context, sn string, isOnline bool) error
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

func (s *DroneImpl) SaveDroneTopo(ctx context.Context, data dto.UpdateTopoPayload) error {
	rc := ctx.Value(dto.SNKey).(string)
	// 如果没有子设备，按照遥控器SN删除无人机
	if len(data.SubDevices) == 0 {
		return s.r.RemoveDroneBySN(ctx, rc)
	}

	// 保存无人机信息
	subDevice := data.SubDevices[0]
	drone := entity.Drone{
		SN:      subDevice.SN,
		Type:    subDevice.Type,
		SubType: subDevice.SubType,
	}
	s.l.Info("SaveDroneTopo", slog.Any("data", data))
	if err := s.r.Save(ctx, &drone, rc); err != nil {
		s.l.Error("SaveDroneTopo failed", slog.Any("err", err))
		return err
	}

	return nil
}

func (s *DroneImpl) ListAll(ctx context.Context) ([]api.DroneItemResult, error) {
	ds, err := s.r.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	var res []api.DroneItemResult
	for _, d := range ds {
		// 拷贝数据
		e := api.DroneItemResult{}
		err := copier.Copy(&e, &d)
		if err != nil {
			s.l.Error("ListAll copier failed", slog.Any("err", err))
			return nil, err
		}
		// 计算额外字段
		e.Status = d.StatusText()
		e.ProductType = d.ProductType()
		e.IsRTKAvailable = d.IsRTKAvailable()
		e.IsThermalAvailable = d.IsThermalAvailable()
		res = append(res, e)
	}
	return res, nil
}

func (s *DroneImpl) UpdateOnline(ctx context.Context, sn string) error {
	return s.r.SaveRealtimeStatus(ctx, sn, true)
}

func (s *DroneImpl) UpdateOffline(ctx context.Context, sn string) error {
	return s.r.SaveRealtimeStatus(ctx, sn, false)
}
