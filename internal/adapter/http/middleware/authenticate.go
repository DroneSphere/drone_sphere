package middleware

import (
	"github.com/dronesphere/internal/pkg/token"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strings"
)

const UserClaimsKey = "claims"

func Authenticate(c *fiber.Ctx) error {
	// 获取 Authorization 头
	auth := c.Get("Authorization")
	if auth == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "missing Authorization header",
		})
	}
	t := strings.Split(auth, " ")[1]
	if t == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "missing token",
		})
	}
	// 验证 token
	claims, err := token.ValidateToken(t)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid token",
		})
	}
	// 将 claims 存入上下文
	c.Locals(UserClaimsKey, claims)

	return c.Next()
}
