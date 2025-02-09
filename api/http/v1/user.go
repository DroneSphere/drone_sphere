package v1

type LoginRequest struct {
	Username string `json:"username" example:"admin" binding:"required"`
	Password string `json:"password" example:"admin" binding:"required"`
	// SN 遥控器 SN，仅 Pilot 端登录时需要提供
	SN string `json:"sn" example:"123456"`
}

type LoginResult struct {
	Token    string         `json:"token"`
	User     UserResult     `json:"user"`
	Platform PlatformResult `json:"platform"`
	Params   ParamsResult   `json:"params"`
}

type UserResult struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type PlatformResult struct {
	Platform  string `json:"platform"`
	Workspace string `json:"workspace"`
	Desc      string `json:"desc"`
}

type ParamsResult struct {
	MQTTHost     string `json:"mqtt_host"`
	MQTTUsername string `json:"mqtt_username"`
	MQTTPassword string `json:"mqtt_password"`
}
