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
	r UserRepo
	l *slog.Logger
}

func NewUserSvc(r UserRepo, l *slog.Logger) UserSvc {
	return &UserImpl{
		r: r,
		l: l,
	}
}

func (s *UserImpl) Repo() UserRepo {
	return s.r
}

const ErrInvalidPassword = "invalid password"

func (s *UserImpl) Login(email, password string) (entity.User, error) {
	u, err := s.r.SelectByEmail(email)
	if err != nil {
		return entity.User{}, err
	}
	s.l.Info("用户登录", slog.Any("user", u))
	if misc.ComparePassword(password, u.Password) {
		s.l.Warn("密码错误", slog.Any("user", u), slog.Any("password", password))
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
	if err := s.r.SaveUser(user); err != nil {
		return err
	}
	return s.r.UpdatePasswordByUsername(user.Username, misc.EncryptWithBcrypt(user.Password))
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
	_, err := s.r.SelectByID(userID)
	if err != nil {
		s.l.Error("查询用户失败", slog.Any("userID", userID), slog.Any("err", err))
		return err
	}

	// 验证旧密码
	// if !misc.ComparePassword(oldPassword, user.Password) {
	// 	s.l.Warn("旧密码验证失败", slog.Any("userID", userID))
	// 	return errors.New("旧密码错误")
	// }

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
