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
	if err := r.tx.Where("area_id = ?", id).First(&area).Error; err != nil {
		r.l.Error("通过ID查询区域失败", slog.Any("id", id), slog.Any("error", err))
		return nil, err
	}
	return &area, nil
}

// SelectByName 根据名称查询区域
func (r *AreaDefaultRepo) SelectByName(ctx context.Context, name string) (*po.Area, error) {
	var area po.Area
	if err := r.tx.Where("area_name = ?", name).First(&area).Error; err != nil {
		r.l.Error("通过名称查询区域失败", slog.Any("name", name), slog.Any("error", err))
		return nil, err
	}
	return &area, nil
}

// FetchAll 查询所有区域
func (r *AreaDefaultRepo) SelectAll(ctx context.Context, name string, created_at_begin, created_at_end string, page, pageSize int) ([]*po.Area, int64, error) {
	var areas []*po.Area
	query := r.tx.WithContext(ctx).Where("state = 0")
	if name != "" {
		query = query.Where("area_name LIKE ?", "%"+name+"%")
	}
	if created_at_begin != "" {
		// 如果只有日期没有时间，默认时间为00:00:00
		if len(created_at_begin) == 10 {
			created_at_begin += " 00:00:00"
		}
		query = query.Where("created_time >= ?", created_at_begin)
	}
	if created_at_end != "" {
		// 如果只有日期没有时间，默认时间为23:59:59
		if len(created_at_end) == 10 {
			created_at_end += " 23:59:59"
		}
		query = query.Where("created_time <= ?", created_at_end)
	}

	var total int64
	err := query.Model(&po.Area{}).Count(&total).Error
	if err != nil {
		r.l.Error("查询区域总数失败", slog.Any("name", name), slog.Any("error", err))
		return nil, 0, err
	}
	r.l.Info("查询区域总数", slog.Any("name", name), slog.Any("total", total))
	if total == 0 {
		return nil, 0, nil // 如果没有数据，直接返回空切片
	}

	if page > 0 && pageSize > 0 {
		query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	if err := query.Find(&areas).Error; err != nil {
		r.l.Error("查询所有区域失败", slog.Any("name", name), slog.Any("error", err))
		return nil, 0, err
	}

	return areas, total, nil
}

// DeleteByID 根据ID删除区域
func (r *AreaDefaultRepo) DeleteByID(ctx context.Context, id uint) error {
	if err := r.tx.Unscoped().Where("area_id = ?", id).Delete(&po.Area{}).Error; err != nil {
		r.l.Error("删除区域失败", slog.Any("id", id), slog.Any("error", err))
		return err
	}
	return nil
}
