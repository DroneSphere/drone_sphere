package dto

// HostInfo 主机设备信息
type HostInfo struct {
	Latitude        float64 `json:"latitude"`         // 纬度
	Longitude       float64 `json:"longitude"`        // 经度
	Height          float64 `json:"height"`           // 椭球高度（单位：米）
	AttitudeHead    float64 `json:"attitude_head"`    // 设备朝向（单位：度）
	Elevation       float64 `json:"elevation"`        // 相对起飞高度（单位：米）
	HorizontalSpeed float64 `json:"horizontal_speed"` // 水平速度（单位：米/秒）
	VerticalSpeed   float64 `json:"vertical_speed"`   // 垂直速度（单位：米/秒）
}

// WSDeviceOsdPayload 设备遥感信息
type WSDeviceOsdPayload struct {
	WSCommon
	Data struct {
		Host HostInfo `json:"host"` // 主机设备信息
		SN   string   `json:"sn"`   // 设备序列号
	} `json:"data"`
}

// WSDeviceOnlinePayload 设备上线
type WSDeviceOnlinePayload struct {
	WSCommon
	Data map[string]interface{} `json:"data"` // 允许扩展字段
}

// WSDeviceOfflinePayload 设备下线消息
type WSDeviceOfflinePayload struct {
	WSCommon
	Data map[string]interface{} `json:"data"` // 允许扩展字段
}

// WSDeviceUpdateTopoPayload 设备拓扑更新消息
type WSDeviceUpdateTopoPayload struct {
	WSCommon
	Data map[string]interface{} `json:"data"` // 允许扩展字段
}
