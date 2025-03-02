package v1

import (
	"context"
	"log/slog"

	"github.com/asaskevich/EventBus"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

type SearchAreaRouter struct {
	svc service.SearchAreaSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func newSearchAreaRouter(handler fiber.Router, svc service.SearchAreaSvc, eb EventBus.Bus, l *slog.Logger) {
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
		h.Delete("/:id", r.deleteArea)
	}
}

func (r *SearchAreaRouter) toItemResult(area *entity.SearchArea) *api.AreaResult {
	var points []api.PointResult
	for _, point := range area.Points {
		points = append(points, api.PointResult{
			Index: point.Index,
			Lat:   point.Lat,
			Lng:   point.Lng,
		})
	}
	return &api.AreaResult{
		ID:     area.ID,
		Name:   area.Name,
		Points: points,
	}
}

func (r *SearchAreaRouter) toListResult(area []*entity.SearchArea) []api.AreaItemResult {
	var items []api.AreaItemResult
	for _, a := range area {
		var e api.AreaItemResult
		if err := copier.Copy(&e, a); err != nil {
			r.l.Error("Copy Error: ", slog.Any("error", err))
			continue
		}
		// 复制点列表
		var points []api.PointResult
		for _, p := range a.Points {
			points = append(points, api.PointResult{
				Index: p.Index,
				Lat:   p.Lat,
				Lng:   p.Lng,
			})
		}
		items = append(items, e)
	}
	return items
}

// getAllAreas 列出所有搜索区域
//
//	@Router			/areas/list	[get]
//	@Summary		列出所有搜索区域
//	@Description	列出所有搜索区域，不返回每个区域的点列表，仅返回中心点位置
//	@Tags			area
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=[]v1.AreaItemResult}	"成功"
func (r *SearchAreaRouter) getAllAreas(c *fiber.Ctx) error {
	ctx := context.Background()
	areas, err := r.svc.FetchList(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Fail(InternalError))
	}
	return c.JSON(Success(r.toListResult(areas)))
}

// getArea 获取搜索区域
//
//	@Router			/areas	[get]
//	@Summary		获取搜索区域的详细信息
//	@Description	获取搜索区域的详细信息，包括区域的点列表
//	@Tags			area
//	@Accept			json
//	@Produce		json
//	@Param			id		query		int								true	"区域ID"
//	@Param			name	query		string							false	"区域名称"
//	@Success		200		{object}	v1.Response{data=v1.AreaResult}	"成功"
func (r *SearchAreaRouter) getArea(c *fiber.Ctx) error {
	var param api.AreaFetchParams
	if err := c.QueryParser(&param); err != nil {
		r.l.Error("getArea Error: ", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(Fail(InvalidParams))
	}
	area, err := r.svc.FetchArea(context.Background(), &param)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Fail(InternalError))
	}
	return c.JSON(Success(r.toItemResult(area)))
}

var duplicateAreaFailed = ErrorBody{Code: 40000, Msg: "已存在同名区域"}

// createArea 创建搜索区域
//
//	@Router			/areas	[post]
//	@Summary		创建搜索区域
//	@Description	创建搜索区域
//	@Tags			area
//	@Accept			json
//	@Produce		json
//	@Param			req	body		v1.CreateAreaRequest			true	"请求体"
//	@Success		200	{object}	v1.Response{data=v1.AreaResult}	"成功"
//	@Failure		400	{object}	v1.ErrorBody					"参数错误"
func (r *SearchAreaRouter) createArea(c *fiber.Ctx) error {
	ctx := context.Background()
	req := new(api.CreateAreaRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(InvalidParams))
	}
	area, err := r.svc.SaveArea(ctx, req.ToEntity())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(duplicateAreaFailed))
	}
	return c.JSON(Success(area))
}

// deleteArea 删除搜索区域
//
//	@Router			/areas/:id	[delete]
//	@Summary		删除搜索区域
//	@Description	删除搜索区域
//	@Tags			area
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int		true	"区域ID"
//	@Success		200	{object}	v1.Response	"成功"
func (r *SearchAreaRouter) deleteArea(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(InvalidParams))
	}
	if err := r.svc.DeleteByID(context.Background(), uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Fail(InternalError))
	}
	return c.JSON(Success(nil))
}
