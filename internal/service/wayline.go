package service

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/po"
)

type (
	WaylineSvc interface {
		Repo() WaylineRepo
	}

	WaylineRepo interface {
		SelectAll(ctx context.Context) ([]po.Wayline, error)
	}
)

type WaylineImpl struct {
	r WaylineRepo
	l *slog.Logger
}

func NewWaylineImpl(r WaylineRepo, l *slog.Logger) WaylineSvc {
	return &WaylineImpl{
		r: r,
		l: l,
	}
}

func (w *WaylineImpl) Repo() WaylineRepo {
	return w.r
}
