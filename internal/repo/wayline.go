package repo

import (
	"context"
	"fmt"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/jinzhu/copier"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"log/slog"
)

type WaylineGormRepo struct {
	tx     *gorm.DB
	s3     *minio.Client
	l      *slog.Logger
	bucket string
}

func NewWaylineGormRepo(db *gorm.DB, s3 *minio.Client, l *slog.Logger) *WaylineGormRepo {
	//if err := db.AutoMigrate(&po.Wayline{}); err != nil {
	//	l.Error("Auto Migration Error: ", slog.Any("error", err))
	//	panic(err)
	//}
	return &WaylineGormRepo{
		tx:     db,
		s3:     s3,
		l:      l,
		bucket: "kmz",
	}
}

func (r *WaylineGormRepo) FetchKMZByKey(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf("./%s", key)
	err := r.s3.FGetObject(ctx, r.bucket, key, path, minio.GetObjectOptions{})
	if err != nil {
		r.l.Error("FGetObject Error: ", slog.Any("error", err))
		return "", err
	}
	return path, nil
}

func (r *WaylineGormRepo) SaveToS3(ctx context.Context, path, key string) (string, error) {
	contentType := "application/zip"

	info, err := r.s3.FPutObject(ctx, r.bucket, key, path, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		r.l.Error("FPutObject Error: ", slog.Any("error", err))
		return "", err
	}
	r.l.Info("FPutObject Info: ", slog.Any("info", info))
	return key, nil
}

func (r *WaylineGormRepo) Save(ctx context.Context, wayline entity.Wayline) (entity.Wayline, error) {
	// 数据装配
	p := &po.Wayline{}
	if err := copier.Copy(p, wayline); err != nil {
		r.l.Error("Copy Error: ", slog.Any("error", err))
		return entity.Wayline{}, err
	}
	p.Points = wayline.Points
	// 保存到数据库
	if err := r.tx.Create(&p).Error; err != nil {
		r.l.Error("Create Error: ", slog.Any("error", err))
		return entity.Wayline{}, err
	}
	// 装配返回值
	wayline.ID = p.ID
	return wayline, nil
}

func (r *WaylineGormRepo) FetchAll(ctx context.Context) ([]entity.Wayline, error) {
	var pos []po.Wayline
	if err := r.tx.Find(&pos).Error; err != nil {
		r.l.Error("Find Error: ", slog.Any("error", err))
		return nil, err
	}
	r.l.Info("Find Result: ", slog.Any("pos", pos))
	var res []entity.Wayline
	for _, p := range pos {
		t := entity.Wayline{}
		if err := copier.Copy(&t, &p); err != nil {
			r.l.Error("Copy Error: ", slog.Any("error", err))
			return nil, err
		}
		t.Points = p.Points
		res = append(res, t)
	}
	return res, nil
}
