package services

import (
	"achievements-uas/repositories"
	"achievements-uas/utils"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"time"
)

type AuthService struct {
	AuthRepo      *repository.AuthRepository
	RolePermRepo  *repository.RolePermissionRepository
}

func NewAuthService(
	authRepo *repository.AuthRepository,
	rolePermRepo *repository.RolePermissionRepository,
) *AuthService {
	return &AuthService{
		AuthRepo:     authRepo,
		RolePermRepo: rolePermRepo,
	}
}

// ========================================================
// LOGIN
// ========================================================
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

	// ============= GET PERMISSIONS ============
	perms, err := s.RolePermRepo.GetPermissionsByRole(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed fetch permissions"})
	}

	// ============= GENERATE TOKEN ============
	accessToken, err := utils.GenerateAccessToken(user, perms)
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
			"role_id":     user.RoleID,
			"permissions": perms,
		},
	})
}

// ========================================================
// REFRESH TOKEN
// ========================================================
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

	// PERMISSIONS DIAMBIL DARI RolePermissionRepository
	perms, err := s.RolePermRepo.GetPermissionsByRole(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed fetch permissions"})
	}

	accessToken, err := utils.GenerateAccessToken(user, perms)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed generate token"})
	}

	return c.JSON(fiber.Map{
		"status":       "success",
		"access_token": accessToken,
	})
}

// ========================================================
// LOGOUT
// ========================================================
func (s *AuthService) Logout(c *fiber.Ctx) error {
	token := c.Locals("token").(string)
	exp := time.Now().Add(15 * time.Minute)

	utils.BlacklistToken(token, exp)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "logged out",
	})
}

// ========================================================
// PROFILE
// ========================================================
func (s *AuthService) Profile(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*utils.JWTClaims)

	user, err := s.AuthRepo.GetProfile(claims.ID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	perms, err := s.RolePermRepo.GetPermissionsByRole(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed load permissions"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"role_id":     user.RoleID,
			"permissions": perms,
		},
	})
}
