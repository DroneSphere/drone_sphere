package v1

import (
	"bufio"
	"fmt"
	"log/slog"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	slogfiber "github.com/samber/slog-fiber"
)

// NewRouter -.
// Swagger spec:
//
//	@title			DroneSphere API
//	@description	DroneSphere API
//	@version		1.0
//	@license.name	Apache 2.0
//	@host			lqhirwdzgkvv.sealoshzh.site
//	@BasePath		/api/v1
func NewRouter(app *fiber.App, eb EventBus.Bus, l *slog.Logger, user service.UserSvc, drone service.DroneSvc, sa service.AreaSvc, algo service.DetectAlgoSvc, wl service.WaylineSvc, job service.JobSvc, model service.ModelSvc) {
	sfCfg := slogfiber.Config{
		WithTraceID: true,
		WithSpanID:  true,
		Filters: []slogfiber.Filter{
			slogfiber.IgnorePath("/api/v1/sse"),
			slogfiber.IgnorePath("/api/v1/drone/state/sse"),
		},
	}
	app.Use(slogfiber.NewWithConfig(l, sfCfg))
	app.Use(recover.New())
	app.Use(cors.New())

	// Swagger
	// app.Use(swagger.New(swagger.Config{
	// 	BasePath: "/",
	// 	FilePath: "./docs/http/v1/swagger.json",
	// 	Path:     "swagger",
	// 	Title:    "Server Swagger API Docs",
	// }))

	// Prometheus metrics
	prometheus := fiberprometheus.New("dronesphere")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// K8s probe
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Routers
	api := app.Group("/api/v1")
	{
		newPlatformRouter(api, l)
		newUserRouter(api, user, eb, l)
		newDroneRouter(api, drone, eb, l)
		NewSearchAreaRouter(api, sa, eb, l)
		newDetectAlgoRouter(api, algo, l)
		newJobRouter(api, job, sa, model, l)
		// newWaylineRouter(api, wl, l)
		NewGatewayRouter(api, eb, l)
		NewModelsRouter(api, model, eb, l)
		api.Get("/sse", handleSSE(l))
	}
}

func handleSSE(l *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 设置SSE必需的响应头
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		// 使用流式响应
		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			l.Info("SSE connection established")
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case t := <-ticker.C:
					// 构造消息并尝试写入
					msg := fmt.Sprintf("data: %s\n\n", t.Format("2006-01-02 15:04:05"))
					l.Info("Sending message", "msg", msg)

					// 写入消息并立即刷新
					if _, err := w.WriteString(msg); err != nil {
						l.Error("SSE write error", "error", err)
						return
					}
					if err := w.Flush(); err != nil {
						l.Error("SSE flush error", "error", err)
						return
					}
					l.Info("Message sent and flushed")
				}
			}
		})

		return nil
	}
}
