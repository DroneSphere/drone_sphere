package entity

import "time"

type User struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Avatar      string    `json:"avatar"`
	Password    string    `json:"password"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
}

func NewUser(username, email, avatar, password string) User {
	return User{
		Username: username,
		Email:    email,
		Avatar:   avatar,
		Password: password,
	}
}

func (u *User) IsPasswordMatch(password string) bool {
	return u.Password == password
}
