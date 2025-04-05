package v1

import (
	"context"
	"log/slog"

	"github.com/asaskevich/EventBus"
	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
)

type GatewayRouter struct {
	svc service.GatewaySvc
	eb  EventBus.Bus
	l   *slog.Logger
}

// NewGatewayRouter 创建网关路由处理器
func NewGatewayRouter(handler fiber.Router, svc service.GatewaySvc, eb EventBus.Bus, l *slog.Logger) {
	r := &GatewayRouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/gateway")
	{
		h.Get("/list", r.list)                   // 获取网关列表
		h.Put("/:sn", r.update)                  // 更新网关信息
		h.Get("/sn/:sn", r.getBySN)              // 获取单个网关详情
		h.Get("/sn/:sn/drones", r.getDronesBySN) // 获取网关关联的无人机列表
	}
}

// gatewayItemResult 网关列表项响应结构
type gatewayItemResult struct {
	ID           uint   `json:"id"`             // ID
	Callsign     string `json:"callsign"`       // 呼号
	SN           string `json:"sn"`             // 序列号
	Description  string `json:"description"`    // 描述
	Status       string `json:"status"`         // 在线状态
	ProductModel string `json:"product_model"`  // 产品型号
	CreatedAt    string `json:"created_at"`     // 创建时间
	LastOnlineAt string `json:"last_online_at"` // 最后在线时间
}

// list 获取网关列表
func (r *GatewayRouter) list(c *fiber.Ctx) error {
	ctx := context.Background()
	gateways, err := r.svc.Repo().SelectAll(ctx)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	var res []gatewayItemResult
	for _, g := range gateways {
		var e gatewayItemResult
		if err := copier.Copy(&e, &g); err != nil {
			return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
		}
		e.Status = g.StatusText()
		e.CreatedAt = g.CreatedAt.Format("2006-01-02 15:04:05")
		e.LastOnlineAt = g.LastOnlineAt.Format("2006-01-02 15:04:05")
		res = append(res, e)
	}

	return c.JSON(Success(res))
}

// gatewayDetailResult 网关详情响应结构
type gatewayDetailResult struct {
	ID           uint     `json:"id"`
	SN           string   `json:"sn"`            // 序列号
	Callsign     string   `json:"callsign"`      // 呼号
	Description  string   `json:"description"`   // 描述
	Status       string   `json:"status"`        // 在线状态
	ProductModel string   `json:"product_model"` // 产品型号
	DroneList    []string `json:"drone_list"`    // 关联的无人机列表
}

// getBySN 根据序列号获取网关详情
func (r *GatewayRouter) getBySN(c *fiber.Ctx) error {
	sn := c.Params("sn")
	ctx := context.Background()
	gateway, err := r.svc.Repo().SelectBySN(ctx, sn)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	var res gatewayDetailResult
	if err := copier.Copy(&res, &gateway); err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}
	res.Status = gateway.StatusText()

	// 获取关联的无人机列表
	drones, err := r.svc.Repo().GetConnectedDrones(ctx, sn)
	if err != nil {
		r.l.Error("获取关联无人机失败", "error", err)
	} else {
		for _, drone := range drones {
			res.DroneList = append(res.DroneList, drone.SN)
		}
	}

	return c.JSON(Success(res))
}

// droneListResult 无人机列表响应结构
type droneListResult struct {
	SN        string `json:"sn"`         // 序列号
	Callsign  string `json:"callsign"`   // 呼号
	Status    string `json:"status"`     // 在线状态
	ModelName string `json:"model_name"` // 型号名称
}

// getDronesBySN 获取网关关联的无人机列表
func (r *GatewayRouter) getDronesBySN(c *fiber.Ctx) error {
	sn := c.Params("sn")
	ctx := context.Background()

	// 获取关联的无人机列表
	drones, err := r.svc.Repo().GetConnectedDrones(ctx, sn)
	if err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	var res []droneListResult
	for _, drone := range drones {
		res = append(res, droneListResult{
			SN:        drone.SN,
			Callsign:  drone.Callsign,
			Status:    drone.GetStatusText(),
			ModelName: drone.DroneModel.Name,
		})
	}

	return c.JSON(Success(res))
}

// update 更新网关信息
func (r *GatewayRouter) update(c *fiber.Ctx) error {
	sn := c.Params("sn")
	req := new(api.GatewayUpdateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Fail(ErrorBody{Code: 400, Msg: err.Error()}))
	}

	ctx := context.Background()
	if err := r.svc.Repo().UpdateCallsign(ctx, sn, req.Callsign); err != nil {
		return c.JSON(Fail(ErrorBody{Code: 500, Msg: err.Error()}))
	}

	return c.JSON(Success(nil))
}
