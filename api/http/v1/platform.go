package v1

type PlatformResult struct {
	Platform    string `json:"platform"`
	Workspace   string `json:"workspace"`
	WorkspaceID string `json:"workspace_id"`
	Desc        string `json:"desc"`
}

type ConnectionParamsResult struct {
	Thing ThingParamResult `json:"thing"`
	API   APIParamResult   `json:"api"`
	WS    WSParamResult    `json:"ws"`
}

// ThingParamResult 设备上云模块连接参数
type ThingParamResult struct {
	Host     string `json:"host"` // MQTT 服务器地址, 格式为 tcp://host:port
	Username string `json:"username"`
	Password string `json:"password"`
}

// APIParamResult API 模块连接参数
type APIParamResult struct {
	Host  string `json:"host"` // HTTP 服务器地址, 格式为 http://host:port
	Token string `json:"token"`
}

// WSParamResult WebSocket 模块连接参数
type WSParamResult struct {
	Host  string `json:"host"` // WebSocket 服务器地址, 格式为 ws://host:port
	Token string `json:"token"`
}
