package v1

import (
	"context"
	"log/slog"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/vo"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

type SearchAreaRouter struct {
	svc service.AreaSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func NewSearchAreaRouter(handler fiber.Router, svc service.AreaSvc, eb EventBus.Bus, l *slog.Logger) {
	r := &SearchAreaRouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/areas")
	{
		h.Get("/list", r.getAllAreas)
		h.Get("/", r.getArea)
		h.Post("/", r.createArea)
		h.Put("/:id", r.updateArea)
		h.Delete("/:id", r.deleteArea)
	}
}

type areaResult struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CenterLat   float64 `json:"center_lat"`
	CenterLng   float64 `json:"center_lng"`
	Points      []struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"points"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (r *SearchAreaRouter) toResult(area *entity.Area) *areaResult {
	var points []struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	for _, point := range area.Points {
		points = append(points, struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		}{
			Lat: point.Lat,
			Lng: point.Lng,
		})
	}
	var e areaResult
	_ = copier.Copy(&e, area)
	e.Points = points
	e.CreatedAt = area.CreatedAt.Format("2006-01-02 15:04:05")
	e.UpdatedAt = area.UpdatedAt.Format("2006-01-02 15:04:05")
	return &e
}

// areaItemResult 搜索区域列表结果
type areaItemResult struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CenterLat   float64 `json:"center_lat"`
	CenterLng   float64 `json:"center_lng"`
	Points      []struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"points"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (r *SearchAreaRouter) toListResult(area []*entity.Area) []areaItemResult {
	var items []areaItemResult
	for _, a := range area {
		var e areaItemResult
		_ = copier.Copy(&e, a)
		// 复制点列表
		var points []struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		}
		for _, p := range a.Points {
			points = append(points, struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			}{
				Lat: p.Lat,
				Lng: p.Lng,
			})
		}
		e.Points = points
		e.CreatedAt = a.CreatedAt.Format("2006-01-02 15:04:05")
		e.UpdatedAt = a.UpdatedAt.Format("2006-01-02 15:04:05")
		items = append(items, e)
	}
	return items
}

// getAllAreas 获取所有搜索区域
func (r *SearchAreaRouter) getAllAreas(c *fiber.Ctx) error {
	ctx := context.Background()
	var params struct {
		Name           string `query:"name"`
		CreatedAtBegin string `query:"created_at_begin"`
		CreatedAtEnd   string `query:"created_at_end"`
		Page           int    `query:"page"`
		PageSize       int    `query:"page_size"`
	}
	if err := c.QueryParser(&params); err != nil {
		r.l.Error("getAllAreas Error: ", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(Fail(InvalidParams))
	}
	// 调用服务层获取搜索区域列表
	r.l.Info("getAllAreas", slog.Any("params", params))
	areas, total, err := r.svc.FetchAll(ctx, params.Name, params.CreatedAtBegin, params.CreatedAtEnd, params.Page, params.PageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Fail(InternalError))
	}
	return c.JSON(Success(fiber.Map{
		"total": total,
		"items": r.toListResult(areas),
	}))
}

// getArea 获取搜索区域
func (r *SearchAreaRouter) getArea(c *fiber.Ctx) error {
	var params struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	if err := c.QueryParser(&params); err != nil {
		r.l.Error("getArea Error: ", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(Fail(InvalidParams))
	}
	area, err := r.svc.FetchArea(context.Background(), params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Fail(InternalError))
	}
	return c.JSON(Success(r.toResult(area)))
}

// createArea 创建搜索区域
func (r *SearchAreaRouter) createArea(c *fiber.Ctx) error {
	var body struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Points      []vo.GeoPoint `json:"points"`
	}
	ctx := context.Background()
	if err := c.BodyParser(&body); err != nil { // 修改了这里，添加 & 符号传递指针{
		r.l.Error("createArea Error: ", slog.Any("error", err))
		return c.JSON(FailWithMsg(err.Error()))
	}
	area, err := r.svc.CreateArea(ctx, body.Name, body.Description, body.Points)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(area))
}

func (r *SearchAreaRouter) updateArea(c *fiber.Ctx) error {
	// 创建 ctx 并获取参数
	ctx := context.Background()
	id, _ := c.ParamsInt("id")
	var body struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		Points      []vo.GeoPoint `json:"points"`
	}
	if err := c.BodyParser(&body); err != nil { // 这里同样需要修改为指针
		return c.JSON(Fail(InvalidParams))
	}
	// 调用服务层更新区域
	area, err := r.svc.UpdateArea(ctx, uint(id), body.Name, body.Description, body.Points)
	if err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(area))
}

func (r *SearchAreaRouter) deleteArea(c *fiber.Ctx) error {
	ctx := context.Background()
	id, _ := c.ParamsInt("id")

	if err := r.svc.Repo().DeleteByID(ctx, uint(id)); err != nil {
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(nil))
}
