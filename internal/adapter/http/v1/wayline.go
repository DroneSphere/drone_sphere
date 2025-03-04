package v1

import (
	"context"
	"fmt"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/model/vo"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"log/slog"
	"os"
)

type WaylineRouter struct {
	svc service.WaylineSvc
	l   *slog.Logger
}

func newWaylineRouter(handler fiber.Router, svc service.WaylineSvc, l *slog.Logger) *WaylineRouter {
	r := &WaylineRouter{
		svc: svc,
		l:   l,
	}
	h := handler.Group("/wayline")
	{
		h.Post("/", r.Create)
		h.Get("/download", r.Download)
		h.Get("/", r.ListAll)
	}
	return r
}

// Create 创建航线
//
//	@Router			/wayline [post]
//	@Summary		创建航线
//	@Description	根据给出的点序列和无人机SN、高度生成航线
//	@Tags			wayline
//	@Accept			json
//	@Produce		json
//	@Param			req	body		v1.CreateWaylineRequest	true	"请求体"
//	@Success		200	{object}	v1.Response{data=nil}	"成功"
func (r *WaylineRouter) Create(c *fiber.Ctx) error {
	var req api.CreateWaylineRequest
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}
	var points []vo.GeoPoint
	for i, p := range req.Points {
		t := vo.GeoPoint{
			Index: i,
			Lat:   p.Lat,
			Lng:   p.Lng,
		}
		points = append(points, t)
	}
	_, err := r.svc.Create(c.Context(), points, req.DroneSN, req.Height)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	return c.JSON(Success(nil))
}

// Download 下载航线文件
//
//	@Router			/wayline/download [get]
//	@Summary		下载航线文件
//	@Description	根据给出的key下载航线文件
//	@Tags			wayline
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=nil}	"成功"
func (r *WaylineRouter) Download(c *fiber.Ctx) error {
	ctx := context.Background()
	fileName := "5048533d-a929-4935-b727-fad6a16753f9.kmz"
	path, err := r.svc.Download(ctx, fileName)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	defer func() {
		if err := os.Remove(path); err != nil {
			r.l.Error("Remove Error: ", slog.Any("error", err))
		}

	}()
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	return c.SendFile(path, false)
}

// ListAll 列出所有航线
//
//	@Router			/wayline [get]
//	@Summary		列出所有航线
//	@Description	列出所有航线
//	@Tags			wayline
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=[]v1.WaylineItemResult}	"成功"
func (r *WaylineRouter) ListAll(c *fiber.Ctx) error {
	ctx := context.Background()
	list, err := r.svc.FetchAll(ctx)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	var res []api.WaylineItemResult
	for _, l := range list {
		var e api.WaylineItemResult
		if err := copier.Copy(&e, &l); err != nil {
			return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
		}
		res = append(res, e)
	}
	return c.JSON(Success(res))
}
