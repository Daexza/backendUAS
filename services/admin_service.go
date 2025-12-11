package services

import (
	"achievements-uas/models"
	"achievements-uas/repositories"
	"achievements-uas/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"net/http"
)

type UserAdminService struct {
	Repo *repository.UserAdminRepository
}

func NewUserAdminService(repo *repository.UserAdminRepository) *UserAdminService {
	return &UserAdminService{Repo: repo}
}

// ----------------------------
// CREATE USER
// ----------------------------
func (s *UserAdminService) Create(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	hash, err := utils.HashPassword(body.Password)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed hash password"})
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: hash,
		FullName:     body.FullName,
		RoleID:       body.RoleID,
		IsActive:     true,
	}

	if err := s.Repo.CreateUser(user); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed create user"})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"status": "success", "data": user})
}

// ----------------------------
// GET ALL USERS
// ----------------------------
func (s *UserAdminService) GetAll(c *fiber.Ctx) error {
	users, err := s.Repo.GetAllUsers()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed get users"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": users})
}

// ----------------------------
// GET USER BY ID
// ----------------------------
func (s *UserAdminService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := s.Repo.GetByID(id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": user})
}

// ----------------------------
// UPDATE USER
// ----------------------------
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
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	user := &models.User{
		ID:       id,
		Username: body.Username,
		Email:    body.Email,
		FullName: body.FullName,
		RoleID:   body.RoleID,
		IsActive: body.IsActive,
	}

	if err := s.Repo.UpdateUser(user); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed update user"})
	}

	return c.JSON(fiber.Map{"status": "success", "data": user})
}

// ----------------------------
// DELETE USER
// ----------------------------
func (s *UserAdminService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := s.Repo.DeleteUser(id); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed delete user"})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "user deleted"})
}

// ----------------------------
// RESET PASSWORD
// ----------------------------
func (s *UserAdminService) UpdatePassword(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil || body.NewPassword == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "password required"})
	}

	if _, err := s.Repo.GetByID(id); err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	hash, err := utils.HashPassword(body.NewPassword)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed hash password"})
	}

	if err := s.Repo.UpdatePassword(id, hash); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed update password"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "password updated"})
}

// ----------------------------
// ASSIGN ROLE
// ----------------------------
func (s *UserAdminService) AssignRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		RoleID string `json:"role_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.RoleID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "role_id required"})
	}

	if err := s.Repo.AssignRole(id, body.RoleID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed assign role"})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "role updated"})
}

// ----------------------------
// CREATE STUDENT PROFILE
// ----------------------------
func (s *UserAdminService) CreateStudentProfile(c *fiber.Ctx) error {
	var body models.Student
	body.ID = uuid.New().String()
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	if err := s.Repo.CreateStudent(&body); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed create student profile"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": body})
}

// ----------------------------
// CREATE LECTURER PROFILE
// ----------------------------
func (s *UserAdminService) CreateLecturerProfile(c *fiber.Ctx) error {
	var body models.Lecturer
	body.ID = uuid.New().String()
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	if err := s.Repo.CreateLecturer(&body); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed create lecturer profile"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": body})
}

// ----------------------------
// SET ADVISOR
// ----------------------------
func (s *UserAdminService) SetAdvisor(c *fiber.Ctx) error {
	studentID := c.Params("id")
	var body struct {
		AdvisorID string `json:"advisor_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.AdvisorID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "advisor_id required"})
	}

	if err := s.Repo.SetAdvisor(studentID, body.AdvisorID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed set advisor"})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "advisor set"})
}
