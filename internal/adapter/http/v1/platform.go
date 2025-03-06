package v1

import (
	api "github.com/dronesphere/api/http/v1"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type PlatformRouter struct {
	l *slog.Logger
}

func newPlatformRouter(handler fiber.Router, l *slog.Logger) {
	r := &PlatformRouter{
		l: l,
	}
	h := handler.Group("/platform")
	{
		h.Get("/", r.getInfo)
		h.Get("/params", r.getConnectionParams)
	}
}

// getInfo 获取平台信息
//
//	@Router			/platform [get]
//	@Summary		获取平台信息
//	@Description	获取平台信息，包含平台名称，工作空间，描述等
//	@Tags			平台
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=v1.PlatformResult}	"成功"
func (r *PlatformRouter) getInfo(c *fiber.Ctx) error {
	result := &api.PlatformResult{
		Platform:    "无人机搜索原型系统",
		Workspace:   "演示工作空间",
		WorkspaceID: "e3dea0f5-37f2-4d79-ae58-490af3228069",
		Desc:        "本工作空间为无人机搜索原型系统演示工作空间，用于展示无人机搜索原型系统的功能和特性。",
	}
	return c.JSON(Success(result))
}

// getConnectionParams 获取连接参数
//
//	@Router			/platform/params [get]
//	@Summary		获取连接参数
//	@Description	获取连接参数，包含设备上云模块连接参数，API 模块连接参数，WebSocket 模块连接参数等
//	@Tags			平台
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	v1.Response{data=v1.ConnectionParamsResult}	"成功"
func (r *PlatformRouter) getConnectionParams(c *fiber.Ctx) error {
	result := &api.ConnectionParamsResult{
		Thing: api.ThingParamResult{
			Host:     "tcp://47.245.40.222:1883",
			Username: "drone",
			Password: "drone",
		},
		API: api.APIParamResult{
			Host:  "https://192.168.1.112:10087",
			Token: "123456",
		},
		WS: api.WSParamResult{
			Host:  "ws://192.168.1.112:10088",
			Token: "123456",
		},
	}
	return c.JSON(Success(result))
}
