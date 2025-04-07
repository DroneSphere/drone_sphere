package v1

import (
	"log/slog"

	api "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/configs"
	"github.com/gofiber/fiber/v2"
)

type PlatformRouter struct {
	l   *slog.Logger
	cfg *configs.Config // 添加配置对象
}

func newPlatformRouter(handler fiber.Router, l *slog.Logger, cfg *configs.Config) {
	r := &PlatformRouter{
		l:   l,
		cfg: cfg,
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
		Platform:    r.cfg.Platform.Name,
		Workspace:   r.cfg.Platform.Workspace,
		WorkspaceID: r.cfg.Platform.WorkspaceID,
		Desc:        r.cfg.Platform.Desc,
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
			Host:     r.cfg.Platform.Thing.Host,
			Username: r.cfg.Platform.Thing.Username,
			Password: r.cfg.Platform.Thing.Password,
		},
		API: api.APIParamResult{
			Host:  r.cfg.Platform.API.Host,
			Token: r.cfg.Platform.API.Token,
		},
		WS: api.WSParamResult{
			Host:  r.cfg.Platform.WS.Host,
			Token: r.cfg.Platform.WS.Token,
		},
	}
	return c.JSON(Success(result))
}
