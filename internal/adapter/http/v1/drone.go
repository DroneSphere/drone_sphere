package v1

import (
	"bufio"
	"context"
	"fmt"
	"github.com/asaskevich/EventBus"
	"github.com/bytedance/sonic"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"log/slog"
	"time"
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
	drones, err := r.svc.ListAll(ctx)
	if err != nil {
		r.l.Warn("ListError", slog.Any("err", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	var res []api.DroneItemResult
	for _, d := range drones {
		var e api.DroneItemResult
		if err := copier.Copy(&e, &d); err != nil {
			r.l.Warn("CopyError", slog.Any("err", err))
			return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
		}
		// 检查是否在线
		e.Status = d.StatusText()
		res = append(res, e)
	}

	return c.JSON(Success(res))
}

func (r *DroneRouter) pushState(c *fiber.Ctx) error {
	// 设置SSE必需的响应头
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	// 使用流式响应
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		r.l.Info("SSE connection established")
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 构造消息并尝试写入
				drone, err := r.svc.FetchState(context.Background(), "1581F5FHC246H00DRM66")
				if err != nil {
					r.l.Error("Fetch drone state failed", "error", err)
					return
				}
				res := api.DroneState{
					SN:      drone.SN,
					Lat:     drone.Latitude,
					Lng:     drone.Longitude,
					Height:  drone.Height,
					Heading: drone.GetHeading(),
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
