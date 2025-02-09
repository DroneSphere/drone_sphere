package v1

import (
	"context"
	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type DroneRouter struct {
	svc service.DroneSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func newDroneRouter(handler fiber.Router, svc service.DroneSvc, eb EventBus.Bus, l *slog.Logger) {
	r := &DroneRouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/drone")
	{
		h.Get("/list", r.list)
	}
}

// list 列出所有无人机
//
//	@Router			/drone/list [get]
//	@Summary		列出所有无人机
//	@Description	列出所有绑定的无人机，包含不在线的无人机
//	@Tags			drone
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=[]v1.DroneItemResult}	"成功"
func (r *DroneRouter) list(c *fiber.Ctx) error {
	ctx := context.Background()
	drones, err := r.svc.ListAll(ctx)
	if err != nil {
		r.l.Warn("ListError", slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	return c.JSON(Success(drones))
}
