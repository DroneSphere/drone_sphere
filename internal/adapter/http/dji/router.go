package dji

import (
	"log/slog"

	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	slogfiber "github.com/samber/slog-fiber"
)

// NewRouter -.
// Swagger spec:
//
//	@title			上云API模块API
//	@description	上云API需要的API模块接口
//	@version		1.0
//	@license.name	Apache 2.0
//	@host			example
//	@BasePath		/
func NewRouter(app *fiber.App, eb EventBus.Bus, l *slog.Logger, drone service.DroneSvc, wayline service.WaylineSvc) {
	sfCfg := slogfiber.Config{
		WithTraceID: true,
	}
	app.Use(slogfiber.NewWithConfig(l, sfCfg))
	app.Use(recover.New())
	app.Use(cors.New())

	// Swagger
	// app.Use(swagger.New(swagger.Config{
	// 	BasePath: "/",
	// 	FilePath: "./docs/http/dji/swagger.json",
	// 	Path:     "swagger",
	// 	Title:    "DJI Swagger API Docs",
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
	api := app.Group("/")
	{
		newTSARouter(api, drone, eb, l)
		NewWaylineRouter(api, wayline, eb, l)
	}
}
