package v1

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/EventBus"
	"github.com/bytedance/sonic"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/model/entity"
	"github.com/dronesphere/internal/model/ro"
	"github.com/dronesphere/internal/repo"
	"github.com/dronesphere/internal/service"
	"github.com/dronesphere/pkg/coordinate"
	"github.com/gofiber/contrib/websocket"
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
		h.Post("/", r.create)
		h.Get("/sn/:sn", r.getBySN)
		h.Get("/state/sse", r.pushState)
		h.Get("/models", r.getModels)          // 添加获取无人机型号列表的路由
		h.Delete("/:sn", r.delete)             // 添加删除无人机的路由
		h.Post("/:sn/live/start", r.startLive) // 添加启动直播的路由
		h.Post("/:sn/live/stop", r.stopLive)   // 添加停止直播的路由
	}
	h.Use("/:sn/control", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	h.Get("/:sn/control", websocket.New(r.handleDroneControl))
}

// handleDroneControl 处理无人机控制的 WebSocket 连接
func (r *DroneRouter) handleDroneControl(c *websocket.Conn) {
	// 获取无人机序列号
	sn := c.Params("sn")
	if sn == "" {
		r.l.Error("无人机序列号(SN)不能为空")
		c.Close()
		return
	}

	r.l.Info("无人机控制连接已建立", slog.String("sn", sn))
	ctx := context.Background()
	r.svc.CheckControlConnection(ctx, c, sn)
	// 处理 WebSocket 消息循环
	for {
		// 读取客户端消息
		mt, msg, err := c.ReadMessage()
		if err != nil {
			r.l.Info("无人机控制连接已关闭", slog.String("sn", sn))
			break
		}
		r.l.Info("收到无人机控制消息", slog.String("sn", sn), slog.String("message", string(msg)))

		r.svc.HandleControlSession(ctx, c, sn, mt, string(msg))
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

	// 从请求中获取查询参数
	sn := c.Query("sn")
	callsign := c.Query("callsign")
	modelIDStr := c.Query("model_id")
	pageStr := c.Query("page", "1")           // 默认页码为1
	pageSizeStr := c.Query("page_size", "10") // 默认每页10条数据

	// 解析 model_id 参数
	var modelID uint = 0
	if modelIDStr != "" {
		id, err := strconv.ParseUint(modelIDStr, 10, 32)
		if err != nil {
			return c.JSON(Fail(ErrorBody{Code: 400, Msg: "无效的型号ID参数: " + err.Error()}))
		}
		modelID = uint(id)
	}
	// 解析页码和每页大小
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "无效的页码参数"}))
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "无效的每页大小参数"}))
	}

	// 调用仓库层方法，传递查询条件
	drones, total, err := r.svc.Repo().SelectAll(ctx, sn, callsign, modelID, page, pageSize)
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

	return c.JSON(Success(fiber.Map{
		"total": total,
		"items": res,
	}))
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
				lng, lat := coordinate.WGS84ToGCJ02(drone.Longitude, drone.Latitude)
				res := api.DroneState{
					SN:      sn,
					Lat:     lat,
					Lng:     lng,
					Height:  drone.Height,
					Speed:   drone.HorizontalSpeed,
					Battery: drone.Battery.CapacityPercent,
					Heading: drone.AttitudeHead,
					Pitch:   drone.AttitudePitch, // 飞行器俯仰角
					Yaw:     drone.AttitudeHead,  // 飞行器偏航角
					Roll:    drone.AttitudeRoll,  // 飞行器横滚角
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

func (r *DroneRouter) create(c *fiber.Ctx) error {
	// 解析请求体
	req := new(api.DroneCreateRequest)
	if err := c.BodyParser(req); err != nil {
		r.l.Error("无人机创建请求解析失败", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: "请求参数解析失败: " + err.Error()}))
	}

	// 验证必要字段
	if req.SN == "" {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "无人机序列号(SN)不能为空"}))
	}
	if req.DroneModelID == 0 {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "无人机型号ID不能为0"}))
	}

	// 检查是否已存在相同SN的无人机
	ctx := context.Background()
	_, err := r.svc.Repo().SelectBySN(ctx, req.SN)
	if err == nil {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "序列号为 " + req.SN + " 的无人机已存在"}))
	}

	// 默认呼号使用SN的前8位
	if req.Callsign == "" && len(req.SN) >= 8 {
		req.Callsign = req.SN[:8]
	} else if req.Callsign == "" {
		req.Callsign = req.SN
	}

	// 创建无人机实体
	drone := entity.Drone{
		SN:           req.SN,
		Callsign:     req.Callsign,
		Description:  req.Description,
		DroneModelID: req.DroneModelID,
		VariationID:  req.VariationID,
		Status:       ro.DroneStatusOffline, // 新创建的无人机默认为离线状态
	}

	// 保存无人机信息
	if err := r.svc.Repo().Save(ctx, drone); err != nil {
		r.l.Error("创建无人机失败", slog.Any("drone", drone), slog.Any("error", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: "创建无人机失败: " + err.Error()}))
	}

	// 获取创建后的无人机详情
	createdDrone, err := r.svc.Repo().SelectBySN(ctx, req.SN)
	if err != nil {
		r.l.Error("获取新创建的无人机信息失败", slog.Any("sn", req.SN), slog.Any("error", err))
		return c.JSON(Success(nil)) // 返回成功但没有详细信息
	}

	// 转换为响应结果
	var result droneDetailResult
	if err := copier.Copy(&result, &createdDrone); err != nil {
		r.l.Error("复制无人机信息失败", slog.Any("error", err))
		return c.JSON(Success(nil)) // 返回成功但没有详细信息
	}

	result.Status = createdDrone.StatusText()
	result.ProductModel = createdDrone.GetModelName()
	result.ProductModelKey = fmt.Sprintf("%d-%d-%d", 0, createdDrone.Type, createdDrone.SubType)
	result.IsRTKAvailable = createdDrone.IsRTKAvailable()
	result.IsThermalAvailable = createdDrone.IsThermalAvailable()

	// 返回创建成功的结果
	return c.JSON(Success(result))
}

