package v1

import (
	"context"
	"log/slog"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/event"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
)

type UserRouter struct {
	svc service.UserSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func newUserRouter(handler fiber.Router, svc service.UserSvc, eb EventBus.Bus, l *slog.Logger) {
	r := &UserRouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/user")
	{
		h.Post("/login", r.login)
	}
}

// login godoc
//
//	@Summary		login
//	@Description	User login
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	string
//	@Router			/user/login [post]
func (r *UserRouter) login(c *fiber.Ctx) error {
	r.l.Debug("login")
	ctx := context.WithValue(context.Background(), "sn", "111")
	r.eb.Publish(event.UserLoginSuccessEvent, ctx)
	return c.SendString("login")
}
