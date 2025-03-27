package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/vo"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hashicorp/go-multierror"
)

type (
	AreaSvc interface {
		Repo() AreaRepo
		CreateArea(ctx context.Context, name, description string, points []vo.GeoPoint) (*entity.Area, error)
		UpdateArea(ctx context.Context, id uint, name, description string, points []vo.GeoPoint) (*entity.Area, error)
		FetchArea(ctx context.Context, params struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		}) (*entity.Area, error)
		FetchAll(ctx context.Context, name string) ([]*entity.Area, error)
	}
	AreaRepo interface {
		Save(ctx context.Context, area *entity.Area) error
		SelectByID(ctx context.Context, id uint) (*entity.Area, error)
		SelectByName(ctx context.Context, name string) (*entity.Area, error)
		FetchAll(ctx context.Context, name string) ([]*entity.Area, error)
		DeleteByID(ctx context.Context, id uint) error
	}
)

type AreaImpl struct {
	r    AreaRepo
	l    *slog.Logger
	mqtt mqtt.Client
}

func NewAreaImpl(r AreaRepo, l *slog.Logger, mqtt mqtt.Client) AreaSvc {
	return &AreaImpl{
		r:    r,
		l:    l,
		mqtt: mqtt,
	}
}

func (s *AreaImpl) Repo() AreaRepo {
	return s.r
}

func (s *AreaImpl) CreateArea(ctx context.Context, name, description string, points []vo.GeoPoint) (*entity.Area, error) {
	entity := &entity.Area{
		Name:        name,
		Description: description,
		Points:      points,
	}
	_ = entity.CalcCenter()
	err := s.r.Save(ctx, entity)
	if err != nil {
		s.l.Error("SaveArea Error: ", slog.Any("error", err))
		return nil, err
	}
	area, err := s.r.SelectByName(ctx, name)
	if err != nil {
		s.l.Error("SelectByName Error: ", slog.Any("error", err))
		return nil, err
	}
	return area, nil
}

func (s *AreaImpl) UpdateArea(ctx context.Context, id uint, name, description string, points []vo.GeoPoint) (*entity.Area, error) {
	// 查询区域信息
	entity, err := s.r.SelectByID(ctx, id)
	if err != nil {
		s.l.Error("SelectByID Error: ", slog.Any("error", err))
		return nil, err
	}

	// 更新区域信息
	if name != "" {
		entity.Name = name
	}
	if description != "" {
		entity.Description = description
	}
	if len(points) > 0 {
		entity.Points = points
	}
	_ = entity.CalcCenter()

	// 保存更新后的区域
	err = s.r.Save(ctx, entity)
	if err != nil {
		s.l.Error("SaveArea Error: ", slog.Any("error", err))
		return nil, err
	}

	// 返回更新后的区域
	area, err := s.r.SelectByID(ctx, id)
	if err != nil {
		s.l.Error("SelectByID Error: ", slog.Any("error", err))
		return nil, err
	}
	return area, nil
}

func (s *AreaImpl) FetchArea(ctx context.Context, params struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}) (*entity.Area, error) {
	var area *entity.Area
	var multiErr error
	if params.ID != 0 {
		a, err := s.r.SelectByID(ctx, params.ID)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		area = a
	} else if params.Name != "" {
		a, err := s.r.SelectByName(ctx, params.Name)
		if err != nil {
			multiErr = multierror.Append(multiErr)
		}
		area = a
	} else {
		return nil, errors.New("invalid params")
	}
	if multiErr != nil {
		s.l.Error("FetchArea Error: ", slog.Any("error", multiErr))
		return nil, multiErr
	}

	return area, nil
}

func (s *AreaImpl) FetchAll(ctx context.Context, name string) ([]*entity.Area, error) {
	areas, err := s.r.FetchAll(ctx, name)
	if err != nil {
		s.l.Error("FetchList Error: ", slog.Any("error", err))
		return nil, err
	}
	return areas, nil
}