// delete 删除无人机
//
//	@Router			/drone/:sn [delete]
//	@Summary		删除无人机
//	@Description	通过将无人机的state设置为-1来软删除无人机
//	@Tags			drone
//	@Produce		json
//	@Param			sn	path		string				true	"无人机SN"
//	@Success		200	{object}	Response{data=nil}	"成功"
//	@Failure		400	{object}	Response{data=ErrorBody}	"请求参数错误"
//	@Failure		500	{object}	Response{data=ErrorBody}	"服务器内部错误"
func (r *DroneRouter) delete(c *fiber.Ctx) error {
	// 获取路径参数中的无人机序列号
	sn := c.Params("sn")
	if sn == "" {
		return c.JSON(Fail(ErrorBody{Code: 400, Msg: "无人机序列号(SN)不能为空"}))
	}

	r.l.Info("尝试删除无人机", slog.String("sn", sn))

	// 检查无人机是否存在
	ctx := context.Background()
	_, err := r.svc.Repo().SelectBySN(ctx, sn)
	if err != nil && err.Error() != "no realtime data" {
		r.l.Error("删除无人机失败：无人机不存在", slog.String("sn", sn), slog.Any("error", err))
		return c.JSON(Fail(ErrorBody{Code: 404, Msg: "无人机不存在：" + err.Error()}))
	}

	// 构造更新字段映射，将state设置为-1表示删除
	updates := map[string]interface{}{
		"state": -1, // state = -1 表示软删除
	}

	// 更新无人机状态为删除状态
	if err := r.svc.Repo().UpdateDroneInfo(ctx, sn, updates); err != nil {
		r.l.Error("删除无人机失败", slog.String("sn", sn), slog.Any("error", err))
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: "删除无人机失败：" + err.Error()}))
	}

	r.l.Info("成功删除无人机", slog.String("sn", sn))
	return c.JSON(Success(nil))
}
func (r *DroneRouter) startLive(c *fiber.Ctx) error {
	sn := c.Params("sn")
	if sn == "" {
		return c.JSON(Fail(ErrorBody{Code: fiber.StatusBadRequest, Msg: "无人机序列号(SN)不能为空"}))
	}

	r.l.Info("尝试启动无人机直播", slog.String("sn", sn))

	// 调用服务层方法启动直播
	pullURL, err := r.svc.StartLiveBySN(c.Context(), sn)
	if err != nil {
		// 根据错误类型返回不同的HTTP状态码
		if err.Error() == fmt.Sprintf("无人机 %s 没有云台信息", sn) || strings.Contains(err.Error(), "获取无人机信息失败") {
			r.l.Error("启动无人机直播失败：无人机信息或云台信息错误", slog.String("sn", sn), slog.Any("error", err))
			return c.JSON(Fail(ErrorBody{Code: fiber.StatusNotFound, Msg: "启动无人机直播失败：" + err.Error()}))
		}
		r.l.Error("启动无人机直播失败", slog.String("sn", sn), slog.Any("error", err))
		return c.JSON(Fail(ErrorBody{Code: fiber.StatusInternalServerError, Msg: "启动无人机直播失败：" + err.Error()}))
	}

	r.l.Info("成功启动无人机直播", slog.String("sn", sn), slog.String("pull_url", pullURL))
	return c.JSON(Success(pullURL))
}

