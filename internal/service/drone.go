package service

import (
	"context"
	"github.com/dronesphere/internal/model/entity"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
)

type (
	DroneSvc interface {
		OnTopoUpdate(ctx context.Context) error
		ListAll() ([]entity.Drone, error)
	}

	DroneRepo interface {
		ListAll() ([]entity.Drone, error)
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

func (s *DroneImpl) OnTopoUpdate(ctx context.Context) error {
	s.l.Info("OnTopoUpdate")
	sn := ctx.Value("sn").(string)
	s.l.Info("Get sn", slog.Any("sn", sn))
	topic := "topo/" + sn
	s.mqtt.Subscribe(topic, 0, func(c mqtt.Client, m mqtt.Message) {
		s.l.Info("Received message", slog.Any("topic", m.Topic()), slog.Any("message", string(m.Payload())))
		s.mqtt.Publish(topic, 0, false, m.Payload())
	})
	return nil
}

func (s *DroneImpl) ListAll() ([]entity.Drone, error) {
	return s.r.ListAll()
}
