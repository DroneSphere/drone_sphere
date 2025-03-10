package service

import (
	"errors"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/pkg/misc"
	"log/slog"
)

type (
	UserSvc interface {
		Repo() UserRepo
		Login(email, password string) (entity.User, error)
		Register(user entity.User) error
	}

	UserRepo interface {
		GetByUsername(username string) (string, error)
		SaveUser(user entity.User) error
		SelectByID(id uint) (entity.User, error)
		SelectByUsername(username string) (entity.User, error)
		SelectByEmail(email string) (entity.User, error)
		UpdatePasswordByUsername(username, password string) error
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
