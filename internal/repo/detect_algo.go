package repo

import (
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"log/slog"
)

type DetectAlgoGormRepo struct {
	tx *gorm.DB
	l  *slog.Logger
}

func NewDetectAlgoGormRepo(tx *gorm.DB, l *slog.Logger) *DetectAlgoGormRepo {
	//if err := tx.AutoMigrate(&po.DetectAlgo{}); err != nil {
	//	l.Error("Failed to auto migrate DetectAlgo", slog.Any("err", err))
	//	panic(err)
	//}
	return &DetectAlgoGormRepo{tx: tx, l: l}
}

func (r *DetectAlgoGormRepo) toPO(algo entity.DetectAlgo) (po.DetectAlgo, error) {
	var saved po.DetectAlgo
	if err := copier.Copy(&saved, &algo); err != nil {
		return po.DetectAlgo{}, err
	}
	return saved, nil
}

func (r *DetectAlgoGormRepo) toEntity(saved po.DetectAlgo) (entity.DetectAlgo, error) {
	var algo entity.DetectAlgo
	if err := copier.Copy(&algo, &saved); err != nil {
		return entity.DetectAlgo{}, err
	}
	return algo, nil
}

func (r *DetectAlgoGormRepo) FetchAll() (algos []entity.DetectAlgo, err error) {
	var saved []po.DetectAlgo
	if err := r.tx.Find(&saved).Error; err != nil {
		return nil, err
	}
	for _, s := range saved {
		algo, err := r.toEntity(s)
		if err != nil {
			return nil, err
		}
		algos = append(algos, algo)
	}
	return algos, nil
}

func (r *DetectAlgoGormRepo) FetchByID(id uint) (algo entity.DetectAlgo, err error) {
	var saved po.DetectAlgo
	if err := r.tx.First(&saved, id).Error; err != nil {
		return entity.DetectAlgo{}, err
	}
	return r.toEntity(saved)
}

func (r *DetectAlgoGormRepo) Save(algo entity.DetectAlgo) (err error) {
	saved, err := r.toPO(algo)
	if err != nil {
		return err
	}
	if err := r.tx.Save(&saved).Error; err != nil {
		return err
	}
	return nil
}

func (r *DetectAlgoGormRepo) DeleteByID(id uint) error {
	if err := r.tx.Where("id = ?", id).Delete(&po.DetectAlgo{}).Error; err != nil {
		return err
	}
	return nil
}
