package v1

// GatewayUpdateRequest 更新网关设备信息的请求
type GatewayUpdateRequest struct {
	Callsign    string `json:"callsign"`    // 呼号
	Description string `json:"description"` // 描述
}

// GatewayInfo 网关设备信息响应
type GatewayInfo struct {
	ID          uint   `json:"id"`          // ID
	Callsign    string `json:"callsign"`    // 呼号
	Description string `json:"description"` // 描述
}

// GatewayDetailInfo 网关设备详情响应
type GatewayDetailInfo struct {
	GatewayInfo
	DroneList []string `json:"drone_list"` // 关联的无人机列表
}
