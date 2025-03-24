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
