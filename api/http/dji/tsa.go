package dji

// DeviceTopoRsp 设备拓扑响应
type DeviceTopoRsp struct {
	// Hosts 飞机设备拓扑集合
	Hosts []TopoHostDeviceRsp `json:"hosts"`
	// Gateways 网关设备拓扑集合，json key 为 parents
	Gateways []TopoGatewayDeviceRsp `json:"parents"`
}

// TopoHostDeviceRsp 飞机设备拓扑集合
type TopoHostDeviceRsp struct {
	DeviceCallsign string         `json:"device_callsign"`
	DeviceModel    DeviceModelRsp `json:"device_model"`
	IconUrls       struct {
		NormalIconURL   string `json:"normal_icon_url"`
		SelectedIconURL string `json:"selected_icon_url"`
	} `json:"icon_urls"`
	OnlineStatus bool   `json:"online_status"`
	SN           string `json:"sn"`
	UserCallsign string `json:"user_callsign"`
	UserID       string `json:"user_id"`
}

// TopoGatewayDeviceRsp 网关设备拓扑集合
type TopoGatewayDeviceRsp struct {
	DeviceCallsign string         `json:"device_callsign"`
	DeviceModel    DeviceModelRsp `json:"device_model"`
	IconUrls       struct {
		NormalIconURL   string `json:"normal_icon_url"`
		SelectedIconURL string `json:"selected_icon_url"`
	} `json:"icon_urls"`
	OnlineStatus bool   `json:"online_status"`
	SN           string `json:"sn"`
	UserCallsign string `json:"user_callsign"`
	UserID       string `json:"user_id"`
}

// DeviceModelRsp 设备模型响应
type DeviceModelRsp struct {
	Key     string `json:"key"`
	Domain  string `json:"domain"`
	Type    string `json:"type"`
	SubType string `json:"sub_type"`
}
