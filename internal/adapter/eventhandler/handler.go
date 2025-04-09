package eventhandler

import (
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/repo"
	"github.com/dronesphere/internal/service"
)

// NewHandler 创建事件处理器
func NewHandler(eb EventBus.Bus, l *slog.Logger, mq mqtt.Client, drone service.DroneSvc, gatewaySvc service.GatewaySvc, modelRepo *repo.ModelDefaultRepo, gatewayRepo repo.GatewayRepo) {
	// 注册无人机事件处理器
	registerDroneHandlers(eb, l, mq, drone, gatewaySvc, modelRepo)

	// 注册网关事件处理器
	gatewayHandler := NewGatewayHandler(eb, mq, gatewayRepo, l)
	gatewayHandler.Subscribe(eb)
}
