package v1

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Avatar   string `json:"avatar"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" example:"admin" binding:"required"`
	// SN 遥控器 SN，仅 Pilot 端登录时需要提供
	SN string `json:"sn" example:"123456"`
}

type LoginResult struct {
	Token     string          `json:"token"`
	User      UserResult      `json:"user"`
	Workspace WorkspaceResult `json:"workspace"`
}

type UserResult struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

type WorkspaceResult struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
