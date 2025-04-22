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

// 用户列表响应结构
type UserListResult struct {
	Users []UserResult `json:"users"`
	Total int64        `json:"total"` // 总用户数
}

// 修改密码请求结构
type ChangePasswordRequest struct {
	UserID      uint   `json:"userId" binding:"required"`      // 用户ID
	OldPassword string `json:"oldPassword" binding:"required"` // 旧密码
	NewPassword string `json:"newPassword" binding:"required"` // 新密码
}

// 创建用户请求结构
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Email    string `json:"email" binding:"required"`    // 邮箱
	Avatar   string `json:"avatar"`                      // 头像URL
	Password string `json:"password" binding:"required"` // 密码
}

type WorkspaceResult struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
