package repo

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/po"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AreaDefaultRepo struct {
	tx        *gorm.DB
	rds       *redis.Client
	l         *slog.Logger
	rdsPrefix string
}

func NewAreaDefaultRepo(db *gorm.DB, rds *redis.Client, l *slog.Logger) *AreaDefaultRepo {
	return &AreaDefaultRepo{
		tx:        db,
		rds:       rds,
		l:         l,
		rdsPrefix: "drone:",
	}
}

// Save 保存区域信息
func (r *AreaDefaultRepo) Save(ctx context.Context, area *po.Area) error {
	if err := r.tx.Save(area).Error; err != nil {
		r.l.Error("保存区域失败", slog.Any("error", err))
		return err
	}
	return nil
}

// SelectByID 根据ID查询区域
func (r *AreaDefaultRepo) SelectByID(ctx context.Context, id uint) (*po.Area, error) {
	var area po.Area
	if err := r.tx.Where("id = ?", id).First(&area).Error; err != nil {
		r.l.Error("通过ID查询区域失败", slog.Any("id", id), slog.Any("error", err))
		return nil, err
	}
	return &area, nil
}

// SelectByName 根据名称查询区域
func (r *AreaDefaultRepo) SelectByName(ctx context.Context, name string) (*po.Area, error) {
	var area po.Area
	if err := r.tx.Where("name = ?", name).First(&area).Error; err != nil {
		r.l.Error("通过名称查询区域失败", slog.Any("name", name), slog.Any("error", err))
		return nil, err
	}
	return &area, nil
}

// FetchAll 查询所有区域
func (r *AreaDefaultRepo) FetchAll(ctx context.Context, name string) ([]*po.Area, error) {
	var areas []*po.Area
	tx := r.tx
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if err := tx.Find(&areas).Error; err != nil {
		r.l.Error("查询所有区域失败", slog.Any("name", name), slog.Any("error", err))
		return nil, err
	}
	return areas, nil
}

// DeleteByID 根据ID删除区域
func (r *AreaDefaultRepo) DeleteByID(ctx context.Context, id uint) error {
	if err := r.tx.Unscoped().Where("id = ?", id).Delete(&po.Area{}).Error; err != nil {
		r.l.Error("删除区域失败", slog.Any("id", id), slog.Any("error", err))
		return err
	}
	return nil
}
