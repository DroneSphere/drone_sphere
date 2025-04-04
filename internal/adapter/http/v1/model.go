package v1

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
)

// 无人机型号创建请求
type CreateDroneModelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      int    `json:"domain"`
	Type        int    `json:"type"`
	SubType     int    `json:"sub_type"`
	GatewayID   uint   `json:"gateway_id"`
	GimbalIDs   []uint `json:"gimbal_ids,omitempty"`
	PayloadIDs  []uint `json:"payload_ids,omitempty"`
}

// 云台型号创建请求
type CreateGimbalModelRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	Product            string `json:"product"`
	Domain             int    `json:"domain"`
	Type               int    `json:"type"`
	SubType            int    `json:"sub_type"`
	Gimbalindex        int    `json:"gimbalindex"`
	IsThermalAvailable bool   `json:"is_thermal_available"`
}

// 网关型号创建请求
type CreateGatewayModelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      int    `json:"domain"`
	Type        int    `json:"type"`
	SubType     int    `json:"sub_type"`
}

// 负载型号创建请求
type CreatePayloadModelRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	IsRTKAvailable bool   `json:"is_rtk_available"`
}

// 批量创建型号请求
type BatchCreateModelsRequest struct {
	DroneModels   []CreateDroneModelRequest   `json:"drone_models"`
	GimbalModels  []CreateGimbalModelRequest  `json:"gimbal_models"`
	GatewayModels []CreateGatewayModelRequest `json:"gateway_models"`
	PayloadModels []CreatePayloadModelRequest `json:"payload_models"`
}

