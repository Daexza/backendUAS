package services

import (
	"net/http"

	"achievements-uas/app/repository"
	"achievements-uas/utils"
	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	AuthRepo     *repository.AuthRepository
	RoleRepo     *repository.RoleRepository
	RolePermRepo *repository.RolePermissionRepository
}

func NewAuthService(
	authRepo *repository.AuthRepository,
	roleRepo *repository.RoleRepository,
	rolePermRepo *repository.RolePermissionRepository,
) *AuthService {
	return &AuthService{
		AuthRepo:     authRepo,
		RoleRepo:     roleRepo,
		RolePermRepo: rolePermRepo,
	}
}

//
// ======================= LOGIN =======================
//
func (s *AuthService) Login(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).
			JSON(fiber.Map{"error": "invalid input"})
	}

	user, err := s.AuthRepo.GetByUsernameOrEmail(body.Username)
	if err != nil {
		return c.Status(http.StatusUnauthorized).
			JSON(fiber.Map{"error": "invalid credentials"})
	}

	if !user.IsActive {
		return c.Status(http.StatusForbidden).
			JSON(fiber.Map{"error": "user inactive"})
	}

	if !utils.VerifyPassword(user.PasswordHash, body.Password) {
		return c.Status(http.StatusUnauthorized).
			JSON(fiber.Map{"error": "invalid credentials"})
	}

	// ===== ambil ROLE NAME =====
	roleName, err := s.RoleRepo.GetNameByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed load role"})
	}

	// ===== ambil PERMISSIONS =====
	perms, err := s.RolePermRepo.GetPermissionsByRole(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed fetch permissions"})
	}

	// ===== generate token =====
	accessToken, err := utils.GenerateAccessToken(user, roleName, perms)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed generate token"})
	}

	refreshToken, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed generate refresh token"})
	}

	return c.JSON(fiber.Map{
		"status":        "success",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"role":        roleName,
			"permissions": perms,
		},
	})
}

//
// ======================= REFRESH =======================
//
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	claims, err := utils.ParseRefreshToken(body.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	user, err := s.AuthRepo.GetProfile(claims.ID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "user not found"})
	}

	roleName, err := s.RoleRepo.GetNameByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed load role"})
	}

	perms, err := s.RolePermRepo.GetPermissionsByRole(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed fetch permissions"})
	}

	accessToken, err := utils.GenerateAccessToken(user, roleName, perms)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed generate token"})
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"access_token": accessToken,
	})
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
    // Ambil token dari Locals (yang sudah dibersihkan oleh middleware)
    tokenString, ok := c.Locals("token").(string)
    if !ok {
        return c.Status(500).JSON(fiber.Map{"error": "Token not found in context"})
    }

    claims := c.Locals("claims").(*utils.JWTClaims)
    
    // Gunakan waktu expiry dari token
    utils.BlacklistToken(tokenString, claims.ExpiresAt.Time)

    return c.JSON(fiber.Map{"message": "Logout success"})
}

func (s *AuthService) Profile(c *fiber.Ctx) error {
    // Ambil data dari Locals
    raw := c.Locals("claims")
    
    // Safe Assertion: Pastikan menggunakan pointer (*)
    claims, ok := raw.(*utils.JWTClaims)
    if !ok {
        return c.Status(500).JSON(fiber.Map{
            "error": "Internal Server Error",
            "message": "Failed to parse authentication claims",
        })
    }

    // Ambil user dari database menggunakan claims.ID
    user, err := s.AuthRepo.GetProfile(claims.ID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    return c.Status(200).JSON(fiber.Map{
        "status": "success",
        "data": fiber.Map{
            "user": user,
            "role": claims.Role,
        },
    })
}