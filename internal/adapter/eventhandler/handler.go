package eventhandler

import (
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/repo"
	"github.com/dronesphere/internal/service"
)

// NewHandler 创建事件处理器
// 添加 modelRepo 参数，用于提供无人机型号相关的查询功能
func NewHandler(eb EventBus.Bus, l *slog.Logger, mq mqtt.Client, drone service.DroneSvc, modelRepo *repo.ModelDefaultRepo) {
	registerDroneHandlers(eb, l, mq, drone, modelRepo)
}
