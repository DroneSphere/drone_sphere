package v1

import (
	"log/slog"

	"github.com/asaskevich/EventBus"
	"github.com/gofiber/fiber/v2"
)

type ModelRouter struct {
	eb EventBus.Bus
	l  *slog.Logger
}

func NewModelsRouter(handler fiber.Router, eb EventBus.Bus, l *slog.Logger) {
	r := &ModelRouter{
		eb: eb,
		l:  l,
	}
	h := handler.Group("/models")
	{
		h.Get("/gateways", r.getGatewayModels)
		h.Get("/drones", r.getDroneModels)
		h.Get("/gimbals", r.getGimbalModels)
		h.Get("/payloads", r.getPaylodModels)
	}

}

// GatewayModelItemResult 网关型号列表项
type GatewayModelItemResult struct {
	// 描述，自定义的型号描述
	Description string `json:"description,omitempty"`
	// 领域，DJI 文档指定
	Domain int `json:"domain"`
	// 数据库中的ID
	ID uint `json:"id"`
	// 型号名称，DJI 文档中收录的标准名称
	Name string `json:"name"`
	// 子型号，DJI 文档指定
	SubType int `json:"sub_type"`
	// 主型号，DJI 文档指定
	Type int `json:"type"`
}

func (r *ModelRouter) getGatewayModels(c *fiber.Ctx) error {
	// Define the hardcoded gateway models data
	models := []GatewayModelItemResult{
		{
			ID:          1,
			Name:        "DJI 带屏遥控器行业版",
			Description: "搭配 Matrice 300 RTK",
			Domain:      2,
			Type:        56,
			SubType:     0,
		},
		{
			ID:          2,
			Name:        "DJI RC Plus",
			Description: "搭配Matrice 350 RTK,Matrice 300 RTK,Matrice 30/30T",
			Domain:      2,
			Type:        119,
			SubType:     0,
		},
		{
			ID:          3,
			Name:        "DJI RC Plus 2",
			Description: "搭配>DJI Matrice 4 系列",
			Domain:      2,
			Type:        174,
			SubType:     0,
		},
		{
			ID:          4,
			Name:        "DJI RC Pro 行业版",
			Description: "搭配 Mavic 3 行业系列",
			Domain:      2,
			Type:        144,
			SubType:     0,
		},
		{
			ID:      5,
			Name:    "大疆机场",
			Domain:  3,
			Type:    1,
			SubType: 0,
		},
		{
			ID:      6,
			Name:    "大疆机场2",
			Domain:  3,
			Type:    2,
			SubType: 0,
		},
		{
			ID:      7,
			Name:    "大疆机场3",
			Domain:  3,
			Type:    3,
			SubType: 0,
		},
	}

	return c.JSON(Success(models))
}

// DroneModelItemResult
type DroneModelItemResult struct {
	Description string `json:"description,omitempty"`
	// 领域
	Domain int `json:"domain"`
	// 对应的网关描述
	GatewayDescription string `json:"gateway_description,omitempty"`
	// 对应的网关ID
	GatewayID uint `json:"gateway_id"`
	// 对应的网关名称
	GatewayName string `json:"gateway_name"`
	// 可搭载云台
	Gimbals []Gimbal `json:"gimbals,omitempty"`
	// 型号ID
	ID uint `json:"id"`
	// 型号名称
	Name string `json:"name"`
	// 自类型
	SubType int `json:"sub_type"`
	// 主类型
	Type int `json:"type"`
}

// 云台型号列表项
type Gimbal struct {
	ID uint `json:"id"`
	// 描述
	Description string `json:"description"`
	// 领域
	Domain int `json:"domain"`
	// 相机位置索引
	Gimbalindex int `json:"gimbalindex"`
	// 云台名称
	Name string `json:"name"`
	// 产品线名称，FPV相机、云台相机和机场相机
	Product string `json:"product"`
	// 子型号
	SubType int `json:"sub_type"`
	// 型号
	Type int `json:"type"`
}

