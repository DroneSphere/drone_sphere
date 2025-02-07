package service

import (
	"context"
	"github.com/dronesphere/internal/model/entity"
	m "github.com/dronesphere/tools/mock_tool"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"os"
	"testing"
)

type MockDroneRepo struct {
	mock.Mock
}

func (m *MockDroneRepo) ListAll() ([]entity.Drone, error) {
	args := m.Called()
	return args.Get(0).([]entity.Drone), args.Error(1)
}

func setup() (DroneSvc, mqtt.Client) {
	r := new(MockDroneRepo)
	mq := m.NewMockMQTTClient()
	svc := NewDroneImpl(r, slog.New(slog.NewTextHandler(os.Stdout, nil)), mq)
	return svc, mq
}

func TestDroneImpl_OnTopoUpdate(t *testing.T) {
	svc, mq := setup()
	sn := "123"
	msg := "hello"
	ctx := context.WithValue(context.Background(), "sn", sn)
	err := svc.OnTopoUpdate(ctx)
	if err != nil {
		t.Errorf("OnTopoUpdate error: %v", err)
	}
	mq.Publish("topo/"+sn, 0, false, []byte(msg))
	mq.Subscribe("topo/"+sn, 0, func(c mqtt.Client, m mqtt.Message) {
		if m.Topic() != "topo/"+sn {
			t.Errorf("topic error: %v", m.Topic())
		}
		if string(m.Payload()) != msg {
			t.Errorf("payload error: %v", string(m.Payload()))
		}
	})
}
