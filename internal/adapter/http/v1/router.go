package v1

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log/slog"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
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
//	@host			localhost:10086
//	@BasePath		/api/v1
func NewRouter(app *fiber.App, eb EventBus.Bus, l *slog.Logger, user service.UserSvc, drone service.DroneSvc) {
	sfCfg := slogfiber.Config{
		WithTraceID: true,
	}
	app.Use(slogfiber.NewWithConfig(l, sfCfg))
	app.Use(recover.New())
	app.Use(cors.New())

	// Swagger
	app.Use(swagger.New(swagger.Config{
		BasePath: "/",
		FilePath: "./docs/openapi/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}))
	// swaggerHandler := ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "DISABLE_SWAGGER_HTTP_HANDLER")
	// e.GET("/swagger/*any", swaggerHandler)

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
		newUserRouter(api, user, eb, l)
		newDroneRouter(api, drone, eb, l)
	}
}
