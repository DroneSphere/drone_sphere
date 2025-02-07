package service

import (
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/entity"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
)

type (
	DroneSvc interface {
		SaveDroneTopo(update dto.UpdateTopoPayload) error
		ListAll() ([]entity.Drone, error)
	}

	DroneRepo interface {
		ListAll() ([]entity.Drone, error)
		Save(d *entity.Drone, rc string) error
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

func (s *DroneImpl) SaveDroneTopo(data dto.UpdateTopoPayload) error {
	if len(data.SubDevices) == 0 {
		return nil
	}
	subDevice := data.SubDevices[0]
	drone := entity.Drone{
		SN:      subDevice.SN,
		Type:    subDevice.Type,
		SubType: subDevice.SubType,
	}
	s.l.Info("SaveDroneTopo", slog.Any("data", data))
	if err := s.r.Save(&drone, ""); err != nil {
		s.l.Error("SaveDroneTopo failed", slog.Any("err", err))
		return err
	}
	return nil
}

func (s *DroneImpl) ListAll() ([]entity.Drone, error) {
	return s.r.ListAll()
}
