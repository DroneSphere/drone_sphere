package v1

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

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

// 更新无人机型号请求
type UpdateDroneModelRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Domain         int    `json:"domain"`
	Type           int    `json:"type"`
	SubType        int    `json:"sub_type"`
	GatewayID      uint   `json:"gateway_id"`
	GimbalIDs      []uint `json:"gimbal_ids,omitempty"`
	PayloadIDs     []uint `json:"payload_ids,omitempty"`
	IsRTKAvailable *bool  `json:"is_rtk_available,omitempty"`
}

// 更新云台型号请求
type UpdateGimbalModelRequest struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	Product            string `json:"product"`
	Domain             int    `json:"domain"`
	Type               int    `json:"type"`
	SubType            int    `json:"sub_type"`
	Gimbalindex        int    `json:"gimbalindex"`
	IsThermalAvailable bool   `json:"is_thermal_available"`
}

// 更新网关型号请求
type UpdateGatewayModelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      int    `json:"domain"`
	Type        int    `json:"type"`
	SubType     int    `json:"sub_type"`
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
		// 获取型号列表接口 (已增加名称查询)
		h.Get("/gateways", r.getGatewayModels)
		h.Get("/drones", r.getDroneModels)
		h.Get("/gimbals", r.getGimbalModels)
		h.Get("/payloads", r.getPayloadModels)

		// 单条型号查询接口
		h.Get("/gateway/:id", r.getGatewayModelByID)
		h.Get("/drone/:id", r.getDroneModelByID)
		h.Get("/gimbal/:id", r.getGimbalModelByID)

		// 单体型号创建接口
		h.Post("/drone", r.createDroneModel)
		h.Post("/gimbal", r.createGimbalModel)
		h.Post("/gateway", r.createGatewayModel)
		h.Post("/payload", r.createPayloadModel)

		// 型号更新接口
		h.Put("/drone/:id", r.updateDroneModel)
		h.Put("/gimbal/:id", r.updateGimbalModel)
		h.Put("/gateway/:id", r.updateGatewayModel)

		// 型号删除接口 (软删除，将状态置为-1)
		h.Delete("/drone/:id", r.deleteDroneModel)
		h.Delete("/gimbal/:id", r.deleteGimbalModel)
		h.Delete("/gateway/:id", r.deleteGatewayModel)

		// 批量创建型号接口
		h.Post("/batch", r.batchCreateModels)

		// 生成无人机变体
		h.Post("/variations/:id", r.generateDroneVariations)
	}
}

