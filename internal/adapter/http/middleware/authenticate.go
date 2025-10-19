package middleware

import (
	"net/http"
	"strings"

	"github.com/dronesphere/internal/pkg/token"
	"github.com/gofiber/fiber/v2"
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

	// 检查Bearer格式
	authParts := strings.Split(auth, " ")
	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid authorization format",
		})
	}

	tokenString := authParts[1]
	if tokenString == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "missing token",
		})
	}

	// 验证 token
	claims, err := token.ValidateToken(tokenString)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid token",
		})
	}

	// 将 claims 存入上下文
	c.Locals(UserClaimsKey, claims)

	return c.Next()
}
