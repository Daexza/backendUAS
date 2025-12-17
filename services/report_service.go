package services

import (
	"context"

	"achievements-uas/app/models"
	"achievements-uas/app/repository"
	"achievements-uas/utils"

	"github.com/gofiber/fiber/v2"
)

type ReportService struct {
	MongoRepo   *repository.AchievementMongoRepository
	StudentRepo *repository.StudentRepository
}

// =====================================================
// FR-011: STATISTICS
// GET /api/v1/reports/statistics
// =====================================================
func (s *ReportService) Statistics(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	var achievements []models.Achievement
	var err error

	switch claims.Role {

	case "MAHASISWA":
		achievements, err = s.MongoRepo.FindByStudentID(ctx, claims.ID)

	case "DOSEN_WALI":
		students, err2 := s.StudentRepo.FindByAdvisorID(claims.ID)
		if err2 != nil {
			return fiber.ErrInternalServerError
		}
		var ids []string
		for _, s := range students {
			ids = append(ids, s.UserID)
		}
		achievements, err = s.MongoRepo.FindByStudentIDs(ctx, ids)

	case "ADMIN":
		achievements, err = s.MongoRepo.FindAll(ctx)

	default:
		return fiber.ErrForbidden
	}

	if err != nil {
		return fiber.ErrInternalServerError
	}

	// ============================
	// AGGREGATION (IN-MEMORY)
	// ============================
	byType := map[string]int{}
	byCompetitionLevel := map[string]int{}
	totalPoints := 0

	for _, a := range achievements {
		byType[a.Type]++
		totalPoints += a.Points

		if a.Details.CompetitionLevel != "" {
			byCompetitionLevel[a.Details.CompetitionLevel]++
		}
	}

	return c.JSON(fiber.Map{
		"totalAchievements": len(achievements),
		"totalPoints":       totalPoints,
		"byType":            byType,
		"competitionLevel":  byCompetitionLevel,
	})
}

// =====================================================
// FR-011: STUDENT REPORT
// GET /api/v1/reports/student/:id
// =====================================================
func (s *ReportService) StudentReport(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)
	studentID := c.Params("id")

	// Role check
	if claims.Role == "MAHASISWA" && claims.ID != studentID {
		return fiber.ErrForbidden
	}

	if claims.Role == "DOSEN_WALI" {
		ok, err := s.StudentRepo.IsAdvisorOf(claims.ID, studentID)
		if err != nil || !ok {
			return fiber.ErrForbidden
		}
	}

	achievements, err := s.MongoRepo.FindByStudentID(ctx, studentID)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	totalPoints := 0
	byType := map[string]int{}

	for _, a := range achievements {
		totalPoints += a.Points
		byType[a.Type]++
	}

	return c.JSON(fiber.Map{
		"studentId":         studentID,
		"totalAchievements": len(achievements),
		"totalPoints":       totalPoints,
		"byType":            byType,
		"achievements":      achievements,
	})
}
