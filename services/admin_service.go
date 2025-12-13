package services

import (
	"achievements-uas/app/models"
	"achievements-uas/app/repository"
	"achievements-uas/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserAdminService struct {
	AdminRepo     *repository.AdminRepository
	RoleRepo      *repository.RoleRepository
	RolePermRepo  *repository.RolePermissionRepository
}

// ==============================================
// CONSTRUCTOR
// ==============================================
func NewAdminService(
	adminRepo *repository.AdminRepository,
	roleRepo *repository.RoleRepository,
	rolePermRepo *repository.RolePermissionRepository,
) *UserAdminService {
	return &UserAdminService{
		AdminRepo:    adminRepo,
		RoleRepo:     roleRepo,
		RolePermRepo: rolePermRepo,
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

	// cek role valid
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

	// cek role valid
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
	return c.JSON(fiber.Map{"status": "success", "message": "user deleted"})
}

// ==============================================
// RESET PASSWORD
// ==============================================
func (s *UserAdminService) UpdatePassword(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&body); err != nil || body.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{"error": "password required"})
	}

	if _, err := s.AdminRepo.GetByID(id); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	hash, _ := utils.HashPassword(body.NewPassword)

	if err := s.AdminRepo.UpdatePassword(id, hash); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update password"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "password updated"})
}

// ==============================================
// ASSIGN ROLE
// ==============================================
func (s *UserAdminService) AssignRole(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		RoleID string `json:"role_id"`
	}

	if err := c.BodyParser(&body); err != nil || body.RoleID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "role_id required"})
	}

	// cek role valid
	if _, err := s.RoleRepo.FindByID(body.RoleID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid role_id"})
	}

	if err := s.AdminRepo.AssignRole(id, body.RoleID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed assign role"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "role updated"})
}

// ==============================================
// CREATE STUDENT PROFILE
// ==============================================
func (s *UserAdminService) CreateStudentProfile(c *fiber.Ctx) error {
	var body models.Student
	body.ID = uuid.New().String()

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	if err := s.AdminRepo.CreateStudent(&body); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed create student profile"})
	}

	return c.JSON(fiber.Map{"status": "success", "data": body})
}

// ==============================================
// CREATE LECTURER PROFILE
// ==============================================
func (s *UserAdminService) CreateLecturerProfile(c *fiber.Ctx) error {
	var body models.Lecturer
	body.ID = uuid.New().String()

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	if err := s.AdminRepo.CreateLecturer(&body); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed create lecturer profile"})
	}

	return c.JSON(fiber.Map{"status": "success", "data": body})
}

// ==============================================
// SET ADVISOR
// ==============================================
func (s *UserAdminService) SetAdvisor(c *fiber.Ctx) error {
	var body struct {
		AdvisorID string `json:"advisor_id"`
	}

	if err := c.BodyParser(&body); err != nil || body.AdvisorID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "advisor_id required"})
	}

	if err := s.AdminRepo.SetAdvisor(c.Params("id"), body.AdvisorID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed set advisor"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "advisor set"})
}
// ==============================================
// GET ALL STUDENTS
// ==============================================
func (s *UserAdminService) GetAllStudents(c *fiber.Ctx) error {
	list, err := s.AdminRepo.GetAllStudents()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get students"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": list})
}

// ==============================================
// GET STUDENT BY ID
// ==============================================
func (s *UserAdminService) GetStudentByID(c *fiber.Ctx) error {
	res, err := s.AdminRepo.GetStudentByID(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": res})
}

// ==============================================
// GET STUDENT ACHIEVEMENTS
// ==============================================
func (s *UserAdminService) GetStudentAchievements(c *fiber.Ctx) error {
	list, err := s.AdminRepo.GetStudentAchievements(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed load achievements"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": list})
}

// ==============================================
// GET ALL LECTURERS
// ==============================================
func (s *UserAdminService) GetAllLecturers(c *fiber.Ctx) error {
	list, err := s.AdminRepo.GetAllLecturers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get lecturers"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": list})
}

// ==============================================
// GET LECTURER ADVISEES
// ==============================================
func (s *UserAdminService) GetLecturerAdvisees(c *fiber.Ctx) error {
	list, err := s.AdminRepo.GetLecturerAdvisees(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get advisees"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": list})
}
