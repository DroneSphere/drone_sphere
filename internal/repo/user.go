package repo

import (
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"gorm.io/gorm"
	"log/slog"
)

type UserGormRepo struct {
	tx *gorm.DB
	l  *slog.Logger
}

func NewUserGormRepo(db *gorm.DB, l *slog.Logger) *UserGormRepo {
	_ = db.AutoMigrate(&po.User{})
	return &UserGormRepo{
		tx: db,
		l:  l,
	}
}

func (r *UserGormRepo) toEntity(user po.User) entity.User {
	return entity.User{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Password: user.Password,
	}
}

func (r *UserGormRepo) GetByUsername(username string) (string, error) {
	return "Hi", nil
}

func (r *UserGormRepo) SaveUser(user entity.User) error {
	p := po.User{
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Password: user.Password,
	}
	return r.tx.Save(&p).Error
}

func (r *UserGormRepo) SelectByID(id uint) (entity.User, error) {
	var u po.User
	if err := r.tx.First(&u, id).Error; err != nil {
		return entity.User{}, err
	}
	return r.toEntity(u), nil
}

func (r *UserGormRepo) SelectByUsername(username string) (entity.User, error) {
	var u po.User
	if err := r.tx.Where("username = ?", username).First(&u).Error; err != nil {
		return entity.User{}, err
	}
	return r.toEntity(u), nil
}

func (r *UserGormRepo) UpdatePasswordByUsername(username, password string) error {
	return r.tx.Model(&po.User{}).Where("username = ?", username).Update("password", password).Error
}
