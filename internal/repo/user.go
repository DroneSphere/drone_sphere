package repo

import (
	"gorm.io/gorm"
	"log/slog"
)

type UserGormRepo struct {
	tx *gorm.DB
	l  *slog.Logger
}

func NewUserGormRepo(db *gorm.DB, l *slog.Logger) *UserGormRepo {
	return &UserGormRepo{
		tx: db,
		l:  l,
	}
}

func (r *UserGormRepo) GetByUsername(username string) (string, error) {
	return "Hi", nil
}
