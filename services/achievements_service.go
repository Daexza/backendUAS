package services

import (
	"achievements-uas/app/models"
	"achievements-uas/app/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type AchievementService struct {
	MongoRepo *repository.AchievementMongoRepository
	PgRepo    *repository.AchievementPGRepository
}

func NewAchievementService(
	mongoRepo *repository.AchievementMongoRepository,
	pgRepo *repository.AchievementPGRepository,
) *AchievementService {
	return &AchievementService{
		MongoRepo: mongoRepo,
		PgRepo:    pgRepo,
	}
}


// ==============================================
// LIST (filtered by role, student sees only theirs)
// ==============================================
func (s *AchievementService) List(c *fiber.Ctx) error {
	// Filter logic bisa ditambah berdasarkan role JWT
	// Untuk contoh, return semua
	return c.JSON(fiber.Map{"status": "success", "data": "list achievements (implement filter)"})
}

// ==============================================
// GET BY ID
// ==============================================
func (s *AchievementService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	a, err := s.MongoRepo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": a})
}

// ==============================================
// CREATE (Mahasiswa)
// ==============================================
func (s *AchievementService) Create(c *fiber.Ctx) error {
	var a models.Achievement
	if err := c.BodyParser(&a); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	a.ID = primitive.NewObjectID()
	a.Status = "draft"
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	if err := s.MongoRepo.Create(&a); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed create achievement"})
	}

	ref := &models.AchievementReference{
		ID:                 uuid.New().String(),
		StudentID:          a.StudentID,
		MongoAchievementID: a.ID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.PgRepo.Create(ref); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed create reference"})
	}

	return c.Status(201).JSON(fiber.Map{"status": "success", "data": a})
}

// ==============================================
// UPDATE (Mahasiswa draft)
// ==============================================
func (s *AchievementService) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	a, err := s.MongoRepo.GetByID(id)
	if err != nil || a.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "achievement not editable"})
	}

	if err := c.BodyParser(a); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	a.UpdatedAt = time.Now()
	if err := s.MongoRepo.Update(a); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update achievement"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": a})
}

// ==============================================
// DELETE (Mahasiswa draft → soft delete)
// ==============================================
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	a, err := s.MongoRepo.GetByID(id)
	if err != nil || a.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "achievement not deletable"})
	}

	if err := s.MongoRepo.SoftDelete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed soft delete"})
	}

	// update PG reference
	if err := s.PgRepo.UpdateStatusByMongoID(a.ID.Hex(), "deleted", "", ""); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update reference"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "achievement deleted"})
}

// ==============================================
// SUBMIT (Mahasiswa draft → submitted)
// ==============================================
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id := c.Params("id")
	a, err := s.MongoRepo.GetByID(id)
	if err != nil || a.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "achievement not submittable"})
	}

	a.Status = "submitted"
	a.UpdatedAt = time.Now()
	if err := s.MongoRepo.Update(a); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update achievement"})
	}

	if err := s.PgRepo.UpdateStatusByMongoID(a.ID.Hex(), "submitted", "", ""); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update reference"})
	}

	// Add history
	h := &models.AchievementHistory{
		AchievementID: a.ID.Hex(),
		StudentID:     a.StudentID,
		Status:        "submitted",
		ChangedBy:     a.StudentID,
		Notes:         "",
	}
	_ = s.PgRepo.AddHistory(h)

	return c.JSON(fiber.Map{"status": "success", "data": a})
}

// ==============================================
// VERIFY (Dosen → submitted → verified)
// ==============================================
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		LecturerID string `json:"lecturer_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.LecturerID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "lecturer_id required"})
	}

	a, err := s.MongoRepo.GetByID(id)
	if err != nil || a.Status != "submitted" {
		return c.Status(400).JSON(fiber.Map{"error": "achievement not verifiable"})
	}

	a.Status = "verified"
	a.UpdatedAt = time.Now()
	if err := s.MongoRepo.Update(a); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update achievement"})
	}

	if err := s.PgRepo.UpdateStatusByMongoID(a.ID.Hex(), "verified", body.LecturerID, ""); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update reference"})
	}

	// Add history
	h := &models.AchievementHistory{
		AchievementID: a.ID.Hex(),
		StudentID:     a.StudentID,
		Status:        "verified",
		ChangedBy:     body.LecturerID,
	}
	_ = s.PgRepo.AddHistory(h)

	return c.JSON(fiber.Map{"status": "success", "data": a})
}

// ==============================================
// REJECT (Dosen → submitted → rejected)
// ==============================================
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		LecturerID string `json:"lecturer_id"`
		Notes      string `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil || body.LecturerID == "" || body.Notes == "" {
		return c.Status(400).JSON(fiber.Map{"error": "lecturer_id and notes required"})
	}

	a, err := s.MongoRepo.GetByID(id)
	if err != nil || a.Status != "submitted" {
		return c.Status(400).JSON(fiber.Map{"error": "achievement not rejectable"})
	}

	a.Status = "rejected"
	a.UpdatedAt = time.Now()
	if err := s.MongoRepo.Update(a); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update achievement"})
	}

	if err := s.PgRepo.UpdateStatusByMongoID(a.ID.Hex(), "rejected", body.LecturerID, body.Notes); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update reference"})
	}

	// Add history
	h := &models.AchievementHistory{
		AchievementID: a.ID.Hex(),
		StudentID:     a.StudentID,
		Status:        "rejected",
		ChangedBy:     body.LecturerID,
		Notes:         body.Notes,
	}
	_ = s.PgRepo.AddHistory(h)

	return c.JSON(fiber.Map{"status": "success", "data": a})
}

// ==============================================
// HISTORY
// ==============================================
func (s *AchievementService) History(c *fiber.Ctx) error {
	id := c.Params("id")
	history, err := s.PgRepo.GetHistoryByMongoID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get history"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": history})
}

// ==============================================
// ATTACHMENTS (Upload file)
// ==============================================
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	// implement sesuai kebutuhan storage, update field Attachments di Mongo
	return c.JSON(fiber.Map{"status": "success", "message": "upload placeholder"})
}
