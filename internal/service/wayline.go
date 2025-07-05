package service

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
)

type (
	WaylineSvc interface {
		Repo() WaylineRepo
		GetWaylineURL(ctx context.Context, workspaceID, waylineID string) (string, error)
		FetchWaylineByJobIDAndDroneSN(ctx context.Context, jobID uint, droneSN string) (*entity.Wayline, error)
		FetchWaylineByJobIDAndDroneKey(ctx context.Context, jobID uint, droneKey string) (*entity.Wayline, error)
	}

	WaylineRepo interface {
		SelectAll(ctx context.Context) ([]po.Wayline, error)
		SelectByID(ctx context.Context, id string) (*po.Wayline, error)
		SelectByJobIDAndDroneSN(ctx context.Context, jobID uint, droneSN string) (*po.Wayline, error)
		SelectByJobIDAndDroneKey(ctx context.Context, jobID uint, droneKey string) (*po.Wayline, error)
		DeleteByID(ctx context.Context, id uint) error
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

// GetWaylineURL 获取航线文件的URL
func (w *WaylineImpl) GetWaylineURL(ctx context.Context, workspaceID, waylineID string) (string, error) {
	// 查询航线信息
	wayline, err := w.r.SelectByID(ctx, waylineID)
	if err != nil {
		w.l.Error("获取航线信息失败", slog.Any("waylineID", waylineID), slog.Any("err", err))
		return "", err
	}

	// 拼接URL
	baseURL := "http://47.245.40.222:9000/kmz"
	url := baseURL + "/" + wayline.S3Key

	w.l.Info("获取航线URL成功", slog.Any("waylineID", waylineID), slog.Any("url", url))
	return url, nil
}

func (w *WaylineImpl) FetchWaylineByJobIDAndDroneSN(ctx context.Context, jobID uint, droneSN string) (*entity.Wayline, error) {
	wayline, err := w.r.SelectByJobIDAndDroneSN(ctx, jobID, droneSN)
	if err != nil {
		w.l.Error("获取航线失败", slog.Any("jobID", jobID), slog.Any("droneSN", droneSN), slog.Any("err", err))
		return nil, err
	}
	return &entity.Wayline{
		Wayline: *wayline,
		Url:     "http://47.245.40.222:9000/kmz/" + wayline.S3Key,
	}, nil
}

func (w *WaylineImpl) FetchWaylineByJobIDAndDroneKey(ctx context.Context, jobID uint, droneKey string) (*entity.Wayline, error) {
	wayline, err := w.r.SelectByJobIDAndDroneKey(ctx, jobID, droneKey)
	if err != nil {
		w.l.Error("获取航线失败", slog.Any("jobID", jobID), slog.Any("droneKey", droneKey), slog.Any("err", err))
		return nil, err
	}
	return &entity.Wayline{
		Wayline: *wayline,
		Url:     "http://47.245.40.222:9000/kmz/" + wayline.S3Key,
	}, nil
}
