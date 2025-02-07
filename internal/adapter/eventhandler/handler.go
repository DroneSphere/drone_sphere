package eventhandler

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log/slog"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
)

func NewHandler(eb EventBus.Bus, l *slog.Logger, mq mqtt.Client, drone service.DroneSvc) {
	registerDroneHandlers(eb, l, mq, drone)
}
