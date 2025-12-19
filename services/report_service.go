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
// FR-011: STATISTICS (Global untuk Admin/Lecturer/Own)
// GET /api/v1/reports/statistics
// =====================================================
func (s *ReportService) Statistics(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	var achievements []models.Achievement
	var err error

	// Penentuan data berdasarkan Role (RBAC)
	switch claims.Role {
	case "Mahasiswa":
		// Mengambil data NIM mahasiswa berdasarkan UserID di token
		student, errInfo := s.StudentRepo.FindByUserID(claims.ID)
		if errInfo != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"})
		}
		achievements, err = s.MongoRepo.FindByStudentID(ctx, student.StudentID)

	case "Dosen Wali":
		// Mengambil semua mahasiswa bimbingan dosen ini
		students, errInfo := s.StudentRepo.FindByAdvisorID(claims.ID)
		if errInfo != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch advisees"})
		}
		
		var nims []string
		for _, st := range students {
			nims = append(nims, st.StudentID)
		}
		
		if len(nims) > 0 {
			achievements, err = s.MongoRepo.FindByStudentIDs(ctx, nims)
		}

	case "Admin":
		// Admin bisa menarik seluruh data prestasi
		achievements, err = s.MongoRepo.FindAll(ctx)

	default:
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden: Role not recognized"})
	}

	// Cek jika ada error saat fetch dari MongoDB
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch achievements data"})
	}

	// =====================================================
	// AGGREGASI DATA (OUTPUT FR-011)
	// =====================================================
	byType := map[string]int{}
	byCompetitionLevel := map[string]int{}
	byPeriod := map[string]int{}    // Total prestasi per periode (Tahun)
	topStudents := map[string]int{} // Top mahasiswa berprestasi (Poin)
	totalPoints := 0

	for _, a := range achievements {
		// 1. Total per Tipe
		byType[a.AchievementType]++
		
		// 2. Akumulasi Poin
		totalPoints += a.Points
		
		// 3. Distribusi Tingkat Kompetisi
		if a.Details.CompetitionLevel != "" {
			byCompetitionLevel[a.Details.CompetitionLevel]++
		}

		// 4. Total per Periode (Tahun)
		yearStr := a.CreatedAt.Format("2006")
		byPeriod[yearStr]++

		// 5. Top Mahasiswa (Berdasarkan NIM)
		topStudents[a.StudentID] += a.Points
	}

	return c.JSON(fiber.Map{
		"totalAchievements": len(achievements),
		"totalPoints":       totalPoints,
		"byType":            byType,             // FR-011: Total per tipe
		"byPeriod":          byPeriod,           // FR-011: Total per periode
		"competitionLevel":  byCompetitionLevel, // FR-011: Distribusi tingkat kompetisi
		"topStudents":       topStudents,        // FR-011: Top mahasiswa
	})
}

// =====================================================
// FR-011: STUDENT REPORT (Detail Kumulatif Individu)
// GET /api/v1/reports/student/:id
// =====================================================
func (s *ReportService) StudentReport(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)
	targetUserID := c.Params("id") // UUID dari URL

	// Security: Mahasiswa tidak boleh intip laporan orang lain
	if claims.Role == "Mahasiswa" && claims.ID != targetUserID {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden: Access denied"})
	}

	// Security: Dosen hanya boleh lihat bimbingannya
	if claims.Role == "Dosen Wali" {
		isAdvisee, err := s.StudentRepo.IsAdvisorOf(claims.ID, targetUserID)
		if err != nil || !isAdvisee {
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden: This student is not your advisee"})
		}
	}

	// Ambil profil untuk mendapatkan NIM
	student, err := s.StudentRepo.FindByUserID(targetUserID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"})
	}

	// Ambil data prestasi dari MongoDB
	achievements, err := s.MongoRepo.FindByStudentID(ctx, student.StudentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch student achievements"})
	}

	totalPoints := 0
	byType := map[string]int{}

	for _, a := range achievements {
		totalPoints += a.Points
		byType[a.AchievementType]++
	}

	return c.JSON(fiber.Map{
		"studentProfile": fiber.Map{
			"nim":           student.StudentID,
			"program_study": student.ProgramStudy,
			"academic_year": student.AcademicYear,
		},
		"statistics": fiber.Map{
			"totalAchievements": len(achievements),
			"totalPoints":       totalPoints,
			"byType":            byType,
		},
		"achievements": achievements,
	})
}