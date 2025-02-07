package v1

type LoginRequest struct {
	Username string `json:"username" example:"admin" binding:"required"`
	Password string `json:"password" example:"admin" binding:"required"`
	// SN 遥控器 SN，仅 Pilot 端登录时需要提供
	SN string `json:"sn" example:"123456"`
}