type ModelRouter struct {
	svc service.ModelSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func NewModelsRouter(handler fiber.Router, svc service.ModelSvc, eb EventBus.Bus, l *slog.Logger) {
	r := &ModelRouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/models")
	{
		// 获取型号列表接口
		h.Get("/gateways", r.getGatewayModels)
		h.Get("/drones", r.getDroneModels)
		h.Get("/gimbals", r.getGimbalModels)
		h.Get("/payloads", r.getPaylodModels)

		// 单体型号创建接口
		h.Post("/drone", r.createDroneModel)
		h.Post("/gimbal", r.createGimbalModel)
		h.Post("/gateway", r.createGatewayModel)
		h.Post("/payload", r.createPayloadModel)

		// 批量创建型号接口
		h.Post("/batch", r.batchCreateModels)

		// 生成无人机变体
		h.Post("/variations/:id", r.generateDroneVariations)
	}
}

func (r *ModelRouter) getGatewayModels(c *fiber.Ctx) error {
	ctx := context.Background()
	models, err := r.svc.Repo().SelectAllGatewayModel(ctx)
	if err != nil {
		r.l.Error("get gateway models", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(models))
}

func (r *ModelRouter) getDroneModels(c *fiber.Ctx) error {
	ctx := context.Background()
	res, err := r.svc.Repo().SelectAllDroneModel(ctx)
	if err != nil {
		r.l.Error("get drone models", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(res))
}

func (r *ModelRouter) getGimbalModels(c *fiber.Ctx) error {
	ctx := context.Background()
	res, err := r.svc.Repo().SelectAllGimbals(ctx)
	if err != nil {
		r.l.Error("get gimbal models", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(res))
}

func (r *ModelRouter) getPaylodModels(c *fiber.Ctx) error {
	ctx := context.Background()
	res, err := r.svc.Repo().SelectAllPayloadModel(ctx)
	if err != nil {
		r.l.Error("get payload models", "error", err)
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(res))
}

// 创建无人机型号
func (r *ModelRouter) createDroneModel(c *fiber.Ctx) error {
	var req CreateDroneModelRequest
	if err := c.BodyParser(&req); err != nil {
		r.l.Error("解析无人机型号创建请求失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 构建无人机型号对象
	droneModel := &po.DroneModel{
		Name:           req.Name,
		Description:    req.Description,
		Domain:         req.Domain,
		Type:           req.Type,
		SubType:        req.SubType,
		GatewayID:      req.GatewayID,
		IsRTKAvailable: false, // 默认不支持RTK，可以根据需要修改
	}

	// 创建无人机型号
	ctx := context.Background()
	if err := r.svc.Repo().CreateDroneModel(ctx, droneModel); err != nil {
		r.l.Error("创建无人机型号失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	// 如果请求中包含云台ID，则建立关联
	if len(req.GimbalIDs) > 0 {
		gimbals, err := r.svc.Repo().SelectGimbalsByIDs(ctx, req.GimbalIDs)
		if err != nil {
			r.l.Error("查询云台型号失败", "error", err)
			return c.JSON(Fail(InternalError))
		}
		droneModel.Gimbals = gimbals
	}

	// 创建无人机型号后，自动生成无人机变体
	variations, err := r.svc.Repo().GenerateDroneVariations(ctx, droneModel.ID)
	if err != nil {
		r.l.Error("生成无人机变体失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(map[string]interface{}{
		"drone_model": droneModel,
		"variations":  variations,
	}))
}

// 创建云台型号
func (r *ModelRouter) createGimbalModel(c *fiber.Ctx) error {
	var req CreateGimbalModelRequest
	if err := c.BodyParser(&req); err != nil {
		r.l.Error("解析云台型号创建请求失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 构建云台型号对象
	gimbalModel := &po.GimbalModel{
		Name:               req.Name,
		Description:        req.Description,
		Product:            req.Product,
		Domain:             req.Domain,
		Type:               req.Type,
		SubType:            req.SubType,
		Gimbalindex:        req.Gimbalindex,
		IsThermalAvailable: req.IsThermalAvailable,
	}

	// 创建云台型号
	ctx := context.Background()
	if err := r.svc.Repo().CreateGimbalModel(ctx, gimbalModel); err != nil {
		r.l.Error("创建云台型号失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(gimbalModel))
}

// 创建网关型号
func (r *ModelRouter) createGatewayModel(c *fiber.Ctx) error {
	var req CreateGatewayModelRequest
	if err := c.BodyParser(&req); err != nil {
		r.l.Error("解析网关型号创建请求失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 构建网关型号对象
	gatewayModel := &po.GatewayModel{
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		Type:        req.Type,
		SubType:     req.SubType,
	}

	// 创建网关型号
	ctx := context.Background()
	if err := r.svc.Repo().CreateGatewayModel(ctx, gatewayModel); err != nil {
		r.l.Error("创建网关型号失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(gatewayModel))
}

// 创建负载型号
func (r *ModelRouter) createPayloadModel(c *fiber.Ctx) error {
	var req CreatePayloadModelRequest
	if err := c.BodyParser(&req); err != nil {
		r.l.Error("解析负载型号创建请求失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 构建负载型号对象
	payloadModel := &po.PayloadModel{
		Name:           req.Name,
		Description:    req.Description,
		Category:       req.Category,
		IsRTKAvailable: req.IsRTKAvailable,
	}

	// 创建负载型号
	ctx := context.Background()
	if err := r.svc.Repo().CreatePayloadModel(ctx, payloadModel); err != nil {
		r.l.Error("创建负载型号失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(payloadModel))
}

// 批量创建型号
func (r *ModelRouter) batchCreateModels(c *fiber.Ctx) error {
	var req BatchCreateModelsRequest
	if err := c.BodyParser(&req); err != nil {
		r.l.Error("解析批量创建型号请求失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 转换请求数据为数据库模型
	var droneModels []po.DroneModel
	var gimbalModels []po.GimbalModel
	var gatewayModels []po.GatewayModel
	var payloadModels []po.PayloadModel

	// 转换网关型号
	for _, gm := range req.GatewayModels {
		gatewayModels = append(gatewayModels, po.GatewayModel{
			Name:        gm.Name,
			Description: gm.Description,
			Domain:      gm.Domain,
			Type:        gm.Type,
			SubType:     gm.SubType,
		})
	}

	// 转换无人机型号
	for _, dm := range req.DroneModels {
		droneModels = append(droneModels, po.DroneModel{
			Name:        dm.Name,
			Description: dm.Description,
			Domain:      dm.Domain,
			Type:        dm.Type,
			SubType:     dm.SubType,
			GatewayID:   dm.GatewayID,
		})
	}

	// 转换云台型号
	for _, gm := range req.GimbalModels {
		gimbalModels = append(gimbalModels, po.GimbalModel{
			Name:               gm.Name,
			Description:        gm.Description,
			Product:            gm.Product,
			Domain:             gm.Domain,
			Type:               gm.Type,
			SubType:            gm.SubType,
			Gimbalindex:        gm.Gimbalindex,
			IsThermalAvailable: gm.IsThermalAvailable,
		})
	}

	// 转换负载型号
	for _, pm := range req.PayloadModels {
		payloadModels = append(payloadModels, po.PayloadModel{
			Name:           pm.Name,
			Description:    pm.Description,
			Category:       pm.Category,
			IsRTKAvailable: pm.IsRTKAvailable,
		})
	}

	// 批量创建型号
	ctx := context.Background()
	if err := r.svc.BatchCreateModels(ctx, droneModels, gimbalModels, gatewayModels, payloadModels); err != nil {
		r.l.Error("批量创建型号失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(map[string]interface{}{
		"message": "批量创建型号成功",
	}))
}

// 生成无人机变体
func (r *ModelRouter) generateDroneVariations(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("生成无人机变体失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	ctx := context.Background()
	variations, err := r.svc.Repo().GenerateDroneVariations(ctx, uint(id))
	if err != nil {
		r.l.Error("生成无人机变体失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(variations))
}
