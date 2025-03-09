package v1

import (
	"context"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/adapter/http/middleware"
	"github.com/dronesphere/internal/event"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/pkg/token"
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
		h.Post("/register", r.register)
		h.Post("/login", r.login)
		h.Use(middleware.Authenticate).Get("/", r.
			get)
	}
}

// register 用户注册接口
//
//	@Router			/user/register [post]
//	@Summary		用户注册
//	@Description	用户注册
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		v1.RegisterRequest					true	"注册参数"
//	@Success		200		{object}	v1.Response{data=v1.LoginResult}	"成功"
func (r *UserRouter) register(c *fiber.Ctx) error {
	req := new(api.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}
	r.l.Info("用户注册", slog.Any("req", req))

	// 执行注册操作
	u := entity.NewUser(
		req.Username,
		req.Email,
		req.Avatar,
		req.Password,
	)
	err := r.svc.Register(u)
	if err != nil {
		r.l.Warn("注册失败", slog.Any("req", req), slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	// 注册成功后自动登录
	// 生成令牌
	t, err := token.GenerateToken(u.ID, u.Username)
	if err != nil {
		r.l.Error("生成令牌失败", slog.Any("req", req), slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	res := api.LoginResult{
		Token: t,
		User: api.UserResult{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Avatar:   u.Avatar,
		},
		Workspace: api.WorkspaceResult{
			ID:   1,
			Name: "default",
		},
	}

	return c.JSON(Success(res))
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
	u, err := r.svc.Login(req.Username, req.Password)
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

	t, err := token.GenerateToken(u.ID, u.Username)
	if err != nil {
		r.l.Error("生成令牌失败", slog.Any("req", req), slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	res := api.LoginResult{
		Token: t,
		User: api.UserResult{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Avatar:   u.Avatar,
		},
		Workspace: api.WorkspaceResult{
			ID:   1,
			Name: "default",
		},
	}

	return c.JSON(Success(res))
}

// get 获取用户信息
//
//	@Router			/user [get]
//	@Summary		获取用户信息
//	@Description	获取用户信息
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=v1.LoginResult}	"成功"
func (r *UserRouter) get(c *fiber.Ctx) error {
	claims := c.Locals(middleware.UserClaimsKey).(*token.CustomClaims)
	r.l.Info("获取用户信息", slog.Any("claims", claims))
	u, err := r.svc.Repo().SelectByID(claims.UserID)
	if err != nil {
		r.l.Warn("获取用户信息失败", slog.Any("claims", claims), slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	res := api.LoginResult{
		User: api.UserResult{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Avatar:   u.Avatar,
		},
		Workspace: api.WorkspaceResult{
			ID:   1,
			Name: "default",
		},
	}

	return c.JSON(Success(res))
}
