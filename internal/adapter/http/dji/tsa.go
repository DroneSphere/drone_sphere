package dji

import (
	"github.com/asaskevich/EventBus"
	api "github.com/dronesphere/api/http/dji"
	"github.com/dronesphere/internal/service"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type TSARouter struct {
	svc service.DroneSvc
	eb  EventBus.Bus
	l   *slog.Logger
}

func newTSARouter(handler fiber.Router, svc service.DroneSvc, eb EventBus.Bus, l *slog.Logger) {
	r := &TSARouter{
		svc: svc,
		eb:  eb,
		l:   l,
	}
	h := handler.Group("/")
	{
		h.Get("/manage/api/v1/workspaces/:"+workspaceIDParamKey+"/devices/topologies", r.getDeviceTopoByWorkspaceID)
	}
}

const workspaceIDParamKey = "workspace_id"

// getDeviceTopoByWorkspaceID 根据workspaceID获取设备拓扑
//
//	@Router			/manage/api/v1/workspaces/{workspace_id}/devices/topologies [get]
//	@Summary		获取设备拓扑列表
//	@Description	PILOT在首次上线后，会发送http请求去获取同一个工作空间下的所有设备列表及其拓扑，
//	@Description	服务端需要把整个设备列表发给PILOT。
//	@Description	同时，当接收到websocket指令通知设备online/offline/update的时候，
//	@Description	也是需要调用该接口进行请求设备拓扑列表进行更新。
//	@Tags			dji
//	@Accept			json
//	@Produce		json
//	@Param			workspace_id	path		string									true	"工作空间ID"
//	@Success		200				{object}	dji.Response{data=[]dji.DeviceTopoRsp}	"成功"
func (r *TSARouter) getDeviceTopoByWorkspaceID(c *fiber.Ctx) error {
	workspaceID := c.Params(workspaceIDParamKey)
	r.l.Info("getDeviceTopoByWorkspaceID", slog.Any("workspaceID", workspaceID))
	drones, rcs, err := r.svc.FetchDeviceTopo(c.Context(), workspaceID)
	if err != nil {
		r.l.Error("Failed to fetch device topo", slog.Any("err", err))
		return c.JSON(Fail(InternalError))
	}

	var hosts []api.TopoHostDeviceRsp
	for _, d := range drones {
		hosts = append(hosts, api.TopoHostDeviceRsp{
			SN: d.SN,
		})
	}
	var gateways []api.TopoGatewayDeviceRsp
	for _, rc := range rcs {
		gateways = append(gateways, api.TopoGatewayDeviceRsp{
			SN: rc.SN,
		})
	}
	rsp := &api.DeviceTopoRsp{
		Hosts:    hosts,
		Gateways: gateways,
	}
	return c.JSON(Success(rsp))

}
