package v1

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/asaskevich/EventBus"
	"github.com/bytedance/sonic"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/repo"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

type DroneRouter struct {
	svc service.DroneSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func newDroneRouter(handler fiber.Router, svc service.DroneSvc, eb EventBus.Bus, l *slog.Logger) {
	r := &DroneRouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/drone")
	{
		h.Get("/list", r.list)
		h.Put("/:sn", r.update)
		h.Get("/sn/:sn", r.getBySN)
		h.Get("/state/sse", r.pushState)
	}
}

// list 列出所有无人机
//
//	@Router			/drone/list [get]
//	@Summary		列出所有无人机
//	@Description	列出所有绑定的无人机，包含不在线的无人机
//	@Tags			drone
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=[]v1.DroneItemResult}	"成功"
func (r *DroneRouter) list(c *fiber.Ctx) error {
	ctx := context.Background()
	drones, err := r.svc.Repo().SelectAll(ctx)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	var res []api.DroneItemResult
	for _, d := range drones {
		var e api.DroneItemResult
		if err := copier.Copy(&e, &d); err != nil {
			return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
		}
		// 检查是否在线
		e.Status = d.StatusText()
		e.CreatedAt = d.CreatedAt.Format("2006年01月02日 15:04:05")
		e.LastOnlineAt = d.UpdatedAt.Format("2006年01月02日 15:04:05")
		res = append(res, e)
	}

	return c.JSON(Success(res))
}

// getBySN  根据序列号获取无人机信息
//
//	@Router			/drone/sn/:sn [get]
//	@Summary		根据序列号获取无人机信息
//	@Description	根据序列号获取无人机信息
//	@Tags			drone
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=v1.DroneDetailResult}	"成功"
func (r *DroneRouter) getBySN(c *fiber.Ctx) error {
	sn := c.Params("sn")
	ctx := context.Background()
	drone, err := r.svc.Repo().SelectBySN(ctx, sn)
	if err != nil && err.Error() != repo.ErrNoRTData {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	var res api.DroneDetailResult
	if err := copier.Copy(&res, &drone); err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	res.IsThermalAvailable = drone.IsThermalAvailable()
	res.IsRTKAvailable = drone.IsRTKAvailable()
	return c.JSON(Success(res))
}

// update 更新无人机信息
//
//	@Router			/drone/:sn [put]
//	@Summary		更新无人机信息
//	@Description	更新无人机信息
//	@Tags			drone
//	@Accept			json
//	@Produce		json
//	@Param			sn		path		string					true	"无人机SN"
//	@Param			request	body		v1.DroneUpdateRequest	true	"无人机信息"
//	@Success		200		{object}	v1.Response{data=nil}	"成功"
func (r *DroneRouter) update(c *fiber.Ctx) error {
	sn := c.Params("sn")
	req := new(api.DroneUpdateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}
	ctx := context.Background()
	if err := r.svc.Repo().UpdateCallsign(ctx, sn, req.Callsign); err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	return c.JSON(Success(nil))
}

func (r *DroneRouter) pushState(c *fiber.Ctx) error {
	// 设置SSE必需的响应头
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// 获取无人机序列号
	sn := c.Query("sn")
	r.l.Info("SSE sn", "sn", sn)

	ctx := context.Background()

	// 使用流式响应
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		r.l.Info("SSE connection established")
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 构造消息并尝试写入
				drone, err := r.svc.Repo().FetchStateBySN(ctx, sn)
				if err != nil {
					r.l.Error("Fetch drone state failed", "error", err)
					return
				}
				res := api.DroneState{
					SN:      sn,
					Lat:     drone.Latitude,
					Lng:     drone.Longitude,
					Height:  drone.Height,
					Speed:   drone.HorizontalSpeed,
					Battery: drone.Battery.CapacityPercent,
				}
				json, err := sonic.Marshal(res)
				if err != nil {
					r.l.Error("SSE marshal error", "error", err)
					return
				}
				msg := fmt.Sprintf("data: %s\n\n", json)
				r.l.Info("Sending message", "msg", msg)

				// 写入消息并立即刷新
				if _, err := w.WriteString(msg); err != nil {
					r.l.Error("SSE write error", "error", err)
					return
				}
				if err := w.Flush(); err != nil {
					r.l.Error("SSE flush error", "error", err)
					return
				}
				r.l.Info("Message sent and flushed")
			}
		}
	})

	return nil
}
