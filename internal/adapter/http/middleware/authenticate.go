package middleware

import (
	"fmt"
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
	fmt.Printf("Authorization header: %s\n", auth)
	tmp := strings.Split(auth, " ")
	fmt.Printf("tmp: %v\n", tmp)
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
