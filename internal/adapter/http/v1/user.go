package v1

import (
	"context"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/event"
	"log/slog"

	"github.com/asaskevich/EventBus"
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

// login 用户登录接口
//
//		@Router			/user/login [post]
//		@Summary		Web/Pilot端统一用户登录
//		@Description 	Web/Pilot端统一用户登录，根据是否携带 SN 切换登录方式
//		@Tags			user
//		@Accept			json
//		@Produce		json
//	    @Param request	body		v1.LoginRequest 	true	"登录参数"
//		@Success		200	{object}	v1.Response{data=string}	"成功"
func (r *UserRouter) login(c *fiber.Ctx) error {
	req := new(api.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		r.l.Info("InvalidParams", slog.Any("err", err))
		c.Status(fiber.StatusBadRequest)
		return c.JSON(Fail(InvalidParams))
	}
	r.l.Info("Login", slog.Any("req", req))

	token, err := r.svc.Login(req.Username, req.Password)
	if err != nil {
		r.l.Warn("LoginError", slog.Any("err", err))
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	if req.SN != "" {
		ctx := context.WithValue(context.Background(), "sn", req.SN)
		r.eb.Publish(event.UserLoginSuccessEvent, ctx)
	}

	return c.JSON(Success(token))
}