func (r *ModelRouter) getGatewayModels(c *fiber.Ctx) error {
	ctx := context.Background()

	// 获取查询参数
	name := c.Query("name", "")

	// 获取所有网关型号
	models, err := r.svc.Repo().SelectAllGatewayModel(ctx, name)
	if err != nil {
		r.l.Error("获取网关型号列表失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(models))
}

func (r *ModelRouter) getDroneModels(c *fiber.Ctx) error {
	ctx := context.Background()

	// 获取查询参数
	name := c.Query("name", "")

	// 获取所有无人机型号
	models, err := r.svc.Repo().SelectAllDroneModel(ctx, name)
	if err != nil {
		r.l.Error("获取无人机型号列表失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	// 如果有名称查询参数，则进行过滤
	if name != "" {
		var filteredModels []interface{}
		for _, model := range models {
			if strings.Contains(strings.ToLower(model.Name), strings.ToLower(name)) {
				filteredModels = append(filteredModels, model)
			}
		}
		return c.JSON(Success(filteredModels))
	}

	return c.JSON(Success(models))
}

func (r *ModelRouter) getGimbalModels(c *fiber.Ctx) error {
	ctx := context.Background()

	// 获取查询参数
	name := c.Query("name", "")

	// 获取所有云台型号
	models, err := r.svc.Repo().SelectAllGimbals(ctx)
	if err != nil {
		r.l.Error("获取云台型号列表失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	// 如果有名称查询参数，则进行过滤
	if name != "" {
		var filteredModels []po.GimbalModel
		for _, model := range models {
			if strings.Contains(strings.ToLower(model.Name), strings.ToLower(name)) {
				filteredModels = append(filteredModels, model)
			}
		}
		return c.JSON(Success(filteredModels))
	}

	return c.JSON(Success(models))
}

func (r *ModelRouter) getPayloadModels(c *fiber.Ctx) error {
	ctx := context.Background()
	res, err := r.svc.Repo().SelectAllPayloadModel(ctx)
	if err != nil {
		r.l.Error("get payload models", "error", err)
		return c.JSON(Fail(InternalError))
	}
	return c.JSON(Success(res))
}

// 根据ID获取网关型号
func (r *ModelRouter) getGatewayModelByID(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析网关型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 获取所有网关型号
	ctx := context.Background()
	allModels, err := r.svc.Repo().SelectAllGatewayModel(ctx, "")
	if err != nil {
		r.l.Error("获取网关型号列表失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	// 查找匹配ID的网关型号
	for _, model := range allModels {
		if model.ID == uint(id) {
			return c.JSON(Success(model))
		}
	}

	// 未找到对应ID的网关型号
	r.l.Error("未找到对应ID的网关型号", "id", id)
	return c.JSON(Fail(NotFound))
}

// 根据ID获取无人机型号
func (r *ModelRouter) getDroneModelByID(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析无人机型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 获取所有无人机型号
	ctx := context.Background()
	allModels, err := r.svc.Repo().SelectAllDroneModel(ctx, "")
	if err != nil {
		r.l.Error("获取无人机型号列表失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	// 查找匹配ID的无人机型号
	for _, model := range allModels {
		if model.ID == uint(id) {
			return c.JSON(Success(model))
		}
	}

	// 未找到对应ID的无人机型号
	r.l.Error("未找到对应ID的无人机型号", "id", id)
	return c.JSON(Fail(NotFound))
}

// 根据ID获取云台型号
func (r *ModelRouter) getGimbalModelByID(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析云台型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 获取所有云台型号
	ctx := context.Background()
	allModels, err := r.svc.Repo().SelectAllGimbals(ctx)
	if err != nil {
		r.l.Error("获取云台型号列表失败", "error", err)
		return c.JSON(Fail(InternalError))
	}

	// 查找匹配ID的云台型号
	for _, model := range allModels {
		if model.ID == uint(id) {
			return c.JSON(Success(model))
		}
	}

	// 未找到对应ID的云台型号
	r.l.Error("未找到对应ID的云台型号", "id", id)
	return c.JSON(Fail(NotFound))
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

// 更新无人机型号
func (r *ModelRouter) updateDroneModel(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析无人机型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 解析请求体
	var req UpdateDroneModelRequest
	if err := c.BodyParser(&req); err != nil {
		r.l.Error("解析无人机型号更新请求失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 直接构建要更新的字段
	ctx := context.Background()

	// 准备返回给客户端的对象
	var targetModel po.DroneModel = po.DroneModel{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		Type:        req.Type,
		SubType:     req.SubType,
		GatewayID:   req.GatewayID,
	}

	// 如果请求包含RTK可用性标志，则更新
	if req.IsRTKAvailable != nil {
		targetModel.IsRTKAvailable = *req.IsRTKAvailable
	}

	// 准备更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["drone_model_name"] = req.Name
	}
	if req.Description != "" {
		updates["drone_model_description"] = req.Description
	}
	if req.Domain != 0 {
		updates["drone_model_domain"] = req.Domain
	}
	if req.Type != 0 {
		updates["drone_model_type"] = req.Type
	}
	if req.SubType != 0 {
		updates["drone_model_sub_type"] = req.SubType
	}
	if req.GatewayID != 0 {
		updates["gateway_model_id"] = req.GatewayID
	}

	// 如果请求包含RTK可用性标志，则更新
	if req.IsRTKAvailable != nil {
		updates["is_rtk_available"] = *req.IsRTKAvailable
	}

	// 直接通过ID更新无人机型号
	if err := r.svc.Repo().UpdateDroneModelFields(ctx, uint(id), updates); err != nil {
		// 如果是记录不存在的错误，返回NotFound
		if strings.Contains(err.Error(), "record not found") {
			r.l.Error("未找到对应ID的无人机型号", "id", id, "error", err)
			return c.JSON(Fail(NotFound))
		}
		r.l.Error("更新无人机型号失败", "id", id, "error", err)
		return c.JSON(Fail(InternalError))
	}

	// 如果请求中包含云台ID，则获取云台信息并建立关联
	if len(req.GimbalIDs) > 0 {
		gimbals, err := r.svc.Repo().SelectGimbalsByIDs(ctx, req.GimbalIDs)
		if err != nil {
			r.l.Error("查询云台型号失败", "error", err)
			return c.JSON(Fail(InternalError))
		}

		// 添加到返回对象
		targetModel.Gimbals = gimbals

		// 构建完整模型用于更新关联信息
		fullModel := po.DroneModel{
			ID:      uint(id),
			Gimbals: gimbals,
		}

		// 更新关联信息
		if err := r.svc.Repo().UpdateDroneModel(ctx, uint(id), &fullModel); err != nil {
			r.l.Error("更新无人机型号关联失败", "id", id, "error", err)
			return c.JSON(Fail(InternalError))
		}
	}

	return c.JSON(Success(targetModel))
}

// 更新云台型号
func (r *ModelRouter) updateGimbalModel(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析云台型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 解析请求体
	var req UpdateGimbalModelRequest
	if err := c.BodyParser(&req); err != nil {
		r.l.Error("解析云台型号更新请求失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 直接构建要更新的字段
	ctx := context.Background()

	// 准备返回给客户端的对象
	targetModel := po.GimbalModel{
		ID:                 uint(id),
		Name:               req.Name,
		Description:        req.Description,
		Product:            req.Product,
		Domain:             req.Domain,
		Type:               req.Type,
		SubType:            req.SubType,
		Gimbalindex:        req.Gimbalindex,
		IsThermalAvailable: req.IsThermalAvailable,
	}

	// 准备更新字段
	updates := map[string]interface{}{
		"gimbal_model_name":        req.Name,
		"gimbal_model_description": req.Description,
		"gimbal_model_product":     req.Product,
		"gimbal_model_domain":      req.Domain,
		"gimbal_model_type":        req.Type,
		"gimbal_model_sub_type":    req.SubType,
		"gimbalindex":              req.Gimbalindex,
		"is_thermal_available":     req.IsThermalAvailable,
	}

	// 直接通过ID更新云台型号
	if err := r.svc.Repo().UpdateGimbalModelFields(ctx, uint(id), updates); err != nil {
		// 如果是记录不存在的错误，返回NotFound
		if strings.Contains(err.Error(), "record not found") {
			r.l.Error("未找到对应ID的云台型号", "id", id, "error", err)
			return c.JSON(Fail(NotFound))
		}
		r.l.Error("更新云台型号失败", "id", id, "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(targetModel))
}

// 更新网关型号
func (r *ModelRouter) updateGatewayModel(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析网关型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 解析请求体
	var req UpdateGatewayModelRequest
	if err := c.BodyParser(&req); err != nil {
		r.l.Error("解析网关型号更新请求失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 直接构建要更新的字段
	ctx := context.Background()

	// 准备返回给客户端的对象
	targetModel := po.GatewayModel{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		Domain:      req.Domain,
		Type:        req.Type,
		SubType:     req.SubType,
	}

	// 准备更新字段
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["gateway_model_name"] = req.Name
	}
	if req.Description != "" {
		updates["gateway_model_description"] = req.Description
	}
	if req.Domain != 0 {
		updates["gateway_model_domain"] = req.Domain
	}
	if req.Type != 0 {
		updates["gateway_model_type"] = req.Type
	}
	if req.SubType != 0 {
		updates["gateway_model_sub_type"] = req.SubType
	}
	if len(updates) == 0 {
		r.l.Error("没有需要更新的字段", "id", id)
		return c.JSON(Fail(InvalidParams))
	}

	// 直接通过ID更新网关型号
	if err := r.svc.Repo().UpdateGatewayModelFields(ctx, uint(id), updates); err != nil {
		// 如果是记录不存在的错误，返回NotFound
		if strings.Contains(err.Error(), "record not found") {
			r.l.Error("未找到对应ID的网关型号", "id", id, "error", err)
			return c.JSON(Fail(NotFound))
		}
		r.l.Error("更新网关型号失败", "id", id, "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(targetModel))
}

// 删除无人机型号（软删除，将state置为-1）
func (r *ModelRouter) deleteDroneModel(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析无人机型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 直接通过ID进行软删除
	ctx := context.Background()

	// 准备软删除更新 (将state置为-1)
	updates := map[string]interface{}{
		"state": -1, // 将state置为-1表示软删除
	}

	// 直接更新无人机型号的状态
	if err := r.svc.Repo().UpdateDroneModelFields(ctx, uint(id), updates); err != nil {
		// 如果是记录不存在的错误，返回NotFound
		if strings.Contains(err.Error(), "record not found") {
			r.l.Error("未找到对应ID的无人机型号", "id", id, "error", err)
			return c.JSON(Fail(NotFound))
		}
		r.l.Error("删除无人机型号失败", "id", id, "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(map[string]interface{}{
		"message": "无人机型号删除成功",
		"id":      id,
	}))
}

// 删除云台型号（软删除，将state置为-1）
func (r *ModelRouter) deleteGimbalModel(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析云台型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 直接通过ID进行软删除
	ctx := context.Background()

	// 准备软删除更新 (将state置为-1)
	updates := map[string]interface{}{
		"state": -1, // 将state置为-1表示软删除
	}

	// 直接更新云台型号的状态
	if err := r.svc.Repo().UpdateGimbalModelFields(ctx, uint(id), updates); err != nil {
		// 如果是记录不存在的错误，返回NotFound
		if strings.Contains(err.Error(), "record not found") {
			r.l.Error("未找到对应ID的云台型号", "id", id, "error", err)
			return c.JSON(Fail(NotFound))
		}
		r.l.Error("删除云台型号失败", "id", id, "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(map[string]interface{}{
		"message": "云台型号删除成功",
		"id":      id,
	}))
}

// 删除网关型号（软删除，将state置为-1）
func (r *ModelRouter) deleteGatewayModel(c *fiber.Ctx) error {
	// 获取路径参数
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		r.l.Error("解析网关型号ID失败", "error", err)
		return c.JSON(Fail(InvalidParams))
	}

	// 直接通过ID进行软删除
	ctx := context.Background()

	// 准备软删除更新 (将state置为-1)
	updates := map[string]interface{}{
		"state": -1, // 将state置为-1表示软删除
	}

	// 直接更新网关型号的状态
	if err := r.svc.Repo().UpdateGatewayModelFields(ctx, uint(id), updates); err != nil {
		// 如果是记录不存在的错误，返回NotFound
		if strings.Contains(err.Error(), "record not found") {
			r.l.Error("未找到对应ID的网关型号", "id", id, "error", err)
			return c.JSON(Fail(NotFound))
		}
		r.l.Error("删除网关型号失败", "id", id, "error", err)
		return c.JSON(Fail(InternalError))
	}

	return c.JSON(Success(map[string]interface{}{
		"message": "网关型号删除成功",
		"id":      id,
	}))
}
