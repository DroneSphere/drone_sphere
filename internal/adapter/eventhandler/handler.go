package eventhandler

import (
	"log/slog"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
)

func NewHandler(eb EventBus.Bus, l *slog.Logger, drone service.DroneSvc) {
	registerDroneHandlers(eb, l, drone)
}
