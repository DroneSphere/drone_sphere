package eventhandler

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/event"
	"github.com/dronesphere/internal/service"
	"log/slog"
)

func registerDroneHandlers(eb EventBus.Bus, l *slog.Logger, drone service.DroneSvc) {
	var err error
	err = eb.Subscribe(event.UserLoginSuccessEvent, drone.OnTopoUpdate)
	if err != nil {
		l.Error(fmt.Sprintf("subscribe event %s failed: %v", event.UserLoginSuccessEvent, err))
	}
}
