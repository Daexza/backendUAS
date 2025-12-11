package middleware

import (
	"strings"
	"achievements-uas/utils"
	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
		}

		parts := strings.Fields(auth)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token format"})
		}
		token := parts[1]

		if IsTokenBlacklisted(token) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token revoked"})
		}

		claims, err := utils.ParseAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		c.Locals("claims", claims)
		c.Locals("token", token)
		return c.Next()
	}
}
