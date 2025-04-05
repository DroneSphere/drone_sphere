package event

const (
	GatewayOnlineEvent     = "gateway.online"     // 网关上线事件
	GatewayOfflineEvent    = "gateway.offline"    // 网关离线事件
	GatewayUpdateTopoEvent = "gateway.updateTopo" // 网关拓扑更新事件
)

// GatewayEventPayload 网关事件基础载荷
type GatewayEventPayload struct {
	SN         string         `json:"sn"`         // 网关序列号
	Timestamp  int64          `json:"timestamp"`  // 事件时间戳
	Properties map[string]any `json:"properties"` // 原始属性
}

// GatewayOnlinePayload 网关上线事件载荷
type GatewayOnlinePayload struct {
	GatewayEventPayload
	ModelID uint `json:"model_id"` // 网关型号ID
}

// GatewayOfflinePayload 网关离线事件载荷
type GatewayOfflinePayload struct {
	GatewayEventPayload
	Reason string `json:"reason"` // 离线原因
}

// GatewayUpdateTopoPayload 网关拓扑更新事件载荷
type GatewayUpdateTopoPayload struct {
	GatewaySN       string   `json:"gateway_sn"`       // 网关序列号
	ConnectedDrones []string `json:"connected_drones"` // 连接的无人机列表
}
