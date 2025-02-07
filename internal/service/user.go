package service

import "log/slog"

type (
	UserSvc interface {
		Login(username, password string) (string, error)
	}

	UserRepo interface {
		GetByUsername(username string) (string, error)
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

func (s *UserImpl) Login(username, password string) (string, error) {
	u, err := s.r.GetByUsername(username)
	if err != nil {
		return "", err
	}
	if u != password {
		return "", nil
	}
	return "token", nil
}
