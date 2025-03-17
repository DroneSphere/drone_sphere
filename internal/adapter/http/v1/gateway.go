package v1

import (
	"log/slog"

	"github.com/asaskevich/EventBus"
	"github.com/dronesphere/internal/model/po"
	"github.com/dronesphere/internal/pkg/misc"
	"github.com/gofiber/fiber/v2"
)

type GatewayRouter struct {
	eb EventBus.Bus
	l  *slog.Logger
}

func NewGatewayRouter(handler fiber.Router, eb EventBus.Bus, l *slog.Logger) {
	r := &GatewayRouter{
		eb: eb,
		l:  l,
	}
	h := handler.Group("/gateway")
	{
		h.Get("/", r.getAll)
	}

}

func (r *GatewayRouter) getAll(c *fiber.Ctx) error {
	type GatewayItemResult struct {
		ID uint `json:"id"`
		// 呼号，网关设备呼号
		Callsign string `json:"callsign,omitempty"`
		// 描述
		Description string `json:"description,omitempty"`
		// 型号信息
		Model po.GatewayModel `json:"model"`
		// 序列号，设备唯一序列号
		SN string `json:"sn"`
		// 状态，-1 为未知，0 为离线，1 为在线
		Status int `json:"status"`
		// 当前用户ID，登录该设备的用户ID
		UserID uint `json:"user_id"`
		// 当前用户名，登录该设备的用户名
		Username string `json:"username"`
	}

	model := po.GatewayModel{
		BaseModel: misc.BaseModel{
			ID: 1,
		},
		Name:        "DJI RC Pro 行业版",
		Description: "",
		Domain:      2,
		Type:        144,
		SubType:     0,
	}

	items := []GatewayItemResult{
		{
			ID:          1,
			Callsign:    "一号遥控器",
			Description: "一号描述",
			Model:       model,
			SN:          "SN123",
			Status:      1,
			UserID:      1,
			Username:    "测试用户1",
		},
		{
			ID:          2,
			Callsign:    "二号遥控器",
			Description: "二号描述",
			Model:       model,
			SN:          "0987654321",
			Status:      0,
			UserID:      2,
			Username:    "测试用户2",
		},
	}

	return c.JSON(Success(items))
}
