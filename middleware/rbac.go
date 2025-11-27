package middleware

import "github.com/gofiber/fiber/v2"

func AllowRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("role")

		allowed := false
		for _, r := range roles {
			if r == userRole {
				allowed = true
				break
			}
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "Akses ditolak",
			})
		}

		return c.Next()
	}
}
