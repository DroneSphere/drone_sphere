package event

const (
	UserLoginSuccessEvent = "user.login.success"
)

type UserLoginSuccess struct {
	Username string
	SN       string
}
