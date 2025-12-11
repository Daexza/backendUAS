package middleware

import (
	"achievements-uas/utils"
	"github.com/gofiber/fiber/v2"
)

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claimsInterface := c.Locals("claims")
		if claimsInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing claims"})
		}
		claims := claimsInterface.(*utils.JWTClaims)

		for _, p := range claims.Permissions {
			if p == permission {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden: insufficient permissions"})
	}
}
