package services

import (
    "achievements-uas/app/models"
    "achievements-uas/app/repository"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type StudentService struct {
    StudentRepo     *repository.StudentRepository
    AchievementRepo *repository.AchievementMongoRepository
}

func NewStudentService(studentRepo *repository.StudentRepository, achRepo *repository.AchievementMongoRepository) *StudentService {
    return &StudentService{
        StudentRepo:     studentRepo,
        AchievementRepo: achRepo,
    }
}

// ======================================================
// GET STUDENT PROFILE (OWN PROFILE ONLY)
// ======================================================
func (s *StudentService) Profile(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)

    student, err := s.StudentRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "student not found"})
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "data":   student,
    })
}

// ======================================================
// GET STUDENT ACHIEVEMENTS (OWN ONLY)
// ======================================================
func (s *StudentService) GetAchievements(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)

    student, err := s.StudentRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "student not found"})
    }

    achRefs, err := s.StudentRepo.GetAchievements(student.ID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to get achievements"})
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "data":   achRefs,
    })
}

// ======================================================
// CREATE ACHIEVEMENT (DRAFT)
// ======================================================
func (s *StudentService) CreateAchievement(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)

    student, err := s.StudentRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "student not found"})
    }

    var a models.Achievement
    if err := c.BodyParser(&a); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
    }

    a.ID = uuid.New().String()
    a.StudentID = student.ID
    a.Status = "draft"

    // simpan ke MongoDB
    if err := s.AchievementRepo.Create(&a); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed create achievement"})
    }

    // insert reference ke PostgreSQL (langsung NO CHECK)
    refID := uuid.New().String()
    _, err = s.StudentRepo.DB.Exec(`
        INSERT INTO achievement_references 
            (id, student_id, mongo_achievement_id, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
    `, refID, student.ID, a.ID, "draft")

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed create reference"})
    }

    return c.Status(201).JSON(fiber.Map{
        "status": "success",
        "data":   a,
    })
}

// ======================================================
// SUBMIT ACHIEVEMENT (ONLY OWNER & ONLY DRAFT)
// ======================================================
func (s *StudentService) SubmitAchievement(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    achID := c.Params("id")

    student, err := s.StudentRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "student not found"})
    }

    // cek reference (cek kepemilikan)
    ref, err := s.StudentRepo.FindAchievementByID(achID, student.ID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
    }

    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "only draft can be submitted"})
    }

    // update PG
    if err := s.StudentRepo.UpdateStatus(ref.ID, "submitted", nil, nil); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed update status"})
    }

    // update Mongo
    ach, err := s.AchievementRepo.GetByID(ref.MongoAchievementID)
    if err == nil {
        ach.Status = "submitted"
        _ = s.AchievementRepo.Update(ach)
    }

    return c.JSON(fiber.Map{
        "status": "success",
        "data":   ach,
    })
}

// ======================================================
// DELETE ACHIEVEMENT (ONLY OWNER & ONLY DRAFT)
// ======================================================
func (s *StudentService) DeleteAchievement(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    achID := c.Params("id")

    student, err := s.StudentRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "student not found"})
    }

    ref, err := s.StudentRepo.FindAchievementByID(achID, student.ID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
    }

    if ref.Status != "draft" {
        return c.Status(400).JSON(fiber.Map{"error": "cannot delete non-draft"})
    }

    _ = s.StudentRepo.SoftDelete(ref.ID)
    _ = s.AchievementRepo.SoftDelete(ref.MongoAchievementID)

    return c.JSON(fiber.Map{
        "status":  "success",
        "message": "achievement deleted",
    })
}
// ======================================================
// UPLOAD ATTACHMENT (ONLY OWNER)
// ======================================================
func (s *StudentService) UploadAttachment(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	achID := c.Params("id")

	// ===========================
	// Ambil student
	// ===========================
	student, err := s.StudentRepo.FindByUserID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "student not found"})
	}

	// ===========================
	// Ambil reference untuk cek kepemilikan
	// ===========================
	ref, err := s.StudentRepo.FindAchievementByID(achID, student.ID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
	}

	// ===========================
	// Ambil file
	// ===========================
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "file is required"})
	}

	// Validasi tipe file
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".pdf" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid file type"})
	}

	// ===========================
	// Simpan file ke folder uploads
	// ===========================
	newFileName := fmt.Sprintf("%s_%d%s", ref.MongoAchievementID, time.Now().Unix(), ext)
	savePath := "./uploads/achievements/" + newFileName

	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to save file"})
	}

	// ===========================
	// Buat attachment struct
	// ===========================
	attachment := models.Attachment{
		FileName:   file.Filename,
		FileURL:    "/uploads/achievements/" + newFileName,
		FileType:   ext,
		UploadedAt: time.Now(),
	}

	// ===========================
	// Masukkan ke Mongo (append ke attachments)
	// ===========================
	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err := s.AchievementRepo.AddAttachment(objID, attachment); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update attachment"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   attachment,
	})
}
