package ws

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/robfig/cron/v3"
	slogfiber "github.com/samber/slog-fiber"
	"log/slog"
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
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	cr := cron.New(cron.WithSeconds())
	app.Get("/", websocket.New(func(c *websocket.Conn) {
		var err error

		//	每秒发送一次
		_, err = cr.AddFunc("0/1 * * * * *", func() {
			DeviceOSDBroadcast(c, l, drone)
		})
		if err != nil {
			l.Error("Cron func create failed", slog.Any("err", err))
		}

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				l.Info("MSG Received")
				break
			}

			if err = c.WriteMessage(mt, msg); err != nil {
				l.Info("write:", err)
				break
			}
		}
	}))
	cr.Start()
}
