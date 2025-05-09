package repo

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/po"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type WaylineGormRepo struct {
	tx     *gorm.DB
	s3     *minio.Client
	l      *slog.Logger
	bucket string
}

func NewWaylineGormRepo(db *gorm.DB, s3 *minio.Client, l *slog.Logger) *WaylineGormRepo {
	// if err := db.AutoMigrate(&po.Wayline{}); err != nil {
	// 	l.Error("Auto Migration Error: ", slog.Any("error", err))
	// 	panic(err)
	// }
	return &WaylineGormRepo{
		tx:     db,
		s3:     s3,
		l:      l,
		bucket: "kmz",
	}
}

func (w *WaylineGormRepo) SelectAll(ctx context.Context) ([]po.Wayline, error) {
	var waylines []po.Wayline
	if err := w.tx.WithContext(ctx).Find(&waylines).Error; err != nil {
		return nil, err
	}
	return waylines, nil
}

// SelectByID 根据航线ID获取航线详情
func (w *WaylineGormRepo) SelectByID(ctx context.Context, id string) (*po.Wayline, error) {
	var wayline po.Wayline
	if err := w.tx.WithContext(ctx).Where("uuid = ?", id).First(&wayline).Error; err != nil {
		w.l.Error("查询航线失败", slog.Any("id", id), slog.Any("err", err))
		return nil, err
	}
	return &wayline, nil
}

func (w *WaylineGormRepo) SelectByJobIDAndDroneSN(ctx context.Context, jobID uint, droneSN string) (*po.Wayline, error) {
	var waylines []po.Wayline
	if err := w.tx.WithContext(ctx).Where("state=0 AND job_id = ? AND drone_sn = ?", jobID, droneSN).Find(&waylines).Error; err != nil {
		w.l.Error("查询航线失败", slog.Any("job_id", jobID), slog.Any("drone_sn", droneSN), slog.Any("err", err))
		return nil, err
	}
	if len(waylines) == 0 || len(waylines) > 1 {
		w.l.Error("查询航线失败", slog.Any("job_id", jobID), slog.Any("drone_sn", droneSN), slog.Any("err", "航线数量不正确"))
		return nil, gorm.ErrRecordNotFound
	}

	return &waylines[0], nil
}
