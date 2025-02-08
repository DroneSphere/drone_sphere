package service

import (
	"context"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
)

type (
	DroneSvc interface {
		SaveDroneTopo(ctx context.Context, update dto.UpdateTopoPayload) error
		ListAll() ([]entity.Drone, error)
	}

	DroneRepo interface {
		ListAll() ([]entity.Drone, error)
		RemoveDroneBySN(ctx context.Context, rc string) error
		Save(ctx context.Context, d *entity.Drone, rc string) error
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

func (s *DroneImpl) ListAll() ([]entity.Drone, error) {
	return s.r.ListAll()
}
