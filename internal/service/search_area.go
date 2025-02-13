package service

import (
	"context"
	"errors"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/model/entity"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hashicorp/go-multierror"
	"log/slog"
)

type (
	SearchAreaSvc interface {
		SaveArea(ctx context.Context, area *entity.SearchArea) (*entity.SearchArea, error)
		FetchArea(ctx context.Context, params *api.AreaFetchParams) (*entity.SearchArea, error)
		FetchList(ctx context.Context) ([]*entity.SearchArea, error)
	}
	SearchAreaRepo interface {
		Save(ctx context.Context, area *entity.SearchArea) (*entity.SearchArea, error)
		FetchByID(ctx context.Context, id uint) (*entity.SearchArea, error)
		FetchByName(ctx context.Context, name string) (*entity.SearchArea, error)
		FetchAll(ctx context.Context) ([]*entity.SearchArea, error)
	}
)

type SearchAreaImpl struct {
	r    SearchAreaRepo
	l    *slog.Logger
	mqtt mqtt.Client
}

func NewSearchAreaImpl(r SearchAreaRepo, l *slog.Logger, mqtt mqtt.Client) SearchAreaSvc {
	return &SearchAreaImpl{
		r:    r,
		l:    l,
		mqtt: mqtt,
	}
}

func (s *SearchAreaImpl) SaveArea(ctx context.Context, area *entity.SearchArea) (*entity.SearchArea, error) {
	area.CalcCenter()
	area, err := s.r.Save(ctx, area)
	if err != nil {
		s.l.Error("SaveArea Error: ", slog.Any("error", err))
		return nil, err
	}
	return area, nil
}

func (s *SearchAreaImpl) FetchArea(ctx context.Context, params *api.AreaFetchParams) (*entity.SearchArea, error) {
	var area *entity.SearchArea
	var multiErr error
	if params.ID != 0 {
		a, err := s.r.FetchByID(ctx, params.ID)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
		area = a
	} else if params.Name != "" {
		a, err := s.r.FetchByName(ctx, params.Name)
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

func (s *SearchAreaImpl) FetchList(ctx context.Context) ([]*entity.SearchArea, error) {
	areas, err := s.r.FetchAll(ctx)
	if err != nil {
		s.l.Error("FetchList Error: ", slog.Any("error", err))
		return nil, err
	}
	return areas, nil
}
