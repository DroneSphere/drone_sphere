package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/asaskevich/EventBus"
	v1 "github.com/dronesphere/api/http/v1"
	"github.com/dronesphere/internal/event"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
)

var eb = EventBus.New()

type MockUserSvc struct {
	mock.Mock
}

func (m *MockUserSvc) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func setup() (*fiber.App, *MockUserSvc) {
	app := fiber.New()
	mockSvc := new(MockUserSvc)
	newUserRouter(app, mockSvc, eb, slog.New(slog.NewTextHandler(os.Stdout, nil)))

	mockSvc.On("Login", "test", "test").Return("token", nil)

	return app, mockSvc
}

func TestUserRouter_login_Success(t *testing.T) {
	app, _ := setup()

	body, _ := sonic.Marshal(v1.LoginRequest{
		Username: "test",
		Password: "test",
	})
	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response.Msg)
	assert.Equal(t, "token", response.Data)
}

func TestUserRouter_login_SvcError(t *testing.T) {
	app, mockSvc := setup()
	mockSvc.On("Login", "error", "error").Return("", assert.AnError)

	body, _ := sonic.Marshal(v1.LoginRequest{
		Username: "error",
		Password: "error",
	})
	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestUserRouter_login_WithSN(t *testing.T) {
	app, _ := setup()

	testSN := "123456"
	body, _ := sonic.Marshal(v1.LoginRequest{
		Username: "test",
		Password: "test",
		SN:       testSN,
	})
	req := httptest.NewRequest("POST", "/user/login", bytes.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	_ = eb.Subscribe(event.RemoteControllerLoggedIn, func(ctx context.Context) {
		sn := ctx.Value("sn")
		assert.Equal(t, testSN, sn)
	})
}
