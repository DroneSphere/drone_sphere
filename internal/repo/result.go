package repo

import (
	"context"
	"log/slog"

	"github.com/dronesphere/internal/model/dto"
	"github.com/dronesphere/internal/model/po"
	"gorm.io/gorm"
)

type ResultDefaultRepo struct {
	tx *gorm.DB
	l  *slog.Logger
}

func NewResultDefaultRepo(db *gorm.DB, l *slog.Logger) *ResultDefaultRepo {
	return &ResultDefaultRepo{
		tx: db,
		l:  l,
	}
}

func (r *ResultDefaultRepo) Create(ctx context.Context, result *po.Result) error {
	if err := r.tx.WithContext(ctx).Create(result).Error; err != nil {
		r.l.Error("创建检测结果失败", slog.Any("err", err))
		return err
	}
	return nil
}

func (r *ResultDefaultRepo) GetByID(ctx context.Context, id uint) (*po.Result, error) {
	var result po.Result
	if err := r.tx.WithContext(ctx).First(&result, id).Error; err != nil {
		r.l.Error("获取检测结果失败", slog.Any("err", err))
		return nil, err
	}
	return &result, nil
}

func (r *ResultDefaultRepo) List(ctx context.Context, query dto.ResultQuery) ([]po.Result, int64, error) {
	var results []po.Result
	var total int64

	tx := r.tx.WithContext(ctx).Model(&po.Result{}).Where("tb_results.state = ?", 0)

	if query.ObjectType != 0 {
		tx = tx.Where("tb_results.object_type = ?", query.ObjectType)
	}

	// 如果提供了JobName，通过关联查询筛选
	if query.JobName != "" {
		tx = tx.Joins("JOIN tb_jobs j ON tb_results.job_id = j.job_id").
			Where("j.job_name LIKE ?", "%"+query.JobName+"%").
			Where("j.state = ?", 0) // 确保只筛选有效的任务
	}

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		r.l.Error("获取检测结果总数失败", slog.Any("err", err))
		return nil, 0, err
	}

	// 分页查询
	if err := tx.Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Order("created_time DESC").
		Find(&results).Error; err != nil {
		r.l.Error("获取检测结果列表失败", slog.Any("err", err))
		return nil, 0, err
	}

	return results, total, nil
}

func (r *ResultDefaultRepo) GetJobOptions(ctx context.Context) ([]dto.JobOption, error) {
	var options []dto.JobOption
	if err := r.tx.WithContext(ctx).
		Model(&po.Job{}).
		Where("state = ?", 0).
		Select("job_id as id, job_name as name").
		Find(&options).Error; err != nil {
		r.l.Error("获取任务选项失败", slog.Any("err", err))
		return nil, err
	}
	return options, nil
}

// DeleteByID 删除检测结果
func (r *ResultDefaultRepo) DeleteByID(ctx context.Context, id uint) error {
	if err := r.tx.WithContext(ctx).Where("result_id = ?", id).Delete(&po.Result{}).Error; err != nil {
		r.l.Error("删除检测结果失败", slog.Any("err", err))
		return err
	}
	// 删除成功
	r.l.Info("删除检测结果成功", slog.Any("id", id))

	return nil
}
