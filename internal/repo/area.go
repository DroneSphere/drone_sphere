package repo

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/vo"
	"github.com/jinzhu/copier"
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

func (r *AreaDefaultRepo) toEntity(p *po.Area) *entity.Area {
	var points []vo.GeoPoint
	for _, point := range p.Points {
		var p vo.GeoPoint
		if err := copier.Copy(&p, point); err != nil {
			r.l.Error("Copy Error: ", slog.Any("error", err))
			return nil
		}
		points = append(points, p)
	}
	var area entity.Area
	if err := copier.Copy(&area, p); err != nil {
		r.l.Error("Copy Error: ", slog.Any("error", err))
		return nil
	}
	area.Points = points
	return &area
}

func (r *AreaDefaultRepo) Save(ctx context.Context, area *entity.Area) error {
	p := po.Area{}
	_ = copier.Copy(&p, area)
	if err := r.tx.Save(&p).Error; err != nil {
		r.l.Error("Create Error: ", slog.Any("error", err))
		return err
	}
	return nil
}

func (r *AreaDefaultRepo) SelectByID(ctx context.Context, id uint) (*entity.Area, error) {
	var p po.Area
	if err := r.tx.Where("id = ?", id).First(&p).Error; err != nil {
		r.l.Error("First Error: ", slog.Any("error", err))
		return nil, err
	}
	area := r.toEntity(&p)
	return area, nil
}

func (r *AreaDefaultRepo) SelectByName(ctx context.Context, name string) (*entity.Area, error) {
	var p po.Area
	if err := r.tx.Where("name = ?", name).First(&p).Error; err != nil {
		r.l.Error("First Error: ", slog.Any("error", err))
		return nil, err
	}
	area := r.toEntity(&p)
	return area, nil
}

func (r *AreaDefaultRepo) FetchAll(ctx context.Context, name string) ([]*entity.Area, error) {
	var ps []po.Area
	tx := r.tx
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if err := tx.Find(&ps).Error; err != nil {
		r.l.Error("Find Error: ", slog.Any("error", err))
		return nil, err
	}
	var areas []*entity.Area
	for _, p := range ps {
		area := r.toEntity(&p)
		areas = append(areas, area)
	}
	return areas, nil
}

func (r *AreaDefaultRepo) DeleteByID(ctx context.Context, id uint) error {
	if err := r.tx.Unscoped().Where("id = ?", id).Delete(&po.Area{}).Error; err != nil {
		r.l.Error("Delete Error: ", slog.Any("error", err))
		return err
	}
	return nil
}
