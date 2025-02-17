package service

import (
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/vo"
	"log/slog"
)

type (
	DetectAlgoSvc interface {
		GetAll() ([]entity.DetectAlgo, error)
		GetByID(id uint) (entity.DetectAlgo, error)
		Create(algo entity.DetectAlgo) (entity.DetectAlgo, error)
		UpdateClasses(id uint, classes []vo.DetectClass) (entity.DetectAlgo, error)
		DeleteByID(id uint) error
	}
	DetectAlgoRepo interface {
		FetchAll() ([]entity.DetectAlgo, error)
		FetchByID(id uint) (entity.DetectAlgo, error)
		Save(algo entity.DetectAlgo) error
		DeleteByID(id uint) error
	}
)

type DetectAlgoImpl struct {
	r DetectAlgoRepo
	l *slog.Logger
}

func NewDetectAlgoImpl(r DetectAlgoRepo, l *slog.Logger) DetectAlgoSvc {
	return &DetectAlgoImpl{
		r: r,
		l: l,
	}
}

func (s *DetectAlgoImpl) GetAll() ([]entity.DetectAlgo, error) {
	algos, err := s.r.FetchAll()
	if err != nil {
		return nil, err
	}
	return algos, nil
}

func (s *DetectAlgoImpl) GetByID(id uint) (entity.DetectAlgo, error) {
	algo, err := s.r.FetchByID(id)
	if err != nil {
		return entity.DetectAlgo{}, err
	}
	return algo, nil
}

func (s *DetectAlgoImpl) Create(algo entity.DetectAlgo) (entity.DetectAlgo, error) {
	err := s.r.Save(algo)
	if err != nil {
		return entity.DetectAlgo{}, err
	}
	return algo, nil
}

func (s *DetectAlgoImpl) UpdateClasses(id uint, classes []vo.DetectClass) (entity.DetectAlgo, error) {
	algo, err := s.r.FetchByID(id)
	if err != nil {
		return entity.DetectAlgo{}, err
	}
	algo.Classes = classes
	err = s.r.Save(algo)
	if err != nil {
		return entity.DetectAlgo{}, err
	}
	return algo, nil
}

func (s *DetectAlgoImpl) DeleteByID(id uint) error {
	err := s.r.DeleteByID(id)
	if err != nil {
		return err
	}
	return nil
}
