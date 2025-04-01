package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/model/vo"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jinzhu/copier"
)

type (
	AreaSvc interface {
		Repo() AreaRepo
		CreateArea(ctx context.Context, name, description string, points []vo.GeoPoint) (*entity.Area, error)
		UpdateArea(ctx context.Context, id uint, name, description string, points []vo.GeoPoint) (*entity.Area, error)
		FetchArea(ctx context.Context, params struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		}) (*entity.Area, error)
		FetchAll(ctx context.Context, name string) ([]*entity.Area, error)
	}

	// 修改 AreaRepo 接口，返回 po 对象而不是 entity 对象
	AreaRepo interface {
		Save(ctx context.Context, area *po.Area) error
		SelectByID(ctx context.Context, id uint) (*po.Area, error)
		SelectByName(ctx context.Context, name string) (*po.Area, error)
		FetchAll(ctx context.Context, name string) ([]*po.Area, error)
		DeleteByID(ctx context.Context, id uint) error
	}
)

type AreaImpl struct {
	r    AreaRepo
	l    *slog.Logger
	mqtt mqtt.Client
}

func NewAreaImpl(r AreaRepo, l *slog.Logger, mqtt mqtt.Client) AreaSvc {
	return &AreaImpl{
		r:    r,
		l:    l,
		mqtt: mqtt,
	}
}

func (s *AreaImpl) Repo() AreaRepo {
	return s.r
}

// toEntity 将 po.Area 转换为 entity.Area
// 在服务层中进行包装，避免循环引用
func (s *AreaImpl) toEntity(p *po.Area) *entity.Area {
	if p == nil {
		return nil
	}

	var points []vo.GeoPoint
	for _, point := range p.Points {
		var p vo.GeoPoint
		if err := copier.Copy(&p, point); err != nil {
			s.l.Error("复制点数据失败", slog.Any("error", err))
			return nil
		}
		points = append(points, p)
	}

	var area entity.Area
	if err := copier.Copy(&area, p); err != nil {
		s.l.Error("复制区域数据失败", slog.Any("error", err))
		return nil
	}
	area.Points = points
	area.CreatedAt = p.CreatedTime
	area.UpdatedAt = p.UpdatedTime
	return &area
}

// toPO 将 entity.Area 转换为 po.Area
// 在服务层中进行包装，避免循环引用
func (s *AreaImpl) toPO(e *entity.Area) *po.Area {
	if e == nil {
		return nil
	}

	p := &po.Area{}
	if err := copier.Copy(p, e); err != nil {
		s.l.Error("复制区域数据失败", slog.Any("error", err))
		return nil
	}
	return p
}

// CreateArea 创建一个新的区域
func (s *AreaImpl) CreateArea(ctx context.Context, name, description string, points []vo.GeoPoint) (*entity.Area, error) {
	// 检查名称是否已存在
	existArea, err := s.r.SelectByName(ctx, name)
	if err == nil && existArea != nil {
		return nil, errors.New("区域名称已存在")
	}

	// 创建新的区域对象
	area := &entity.Area{
		Name:        name,
		Description: description,
		Points:      points,
	}
	// 计算区域中心点
	if err := area.CalcCenter(); err != nil {
		s.l.Error("计算区域中心点失败", slog.Any("error", err))
		return nil, err
	}

	// 转换为 PO 对象并保存
	poArea := s.toPO(area)
	if err := s.r.Save(ctx, poArea); err != nil {
		return nil, err
	}

	// 重新获取保存后的结果并返回 entity 对象
	savedArea, err := s.r.SelectByID(ctx, poArea.ID)
	if err != nil {
		return nil, err
	}

	return s.toEntity(savedArea), nil
}

// UpdateArea 更新区域信息
func (s *AreaImpl) UpdateArea(ctx context.Context, id uint, name, description string, points []vo.GeoPoint) (*entity.Area, error) {
	// 检查区域是否存在
	existArea, err := s.r.SelectByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 如果更改了名称，检查新名称是否可用
	if existArea.Name != name {
		sameNameArea, err := s.r.SelectByName(ctx, name)
		if err == nil && sameNameArea != nil && sameNameArea.ID != id {
			return nil, errors.New("区域名称已存在")
		}
	}

	// 更新区域信息
	existArea.Name = name
	existArea.Description = description

	// 转换并保存点数据
	var poPoints []vo.GeoPoint
	for _, p := range points {
		var point vo.GeoPoint
		if err := copier.Copy(&point, p); err != nil {
			s.l.Error("复制点数据失败", slog.Any("error", err))
			return nil, err
		}
		poPoints = append(poPoints, point)
	}
	existArea.Points = poPoints

	// 保存更新后的区域
	if err := s.r.Save(ctx, existArea); err != nil {
		return nil, err
	}

	// 返回更新后的 entity 对象
	return s.toEntity(existArea), nil
}

// FetchArea 获取单个区域详情
func (s *AreaImpl) FetchArea(ctx context.Context, params struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}) (*entity.Area, error) {
	var area *po.Area
	var err error

	if params.ID > 0 {
		area, err = s.r.SelectByID(ctx, params.ID)
	} else if params.Name != "" {
		area, err = s.r.SelectByName(ctx, params.Name)
	} else {
		return nil, errors.New("缺少查询条件")
	}

	if err != nil {
		return nil, err
	}

	return s.toEntity(area), nil
}

// FetchAll 获取所有区域列表
func (s *AreaImpl) FetchAll(ctx context.Context, name string) ([]*entity.Area, error) {
	// 从仓库层获取 po 对象列表
	areas, err := s.r.FetchAll(ctx, name)
	if err != nil {
		return nil, err
	}

	// 将 po 对象列表转换为 entity 对象列表
	var result []*entity.Area
	for _, area := range areas {
		entityArea := s.toEntity(area)
		if entityArea != nil {
			result = append(result, entityArea)
		}
	}

	return result, nil
}
