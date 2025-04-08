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

// 无人机型号选择项 DTO
type DroneModelOption struct {
	ID   uint   `json:"id"`   // 型号 ID
	Name string `json:"name"` // 型号名称
}

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
		h.Get("/models", r.getModels) // 添加获取无人机型号列表的路由
	}
}

type droneItemResult struct {
	ID                 uint   `json:"id"`                   // ID
	Callsign           string `json:"callsign"`             // 呼号
	SN                 string `json:"sn"`                   // 序列号
	Description        string `json:"description"`          // 描述
	Status             string `json:"status"`               // 在线状态
	ProductModel       string `json:"product_model"`        // 产品型号
	IsRTKAvailable     bool   `json:"is_rtk_available"`     // 是否支持RTK
	IsThermalAvailable bool   `json:"is_thermal_available"` // 是否支持热成像
	CreatedAt          string `json:"created_at"`           // 创建时间
	LastOnlineAt       string `json:"last_online_at"`       // 最后在线时间
}

func (r *DroneRouter) list(c *fiber.Ctx) error {
	ctx := context.Background()
	drones, err := r.svc.Repo().SelectAll(ctx)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	var res []droneItemResult
	for _, d := range drones {
		var e droneItemResult
		if err := copier.Copy(&e, &d); err != nil {
			return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
		}
		// 检查是否在线
		e.Status = d.StatusText()
		e.ProductModel = d.GetModelName()
		e.CreatedAt = d.CreatedAt.Format("2006-01-02 15:04:05")
		e.LastOnlineAt = d.UpdatedAt.Format("2006-01-02 15:04:05")
		for _, g := range d.DroneModel.Gimbals {
			if g.IsThermalAvailable {
				e.IsThermalAvailable = true
				break
			}
		}
		res = append(res, e)
	}

	return c.JSON(Success(res))
}

type droneDetailResult struct {
	ID                 uint   `json:"id"`
	SN                 string `json:"sn" binding:"required"`                // 序列号
	Callsign           string `json:"callsign"`                             // 呼号
	Description        string `json:"description"`                          // 描述
	Domain             int    `json:"domain" binding:"required"`            // 领域
	Type               int    `json:"type" binding:"required"`              // 类型
	SubType            int    `json:"sub_type" binding:"required"`          // 子类型
	ProductModel       string `json:"product_model" binding:"required"`     // 产品型号
	ProductModelKey    string `json:"product_model_key" binding:"required"` // 产品型号标识符
	Status             string `json:"status"`                               // 在线状态
	IsRTKAvailable     bool   `json:"is_rtk_available"`                     // 是否支持RTK◊
	IsThermalAvailable bool   `json:"is_thermal_available"`                 // 是否支持热成像
}

func (r *DroneRouter) getBySN(c *fiber.Ctx) error {
	sn := c.Params("sn")
	ctx := context.Background()
	drone, err := r.svc.Repo().SelectBySN(ctx, sn)
	if err != nil && err.Error() != repo.ErrNoRTData {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	var res droneDetailResult
	if err := copier.Copy(&res, &drone); err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	res.IsThermalAvailable = drone.IsThermalAvailable()
	res.IsRTKAvailable = drone.IsRTKAvailable()
	res.Status = drone.StatusText()
	res.ProductModel = drone.GetModelName()
	res.ProductModelKey = fmt.Sprintf("%d-%d-%d", 0, drone.Type, drone.SubType)
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
	if sn == "" {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "sn is required"}))
	}
	req := new(struct {
		Callsign    string `json:"callsign"`    // 呼号
		Description string `json:"description"` // 描述
	})
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}

	// 构造更新字段映射
	updates := map[string]interface{}{
		"callsign":          req.Callsign,
		"drone_description": req.Description,
	}

	ctx := context.Background()
	if err := r.svc.Repo().UpdateDroneInfo(ctx, sn, updates); err != nil {
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
					Heading: drone.AttitudeHead,
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

// getModels 获取无人机型号列表
//
//	@Router			/drone/models [get]
//	@Summary		获取无人机型号列表
//	@Description	获取当前系统中已有无人机的型号列表，供前端下拉选择器使用
//	@Tags			drone
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=[]dto.DroneModelOption}	"成功"
func (r *DroneRouter) getModels(c *fiber.Ctx) error {
	ctx := context.Background()

	// 调用仓库层方法获取无人机型号列表
	models, err := r.svc.Repo().FetchDroneModelOptions(ctx)
	if err != nil {
		r.l.Error("获取无人机型号列表失败", slog.Any("error", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: "获取无人机型号列表失败: " + err.Error()}))
	}

	// 返回结果
	return c.JSON(Success(models))
}
