package po

import "time"

type User struct {
	ID          uint      `json:"user_id" gorm:"primaryKey;column:user_id"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime"`
	State       int       `json:"state" gorm:"default:0"` // -1: deleted, 0: active
	Username    string    `json:"username" gorm:"unique"`
	Email       string    `json:"email"`
	Avatar      string    `json:"avatar"`
	Password    string    `json:"password"`
}

// TableName 指定 User 表名为 tb_users
func (u User) TableName() string {
	return "tb_users"
}
