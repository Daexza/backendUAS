package services

import (
	"context"
	"strconv"

	"achievements-uas/app/models"
	"achievements-uas/app/repository"
	"achievements-uas/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAdminService struct {
	AdminRepo    *repository.AdminRepository
	RoleRepo     *repository.RoleRepository
	RolePermRepo *repository.RolePermissionRepository

	// ===== FR-010 =====
	AchPgRepo    *repository.AchievementPostgresRepository
	AchMongoRepo *repository.AchievementMongoRepository
}

// ==============================================
// CONSTRUCTOR
// ==============================================
func NewAdminService(
	adminRepo *repository.AdminRepository,
	roleRepo *repository.RoleRepository,
	rolePermRepo *repository.RolePermissionRepository,
	achPgRepo *repository.AchievementPostgresRepository,
	achMongoRepo *repository.AchievementMongoRepository,
) *UserAdminService {
	return &UserAdminService{
		AdminRepo:    adminRepo,
		RoleRepo:     roleRepo,
		RolePermRepo: rolePermRepo,
		AchPgRepo:    achPgRepo,
		AchMongoRepo: achMongoRepo,
	}
}

// ==============================================
// CREATE USER
// ==============================================
func (s *UserAdminService) Create(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	if _, err := s.RoleRepo.FindByID(body.RoleID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid role_id"})
	}

	hash, _ := utils.HashPassword(body.Password)

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: hash,
		FullName:     body.FullName,
		RoleID:       body.RoleID,
		IsActive:     true,
	}

	if err := s.AdminRepo.CreateUser(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed create user"})
	}

	return c.Status(201).JSON(fiber.Map{"status": "success", "data": user})
}

// ==============================================
// GET ALL USERS
// ==============================================
func (s *UserAdminService) GetAll(c *fiber.Ctx) error {
	users, err := s.AdminRepo.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get users"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": users})
}

// ==============================================
// GET USER BY ID
// ==============================================
func (s *UserAdminService) GetByID(c *fiber.Ctx) error {
	user, err := s.AdminRepo.GetByID(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": user})
}

// ==============================================
// UPDATE USER
// ==============================================
func (s *UserAdminService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	if _, err := s.RoleRepo.FindByID(body.RoleID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid role_id"})
	}

	user := &models.User{
		ID:       id,
		Username: body.Username,
		Email:    body.Email,
		FullName: body.FullName,
		RoleID:   body.RoleID,
		IsActive: body.IsActive,
	}

	if err := s.AdminRepo.UpdateUser(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update user"})
	}

	return c.JSON(fiber.Map{"status": "success", "data": user})
}

// ==============================================
// DELETE USER
// ==============================================
func (s *UserAdminService) Delete(c *fiber.Ctx) error {
	if err := s.AdminRepo.DeleteUser(c.Params("id")); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed delete user"})
	}
	return c.JSON(fiber.Map{"status": "success"})
}

// ==============================================
// UPDATE USER PASSWORD (ADMIN)
// ==============================================
func (s *UserAdminService) UpdatePassword(c *fiber.Ctx) error {
	userID := c.Params("id")

	var body struct {
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&body); err != nil || body.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "new_password is required",
		})
	}

	// hash password baru
	hash, err := utils.HashPassword(body.NewPassword)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed hash password",
		})
	}

	// update ke DB
	if err := s.AdminRepo.UpdatePassword(userID, hash); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed update password",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "password updated",
	})
}

//
// =====================================================
// FR-010: VIEW ALL ACHIEVEMENTS (ADMIN)
// =====================================================
func (s *UserAdminService) GetAllAchievements(c *fiber.Ctx) error {
	ctx := context.Background()

	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	offset := (page - 1) * limit

	refs, err := s.AchPgRepo.GetAll(ctx, status, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed load references"})
	}

	var ids []primitive.ObjectID
	for _, r := range refs {
		if oid, err := primitive.ObjectIDFromHex(r.MongoAchievementID); err == nil {
			ids = append(ids, oid)
		}
	}

	data, err := s.AchMongoRepo.FindByIDs(ctx, ids)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed load achievements"})
	}

	return c.JSON(fiber.Map{"status": "success", "data": data})
}

//
// ================= STUDENTS (ADMIN) =================
//

func (s *UserAdminService) GetAllStudents(c *fiber.Ctx) error {
	data, err := s.AdminRepo.GetAllStudents()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get students"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": data})
}

func (s *UserAdminService) GetStudentByID(c *fiber.Ctx) error {
	data, err := s.AdminRepo.GetStudentByID(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": data})
}

func (s *UserAdminService) GetStudentAchievements(c *fiber.Ctx) error {
	data, err := s.AdminRepo.GetStudentAchievements(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get achievements"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": data})
}

func (s *UserAdminService) SetAdvisor(c *fiber.Ctx) error {
	var body struct {
		AdvisorID string `json:"advisor_id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "advisor_id required"})
	}

	if err := s.AdminRepo.SetAdvisor(c.Params("id"), body.AdvisorID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed set advisor"})
	}

	return c.JSON(fiber.Map{"status": "success"})
}

//
// ================= LECTURERS (ADMIN) =================
//

func (s *UserAdminService) GetAllLecturers(c *fiber.Ctx) error {
	data, err := s.AdminRepo.GetAllLecturers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get lecturers"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": data})
}

func (s *UserAdminService) GetLecturerAdvisees(c *fiber.Ctx) error {
	data, err := s.AdminRepo.GetLecturerAdvisees(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get advisees"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": data})
}
