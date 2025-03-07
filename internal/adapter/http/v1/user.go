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
//	@Router			/user/login [post]
//	@Summary		Web/Pilot端统一用户登录
//	@Description	Web/Pilot端统一用户登录，根据是否携带 SN 切换登录方式
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.LoginRequest						true	"登录参数"
//	@Success		200		{object}	v1.Response{data=v1.LoginResult}	"成功"
func (r *UserRouter) login(c *fiber.Ctx) error {
	req := new(api.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}
	r.l.Info("用户登录", slog.Any("req", req))

	// 执行登录操作
	token, err := r.svc.Login(req.Username, req.Password)
	if err != nil {
		r.l.Warn("登录失败", slog.Any("req", req), slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	// 根据是否携带 SN 发送登录成功事件
	// 携带 SN 说明是 Pilot 端登录，需要发送登录成功事件
	if req.SN != "" {
		ctx := context.WithValue(context.Background(), event.RemoteControllerLoginSNKey, req.SN)
		r.eb.Publish(event.RemoteControllerLoggedIn, ctx)
	}

	res := api.LoginResult{
		Token: token,
		User: api.UserResult{
			ID:       "1",
			Username: req.Username,
		},
		Platform: api.PlatformResult{
			Platform:  "DroneSphere",
			Workspace: "default",
			Desc:      "Default workspace for demo",
		},
		Params: api.ParamsResult{
			MQTTHost:     "tcp://47.245.40.222:1883",
			MQTTUsername: "drone",
			MQTTPassword: "drone",
		},
	}

	return c.JSON(Success(res))
}