func (r *ModelRouter) getDroneModels(c *fiber.Ctx) error {
	// Define hardcoded drone models data
	models := []DroneModelItemResult{
		{
			ID:          1,
			Name:        "Mavic 3 行业系列（M3E 相机）",
			Description: "大疆 M300 M3E 无人机",
			Domain:      0,
			Type:        77,
			SubType:     0,
			GatewayName: "DJI RC Pro",
			GatewayID:   1,
		},
		{
			ID:          2,
			Name:        "Mavic 3 行业系列（M3T 相机）",
			Description: "大疆 M3T 无人机",
			Domain:      0,
			Type:        77,
			SubType:     1,
			GatewayName: "DJI RC Pro",
			GatewayID:   1,
		},
		{
			ID:          3,
			Name:        "Matrice 350 RTK",
			Description: "大疆 M350 RTK 无人机",
			Domain:      0,
			Type:        89,
			SubType:     0,
			GatewayName: "DJI 带屏遥控器行业版",
			GatewayID:   1,
			Gimbals: []Gimbal{
				{
					ID:          1,
					Name:        "H20/H20T",
					Description: "大疆 H20T 云台",
				},
				{
					ID:          2,
					Name:        "H20N",
					Description: "大疆 P1 云台",
				},
				{
					ID:          3,
					Name:        "H30/H30T",
					Description: "大疆 H20T 云台",
				},
			},
		},
		{
			ID:          4,
			Name:        "DJI Matrice 4 系列（M4E 相机）",
			Description: "大疆 M4E 无人机",
			Domain:      0,
			Type:        99,
			SubType:     0,
			GatewayName: "DJI RC Plus",
			GatewayID:   1,
		},
	}

	return c.JSON(Success(models))
}

func (r *ModelRouter) getGimbalModels(c *fiber.Ctx) error {
	// Define hardcoded gimbal models data
	models := []Gimbal{
		{
			ID:          1,
			Product:     "飞行器FPV",
			Name:        "Matrice 350 RTK FPV",
			Description: "大疆 M350 RTK FPV 云台",
			Domain:      1,
			Type:        39,
			SubType:     0,
			Gimbalindex: 7,
		},
		{
			ID:          2,
			Product:     "相机",
			Name:        "禅思 H20",
			Description: "大疆 H20 云台，位于飞行器左舷侧",
			Domain:      1,
			Type:        42,
			SubType:     0,
			Gimbalindex: 0,
		},
		{
			ID:          3,
			Product:     "相机",
			Name:        "禅思 H20",
			Description: "大疆 H20 云台，位于飞行器右舷侧",
			Domain:      1,
			Type:        42,
			SubType:     0,
			Gimbalindex: 1,
		},
		{
			ID:          4,
			Product:     "相机",
			Name:        "禅思 H20",
			Description: "大疆 H20 云台，位于飞行器上侧",
			Domain:      1,
			Type:        42,
			SubType:     0,
			Gimbalindex: 2,
		},
		{
			ID:          5,
			Product:     "相机",
			Name:        "DJI Matrice 4E Camera",
			Description: "大疆 M4E 云台",
			Domain:      1,
			Type:        88,
			SubType:     0,
			Gimbalindex: 0,
		},
		{
			ID:          6,
			Product:     "机场相机",
			Name:        "DJI Dock 舱外相机",
			Description: "大疆机场1相机",
			Domain:      1,
			Type:        165,
			SubType:     0,
			Gimbalindex: 7,
		},
	}

	return c.JSON(Success(models))
}

type PayloadItemResult struct {
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	ID          int64  `json:"id"`
	Name        string `json:"name"`
}

func (r *ModelRouter) getPaylodModels(c *fiber.Ctx) error {
	// Define hardcoded payload models data
	models := []PayloadItemResult{
		{
			ID:          1,
			Name:        "DJI Mavic 3 行业系列 RTK 模块",
			Category:    "RTK模块",
			Description: "DJI Mavic 3 行业系列 RTK 模块适配 DJI Mavic 3 行业系列机型，结合网络 RTK 或自定义网络 RTK 服务，或通过 D-RTK 2 移动站，提供高精度厘米级位置定位功能。",
		},
		{
			ID:          2,
			Name:        "DJI Mavic 3 行业系列喊话器",
			Category:    "喊话器",
			Description: "远程传递声音，让应急搜救等任务更高效。可储存多条语音，并支持自动循环播放",
		},
	}

	return c.JSON(Success(models))
}
