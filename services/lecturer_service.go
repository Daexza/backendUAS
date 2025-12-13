package services

import (
    "achievements-uas/app/models"
    "achievements-uas/app/repository"
    "github.com/gofiber/fiber/v2"
)

type LecturerService struct {
    LecturerRepo    *repository.LecturerRepository
    StudentRepo     *repository.StudentRepository
    AchievementRepo *repository.AchievementMongoRepository
}

func NewLecturerService(
    lecturerRepo *repository.LecturerRepository,
    studentRepo *repository.StudentRepository,
    achRepo *repository.AchievementMongoRepository,
) *LecturerService {
    return &LecturerService{
        LecturerRepo:    lecturerRepo,
        StudentRepo:     studentRepo,
        AchievementRepo: achRepo,
    }
}

// ======================================================
// GET ADVISEES
// ======================================================
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)

    lecturer, err := s.LecturerRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
    }

    students, err := s.LecturerRepo.GetAdvisees(lecturer.ID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed get advisees"})
    }

    return c.JSON(fiber.Map{"status": "success", "data": students})
}

// ======================================================
// GET ACHIEVEMENTS OF ADVISEES ONLY
// ======================================================
func (s *LecturerService) GetAchievements(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    lecturer, err := s.LecturerRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
    }

    // list mahasiswa bimbingan
    students, err := s.LecturerRepo.GetAdvisees(lecturer.ID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed get advisees"})
    }

    var studentIDs []string
    for _, st := range students {
        studentIDs = append(studentIDs, st.ID)
    }

    achRefs, err := s.LecturerRepo.GetAchievementsByStudentIDs(studentIDs)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed get achievements"})
    }

    return c.JSON(fiber.Map{"status": "success", "data": achRefs})
}

// ======================================================
// VERIFY ACHIEVEMENT (ONLY FOR OWN ADVISEE, ONLY SUBMITTED)
// ======================================================
func (s *LecturerService) VerifyAchievement(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    lecturer, err := s.LecturerRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
    }

    achID := c.Params("id")

    // ambil ref tanpa batas student
    ref, err := s.StudentRepo.FindAchievementByID(achID, "")
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
    }

    //cek apakah milik mahasiswa bimbingan
    isAdvisee, _ := s.LecturerRepo.IsAdvisee(lecturer.ID, ref.StudentID)
    if !isAdvisee {
        return c.Status(403).JSON(fiber.Map{"error": "not your advisee"})
    }

    if ref.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{"error": "can only verify submitted achievements"})
    }

    // update PostgreSQL
    err = s.LecturerRepo.UpdateStatus(ref.ID, "verified", lecturer.ID, nil)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed update status"})
    }

    // update Mongo
    ach, err := s.AchievementRepo.GetByID(ref.MongoAchievementID)
    if err == nil {
        ach.Status = "verified"
        _ = s.AchievementRepo.Update(ach)
    }

    return c.JSON(fiber.Map{"status": "success", "data": ach})
}

// ======================================================
// REJECT ACHIEVEMENT
// ======================================================
func (s *LecturerService) RejectAchievement(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    lecturer, err := s.LecturerRepo.FindByUserID(userID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
    }

    achID := c.Params("id")

    // ambil ref
    ref, err := s.StudentRepo.FindAchievementByID(achID, "")
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "achievement not found"})
    }

    // ❗ cek apakah ini mahasiswa bimbingan dosen
    isAdvisee, _ := s.LecturerRepo.IsAdvisee(lecturer.ID, ref.StudentID)
    if !isAdvisee {
        return c.Status(403).JSON(fiber.Map{"error": "not your advisee"})
    }

    if ref.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{"error": "can only reject submitted achievements"})
    }

    var body struct {
        Note string `json:"note"`
    }
    _ = c.BodyParser(&body)

    // update PostgreSQL
    err = s.LecturerRepo.UpdateStatus(ref.ID, "rejected", lecturer.ID, &body.Note)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed update status"})
    }

    // update Mongo
    ach, err := s.AchievementRepo.GetByID(ref.MongoAchievementID)
    if err == nil {
        ach.Status = "rejected"
        _ = s.AchievementRepo.Update(ach)
    }

    return c.JSON(fiber.Map{"status": "success", "data": ach})
}
