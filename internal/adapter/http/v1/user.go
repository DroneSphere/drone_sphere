package v1

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/adapter/http/middleware"
	"github.com/dronesphere/internal/event"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/pkg/token"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
)

// RegisterRequest 用户注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Email    string `json:"email" binding:"required"`    // 电子邮件
	Avatar   string `json:"avatar"`                      // 头像URL
	Password string `json:"password" binding:"required"` // 密码
}

// LoginRequest 用户登录请求结构
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`                    // 电子邮件
	Password string `json:"password" example:"admin" binding:"required"` // 密码
	SN       string `json:"sn" example:"123456"`                         // 遥控器SN，仅Pilot端登录时需要提供
}

// UserResult 用户信息结构
type UserResult struct {
	ID          uint   `json:"id"`           // 用户ID
	Username    string `json:"username"`     // 用户名
	Email       string `json:"email"`        // 电子邮件
	Avatar      string `json:"avatar"`       // 头像URL
	CreatedTime string `json:"created_time"` // 创建时间
	UpdatedTime string `json:"updated_time"` // 更新时间
}

// WorkspaceResult 工作空间信息结构
type WorkspaceResult struct {
	ID   uint   `json:"id"`   // 工作空间ID
	Name string `json:"name"` // 工作空间名称
	Type string `json:"type"` // 工作空间类型
}

// LoginResult 登录结果结构
type LoginResult struct {
	Token     string          `json:"token"`     // 访问令牌
	User      UserResult      `json:"user"`      // 用户信息
	Workspace WorkspaceResult `json:"workspace"` // 工作空间信息
}

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

		// 需要认证的路由
		authenticated := h.Use(middleware.Authenticate)
		authenticated.Get("/", r.get)
		authenticated.Get("/list", r.listUsers)                  // 获取用户列表
		authenticated.Post("/change-password", r.changePassword) // 修改密码
		authenticated.Post("/create", r.createUser)              // 创建用户
	}
}

func (r *UserRouter) register(c *fiber.Ctx) error {
	req := new(RegisterRequest)
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

	res := LoginResult{
		Token: t,
		User: UserResult{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Avatar:   u.Avatar,
		},
		Workspace: WorkspaceResult{
			ID:   1,
			Name: "default",
		},
	}

	return c.JSON(Success(res))
}

func (r *UserRouter) login(c *fiber.Ctx) error {
	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}
	r.l.Info("用户登录", slog.Any("req", req))

	// 执行登录操作
	u, err := r.svc.Login(req.Email, req.Password)
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

	res := LoginResult{
		Token: t,
		User: UserResult{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Avatar:   u.Avatar,
		},
		Workspace: WorkspaceResult{
			ID:   1,
			Name: "default",
		},
	}

	return c.JSON(Success(res))
}

func (r *UserRouter) get(c *fiber.Ctx) error {
	claims := c.Locals(middleware.UserClaimsKey).(*token.CustomClaims)
	r.l.Info("获取用户信息", slog.Any("claims", claims))
	u, err := r.svc.Repo().SelectByID(claims.UserID)
	if err != nil {
		r.l.Warn("获取用户信息失败", slog.Any("claims", claims), slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	res := LoginResult{
		User: UserResult{
			ID:       u.ID,
			Username: u.Username,
			Email:    u.Email,
			Avatar:   u.Avatar,
		},
		Workspace: WorkspaceResult{
			ID:   1,
			Name: "default",
		},
	}

	return c.JSON(Success(res))
}

func (r *UserRouter) listUsers(c *fiber.Ctx) error {
	// 用户列表响应结构体（在函数内部定义仅使用一次的结构体）
	type UserListResult struct {
		Users []UserResult `json:"users"` // 用户列表
		Total int64        `json:"total"` // 总用户数
	}

	// 获取用户信息
	users, total, err := r.svc.GetAllUsers()
	if err != nil {
		r.l.Warn("获取用户列表失败", slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	// 构建返回结果
	userResults := make([]UserResult, 0, len(users))
	for _, u := range users {
		userResults = append(userResults, UserResult{
			ID:          u.ID,
			Username:    u.Username,
			Email:       u.Email,
			Avatar:      u.Avatar,
			CreatedTime: u.CreatedTime.Format("2006-01-02 15:04:05"),
			UpdatedTime: u.UpdatedTime.Format("2006-01-02 15:04:05"),
		})
	}

	res := UserListResult{
		Users: userResults,
		Total: total,
	}

	return c.JSON(Success(res))
}

func (r *UserRouter) changePassword(c *fiber.Ctx) error {
	// 修改密码请求结构（在函数内部定义仅使用一次的结构体）
	type ChangePasswordRequest struct {
		UserID      uint   `json:"userId" binding:"required"`      // 用户ID
		OldPassword string `json:"oldPassword" binding:"required"` // 旧密码
		NewPassword string `json:"newPassword" binding:"required"` // 新密码
	}

	req := new(ChangePasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}
	r.l.Info("修改用户密码", slog.Any("userID", req.UserID))

	// 执行密码修改操作
	err := r.svc.ChangePassword(req.UserID, req.OldPassword, req.NewPassword)
	if err != nil {
		r.l.Warn("修改密码失败", slog.Any("userID", req.UserID), slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	return c.JSON(Success(nil))
}

func (r *UserRouter) createUser(c *fiber.Ctx) error {
	// 创建用户请求结构（在函数内部定义仅使用一次的结构体）
	type CreateUserRequest struct {
		Username string `json:"username" binding:"required"` // 用户名
		Email    string `json:"email" binding:"required"`    // 邮箱
		Avatar   string `json:"avatar"`                      // 头像URL
		Password string `json:"password" binding:"required"` // 密码
	}

	req := new(CreateUserRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}
	r.l.Info("创建新用户", slog.Any("req", req))

	// 创建用户实体
	u := entity.NewUser(
		req.Username,
		req.Email,
		req.Avatar,
		req.Password,
	)

	// 执行创建用户操作
	err := r.svc.CreateUser(u)
	if err != nil {
		r.l.Warn("创建用户失败", slog.Any("req", req), slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	// 构建返回结果
	res := UserResult{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Avatar:   u.Avatar,
	}

	return c.JSON(Success(res))
}
