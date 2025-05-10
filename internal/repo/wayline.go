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

func (w *WaylineGormRepo) SelectByJobIDAndDroneKey(ctx context.Context, jobID uint, droneKey string) (*po.Wayline, error) {
	var waylines []po.Wayline
	if err := w.tx.WithContext(ctx).Where("state = 0 AND job_id = ? AND job_drone_key = ?", jobID, droneKey).Find(&waylines).Error; err != nil {
		w.l.Error("查询航线失败", slog.Any("job_id", jobID), slog.Any("drone_key", droneKey), slog.Any("err", err))
		return nil, err
	}
	if len(waylines) == 0 || len(waylines) > 1 {
		w.l.Error("查询航线失败", slog.Any("job_id", jobID), slog.Any("drone_key", droneKey), slog.Any("err", "航线数量不正确"))
		return nil, gorm.ErrRecordNotFound
	}
	return &waylines[0], nil
}

func (w *WaylineGormRepo) DeleteByID(ctx context.Context, id uint) error {
	// 查询要删除的航线信息，同时进行软删除（将状态改为1）
	var wayline po.Wayline
	result := w.tx.WithContext(ctx).
		Model(&po.Wayline{}).
		Where("wayline_id = ? AND state = 0", id).
		First(&wayline).
		Update("state", -1)

	// 检查是否找到并更新了记录
	if result.Error != nil {
		w.l.Error("删除航线失败", slog.Any("id", id), slog.Any("err", result.Error))
		return result.Error
	}

	if result.RowsAffected == 0 {
		w.l.Error("未找到可删除的航线", slog.Any("id", id))
		return gorm.ErrRecordNotFound
	}

	// 如果有S3存储的文件，删除对应的S3对象
	if wayline.S3Key != "" {
		if err := w.s3.RemoveObject(ctx, w.bucket, wayline.S3Key, minio.RemoveObjectOptions{}); err != nil {
			w.l.Error("删除S3航线文件失败",
				slog.Any("id", id),
				slog.Any("key", wayline.S3Key),
				slog.Any("err", err))
			// 注意：即使S3删除失败，数据库删除仍然完成，所以这里不返回错误
			// 如果需要严格保证一致性，可以考虑事务或回滚机制
		} else {
			w.l.Info("已删除S3航线文件",
				slog.Any("id", id),
				slog.Any("key", wayline.S3Key))
		}
	}

	w.l.Info("航线已删除", slog.Any("id", id))
	return nil
}
