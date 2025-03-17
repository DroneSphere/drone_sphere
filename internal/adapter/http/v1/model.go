package v1

import (
	"context"
	"log/slog"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
)

type ModelRouter struct {
	svc service.ModelSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func NewModelsRouter(handler fiber.Router, svc service.ModelSvc, eb EventBus.Bus, l *slog.Logger) {
	r := &ModelRouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/models")
	{
		h.Get("/gateways", r.getGatewayModels)
		h.Get("/drones", r.getDroneModels)
		h.Get("/gimbals", r.getGimbalModels)
		h.Get("/payloads", r.getPaylodModels)
	}

}

func (r *ModelRouter) getGatewayModels(c *fiber.Ctx) error {
	ctx := context.Background()
	models, err := r.svc.Repo().SelectAllGatewayModel(ctx)
	if err != nil {
		r.l.Error("get gateway models", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(models))
}

func (r *ModelRouter) getDroneModels(c *fiber.Ctx) error {
	ctx := context.Background()
	res, err := r.svc.Repo().SelectAllDroneModel(ctx)
	if err != nil {
		r.l.Error("get drone models", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(res))
}

func (r *ModelRouter) getGimbalModels(c *fiber.Ctx) error {
	ctx := context.Background()
	res, err := r.svc.Repo().SelectAllGimbals(ctx)
	if err != nil {
		r.l.Error("get gimbal models", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(res))
}

func (r *ModelRouter) getPaylodModels(c *fiber.Ctx) error {
	ctx := context.Background()
	res, err := r.svc.Repo().SelectAllPayloadModel(ctx)
	if err != nil {
		r.l.Error("get payload models", "error", err)
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(res))
}
