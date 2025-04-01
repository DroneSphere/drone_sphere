package po

type User struct {
	ID          uint   `json:"user_id" gorm:"primaryKey"`
	CreatedTune int64  `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTune int64  `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedTune int64  `json:"deleted_time" gorm:"autoDeleteTime"`
	State       int    `json:"state" gorm:"default:0"` // -1: deleted, 0: active
	Username    string `json:"username" gorm:"unique"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
	Password    string `json:"password"`
}

// TableName 指定 User 表名为 tb_users
func (u User) TableName() string {
	return "tb_users"
}
