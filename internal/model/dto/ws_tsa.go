package dto

// WSDeviceOnlinePayload 设备上线消息
type WSDeviceOnlinePayload struct {
	WSCommon
	Data struct {
		SN          string `json:"sn"`           // 设备序列号
		ConnectedSN string `json:"connected_sn"` // 连接的设备序列号（用于网关连接无人机）
		ProductType int    `json:"product_type"` // 产品类型
	} `json:"data"`
}

// WSDeviceOfflinePayload 设备下线消息
type WSDeviceOfflinePayload struct {
	WSCommon
	Data struct {
		SN             string `json:"sn"`              // 设备序列号
		DisconnectedSN string `json:"disconnected_sn"` // 断开连接的设备序列号
		ProductType    int    `json:"product_type"`    // 产品类型
	} `json:"data"`
}

// WSDeviceUpdateTopoPayload 设备拓扑更新消息
type WSDeviceUpdateTopoPayload struct {
	WSCommon
	Data struct {
		GatewaySN       string   `json:"gateway_sn"`       // 网关序列号
		ConnectedDrones []string `json:"connected_drones"` // 连接的无人机列表
	} `json:"data"`
}
