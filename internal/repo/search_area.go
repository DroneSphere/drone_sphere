package repo

import (
	"context"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/jinzhu/copier"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
)

type SearchAreaGormRepo struct {
	tx        *gorm.DB
	rds       *redis.Client
	l         *slog.Logger
	rdsPrefix string
}

func NewSearchAreaGormRepo(db *gorm.DB, rds *redis.Client, l *slog.Logger) *SearchAreaGormRepo {
	//if err := db.AutoMigrate(&po.ORMSearchArea{}); err != nil {
	//	l.Error("Auto Migration Error: ", slog.Any("error", err))
	//	panic(err)
	//}
	return &SearchAreaGormRepo{
		tx:        db,
		rds:       rds,
		l:         l,
		rdsPrefix: "drone:",
	}
}

func (r *SearchAreaGormRepo) toEntity(p *po.ORMSearchArea) *entity.SearchArea {
	var points []entity.AreaPoint
	for _, point := range p.Points {
		var p entity.AreaPoint
		if err := copier.Copy(&p, point); err != nil {
			r.l.Error("Copy Error: ", slog.Any("error", err))
			return nil
		}
		points = append(points, p)
	}
	var area entity.SearchArea
	if err := copier.Copy(&area, p); err != nil {
		r.l.Error("Copy Error: ", slog.Any("error", err))
		return nil
	}
	area.Points = points
	return &area
}

func (r *SearchAreaGormRepo) toPO(area *entity.SearchArea) *po.ORMSearchArea {
	var p po.ORMSearchArea
	if err := copier.Copy(&p, area); err != nil {
		r.l.Error("Copy Error: ", slog.Any("error", err))
		return nil
	}
	var points []po.ORMAreaPoint
	for _, point := range area.Points {
		var p po.ORMAreaPoint
		if err := copier.Copy(&p, point); err != nil {
			r.l.Error("Copy Error: ", slog.Any("error", err))
			return nil
		}
		points = append(points, p)
	}
	p.Points = points
	return &p
}

func (r *SearchAreaGormRepo) Save(ctx context.Context, area *entity.SearchArea) (*entity.SearchArea, error) {
	p := r.toPO(area)
	if err := r.tx.Save(p).Error; err != nil {
		r.l.Error("Save Error: ", slog.Any("error", err))
		return nil, err
	}
	area.ID = p.ID
	return area, nil
}

func (r *SearchAreaGormRepo) FetchByID(ctx context.Context, id uint) (*entity.SearchArea, error) {
	var p po.ORMSearchArea
	if err := r.tx.Where("id = ?", id).First(&p).Error; err != nil {
		r.l.Error("First Error: ", slog.Any("error", err))
		return nil, err
	}
	area := r.toEntity(&p)
	return area, nil
}

func (r *SearchAreaGormRepo) FetchByName(ctx context.Context, name string) (*entity.SearchArea, error) {
	var p po.ORMSearchArea
	if err := r.tx.Where("name = ?", name).First(&p).Error; err != nil {
		r.l.Error("First Error: ", slog.Any("error", err))
		return nil, err
	}
	area := r.toEntity(&p)
	return area, nil
}

func (r *SearchAreaGormRepo) FetchAll(ctx context.Context) ([]*entity.SearchArea, error) {
	var ps []po.ORMSearchArea
	if err := r.tx.Find(&ps).Error; err != nil {
		r.l.Error("Find Error: ", slog.Any("error", err))
		return nil, err
	}
	var areas []*entity.SearchArea
	for _, p := range ps {
		area := r.toEntity(&p)
		areas = append(areas, area)
	}
	return areas, nil
}
