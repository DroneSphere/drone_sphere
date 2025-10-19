package service

import (
	"errors"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/pkg/misc"
)

type (
	UserSvc interface {
		Repo() UserRepo
		Login(email, password string) (entity.User, error)
		Register(user entity.User) error
		// GetAllUsers 获取所有用户列表
		GetAllUsers() ([]entity.User, int64, error)
		// ChangePassword 修改用户密码
		ChangePassword(userID uint, oldPassword, newPassword string) error
		// CreateUser 管理员创建新用户
		CreateUser(user entity.User) error
		// Logout 用户登出，将token加入黑名单
		Logout(tokenString string) error
	}

	UserRepo interface {
		GetByUsername(username string) (string, error)
		SaveUser(user entity.User) error
		SelectByID(id uint) (entity.User, error)
		SelectByUsername(username string) (entity.User, error)
		SelectByEmail(email string) (entity.User, error)
		UpdatePasswordByUsername(username, password string) error
		// SelectAll 查询所有用户
		SelectAll() ([]entity.User, int64, error)
		// UpdatePasswordByID 根据用户ID更新密码
		UpdatePasswordByID(id uint, password string) error
	}
)

type UserImpl struct {
	r              UserRepo
	l              *slog.Logger
	blacklistSvc   TokenBlacklistService // token黑名单服务
}

func NewUserSvc(r UserRepo, l *slog.Logger) UserSvc {
	return &UserImpl{
		r:            r,
		l:            l,
		blacklistSvc: NewTokenBlacklistService(),
	}
}

func (s *UserImpl) Repo() UserRepo {
	return s.r
}

const (
	ErrInvalidPassword = "invalid password"
	ErrInvalidToken    = "invalid token"
)

func (s *UserImpl) Login(email, password string) (entity.User, error) {
	u, err := s.r.SelectByEmail(email)
	if err != nil {
		return entity.User{}, err
	}
	s.l.Info("用户登录", slog.Any("user", u))
	if !misc.ComparePassword(password, u.Password) {
		s.l.Warn("密码错误", slog.Any("user", u))
		return entity.User{}, errors.New(ErrInvalidPassword)
	}
	return entity.User{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Avatar:   u.Avatar,
	}, nil
}

func (s *UserImpl) Register(user entity.User) error {
	// 先加密密码，然后一次性保存用户
	encryptedPassword := misc.EncryptWithBcrypt(user.Password)
	user.Password = encryptedPassword
	return s.r.SaveUser(user)
}

// GetAllUsers 获取所有用户列表
func (s *UserImpl) GetAllUsers() ([]entity.User, int64, error) {
	users, total, err := s.r.SelectAll()
	if err != nil {
		s.l.Error("获取用户列表失败", slog.Any("err", err))
		return nil, 0, err
	}

	// 清除密码信息
	for i := range users {
		users[i].Password = ""
	}

	return users, total, nil
}

// ChangePassword 修改用户密码
func (s *UserImpl) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 查询用户
	user, err := s.r.SelectByID(userID)
	if err != nil {
		s.l.Error("查询用户失败", slog.Any("userID", userID), slog.Any("err", err))
		return err
	}

	// 验证旧密码
	if !misc.ComparePassword(oldPassword, user.Password) {
		s.l.Warn("旧密码验证失败", slog.Any("userID", userID))
		return errors.New("旧密码错误")
	}

	// 加密新密码并更新
	encryptedPassword := misc.EncryptWithBcrypt(newPassword)
	err = s.r.UpdatePasswordByID(userID, encryptedPassword)
	if err != nil {
		s.l.Error("更新密码失败", slog.Any("userID", userID), slog.Any("err", err))
		return err
	}

	return nil
}

// CreateUser 管理员创建新用户
func (s *UserImpl) CreateUser(user entity.User) error {
	// 检查邮箱是否已存在
	_, err := s.r.SelectByEmail(user.Email)
	if err == nil {
		// 邮箱已存在
		s.l.Warn("创建用户失败，邮箱已存在", slog.Any("email", user.Email))
		return errors.New("邮箱已存在")
	}

	// 检查用户名是否已存在
	_, err = s.r.SelectByUsername(user.Username)
	if err == nil {
		// 用户名已存在
		s.l.Warn("创建用户失败，用户名已存在", slog.Any("username", user.Username))
		return errors.New("用户名已存在")
	}

	// 保存用户
	if err := s.r.SaveUser(user); err != nil {
		s.l.Error("保存用户失败", slog.Any("user", user), slog.Any("err", err))
		return err
	}

	// 更新密码（加密）
	return s.r.UpdatePasswordByUsername(user.Username, misc.EncryptWithBcrypt(user.Password))
}

// Logout 用户登出，将token加入黑名单
func (s *UserImpl) Logout(tokenString string) error {
	if tokenString == "" {
		return errors.New("token不能为空")
	}

	// 将token加入黑名单
	if err := s.blacklistSvc.AddToken(tokenString); err != nil {
		s.l.Error("将token加入黑名单失败", slog.Any("err", err))
		return err
	}

	s.l.Info("用户登出成功，token已加入黑名单", slog.String("token_prefix", tokenString[:min(len(tokenString), 20)]+"..."))

	// 启动一个goroutine来清理过期的token
	go s.blacklistSvc.RemoveExpiredTokens()

	return nil
}

// GetBlacklistService 获取token黑名单服务（供中间件使用）
func (s *UserImpl) GetBlacklistService() TokenBlacklistService {
	return s.blacklistSvc
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