// stopLive 停止无人机直播
// @Router      /drone/{sn}/live/stop [post]
// @Summary     停止无人机直播
// @Description 根据无人机SN停止直播推流
// @Tags        drone
// @Produce     json
// @Param       sn  path        string  true    "无人机SN"
// @Success     200 {object}    Response{data=nil} "成功"
// @Failure     400 {object}    Response{data=ErrorBody}    "请求参数错误"
// @Failure     404 {object}    Response{data=ErrorBody}    "无人机不存在或无直播流"
// @Failure     500 {object}    Response{data=ErrorBody}    "服务器内部错误"
func (r *DroneRouter) stopLive(c *fiber.Ctx) error {
	sn := c.Params("sn")
	if sn == "" {
		return c.JSON(Fail(ErrorBody{Code: fiber.StatusBadRequest, Msg: "无人机序列号(SN)不能为空"}))
	}

	r.l.Info("尝试停止无人机直播", slog.String("sn", sn))

	// 调用服务层方法停止直播
	if err := r.svc.StopLiveBySN(c.Context(), sn); err != nil {
		// 根据错误类型返回不同的HTTP状态码
		// (可以根据 StopLiveBySN 可能返回的特定错误进行更精细的处理)
		if strings.Contains(err.Error(), "获取无人机信息失败") {
			r.l.Error("停止无人机直播失败：无人机信息错误", slog.String("sn", sn), slog.Any("error", err))
			return c.JSON(Fail(ErrorBody{Code: fiber.StatusNotFound, Msg: "停止无人机直播失败：" + err.Error()}))
		}
		if strings.Contains(err.Error(), "无人机当前没有正在进行的直播") {
			r.l.Info("无人机无活跃直播流，无需停止", slog.String("sn", sn))
			return c.JSON(Success(nil)) // 或者返回特定状态码表示无操作
		}
		r.l.Error("停止无人机直播失败", slog.String("sn", sn), slog.Any("error", err))
		return c.JSON(Fail(ErrorBody{Code: fiber.StatusInternalServerError, Msg: "停止无人机直播失败：" + err.Error()}))
	}

	r.l.Info("成功停止无人机直播", slog.String("sn", sn))
	return c.JSON(Success(nil))
}
