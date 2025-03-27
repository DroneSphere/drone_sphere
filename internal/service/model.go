package service

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
)

type (
	ModelSvc interface {
		Repo() ModelRepo
	}

	ModelRepo interface {
		SelectAllDroneVariation(ctx context.Context, query map[string]string) ([]po.DroneVariation, error)
		SelectAllDroneModel(ctx context.Context) ([]entity.DroneModel, error)
		SelectAllGimbals(ctx context.Context) ([]po.GimbalModel, error)
		SelectGimbalsByIDs(ctx context.Context, ids []uint) ([]po.GimbalModel, error)
		SelectAllGatewayModel(ctx context.Context) ([]po.GatewayModel, error)
		SelectAllPayloadModel(ctx context.Context) ([]po.PayloadModel, error)
	}
)

type ModelImpl struct {
	repo ModelRepo
	l    *slog.Logger
}

func NewModelImpl(repo ModelRepo, l *slog.Logger) ModelSvc {
	return &ModelImpl{
		repo: repo,
		l:    l,
	}
}

func (m *ModelImpl) Repo() ModelRepo {
	return m.repo
}
