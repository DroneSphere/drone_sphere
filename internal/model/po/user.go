package po

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Password string `json:"password"`
}

// TableName 指定 User 表名为 tb_users
func (u User) TableName() string {
	return "tb_users"
}
