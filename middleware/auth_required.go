package middleware

import (
	"strings"
	"achievements-uas/utils"
	"github.com/gofiber/fiber/v2"
)

// AuthRequired adalah middleware utama untuk cek Login & Blacklist
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil Header
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized", "message": "Missing token"})
		}

		// 2. Bersihkan Token String
		tokenString := strings.TrimSpace(strings.Replace(auth, "Bearer", "", 1))

		// 3. Cek Blacklist (Logout check)
		if utils.IsTokenBlacklisted(tokenString) {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized", "message": "Token revoked"})
		}

		// 4. Validasi JWT & Ambil Claims
		claims, err := utils.ParseAccessToken(tokenString)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized", "message": "Invalid or expired token"})
		}

		// 5. Simpan ke Locals untuk digunakan di Service/Next Middleware
		c.Locals("claims", claims)
		c.Locals("token", tokenString)

		return c.Next()
	}
}

// RoleRequired digunakan SETELAH AuthRequired untuk cek hak akses
func RoleRequired(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals("claims").(*utils.JWTClaims)
		if !ok {
			return c.Status(500).JSON(fiber.Map{"error": "Internal Server Error", "message": "Claims not found"})
		}

		// Cek apakah role user ada di daftar yang diizinkan
		for _, role := range allowedRoles {
			if claims.Role == role {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error":   "Forbidden",
			"message": "Anda tidak memiliki akses (Role: " + claims.Role + ")",
		})
	}
}