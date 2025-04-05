package ws

import (
	"log/slog"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/asaskevich/EventBus"
	"github.com/bytedance/sonic"
	"github.com/dronesphere/internal/event"
	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/robfig/cron/v3"
	slogfiber "github.com/samber/slog-fiber"
)

func NewRouter(app *fiber.App, eb EventBus.Bus, l *slog.Logger, user service.UserSvc, drone service.DroneSvc) {
	sfCfg := slogfiber.Config{
		WithTraceID: true,
	}
	app.Use(slogfiber.NewWithConfig(l, sfCfg))
	app.Use(cors.New())

	// Prometheus metrics
	prometheus := fiberprometheus.New("dronesphere")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// K8s probe
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	cr := cron.New(cron.WithSeconds())
	app.Get("/", websocket.New(func(c *websocket.Conn) {
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				l.Info("Connection closed")
				break
			}

			// 解析消息
			var wsMsg struct {
				BizCode string         `json:"biz_code"`
				Data    map[string]any `json:"data"`
			}
			if err := sonic.Unmarshal(msg, &wsMsg); err != nil {
				l.Error("解析 WebSocket 消息失败", "error", err)
				continue
			}

			// 根据业务码处理不同类型的消息
			switch wsMsg.BizCode {
			case dto.WSBizCodeDeviceOnline:
				// 设备上线
				if sn, ok := wsMsg.Data["sn"].(string); ok {
					payload := &event.GatewayOnlinePayload{
						GatewayEventPayload: event.GatewayEventPayload{
							SN:         sn,
							Timestamp:  wsMsg.Data["timestamp"].(int64),
							Properties: wsMsg.Data,
						},
					}
					if modelID, ok := wsMsg.Data["model_id"].(float64); ok {
						payload.ModelID = uint(modelID)
					}
					eb.Publish(event.GatewayOnlineEvent, payload)
				}

			case dto.WSBizCodeDeviceOffline:
				// 设备离线
				if sn, ok := wsMsg.Data["sn"].(string); ok {
					reason := ""
					if r, ok := wsMsg.Data["reason"].(string); ok {
						reason = r
					}
					payload := &event.GatewayOfflinePayload{
						GatewayEventPayload: event.GatewayEventPayload{
							SN:         sn,
							Timestamp:  wsMsg.Data["timestamp"].(int64),
							Properties: wsMsg.Data,
						},
						Reason: reason,
					}
					eb.Publish(event.GatewayOfflineEvent, payload)
				}

			case dto.WSBizCodeDeviceUpdateTopo:
				// 拓扑更新
				if gatewaySN, ok := wsMsg.Data["gateway_sn"].(string); ok {
					if connectedDrones, ok := wsMsg.Data["connected_drones"].([]string); ok {
						payload := &event.GatewayUpdateTopoPayload{
							GatewaySN:       gatewaySN,
							ConnectedDrones: connectedDrones,
						}
						eb.Publish(event.GatewayUpdateTopoEvent, payload)
					}
				}
			}

			// 回复客户端
			if err = c.WriteMessage(mt, msg); err != nil {
				l.Error("Failed to write message", "error", err)
				break
			}
		}
	}))
	cr.Start()
}
