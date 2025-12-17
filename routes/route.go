package routes

import (
	"achievements-uas/services"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	authService *services.AuthService,
	adminService *services.UserAdminService,
	achievementService *services.AchievementService,
	reportService *services.ReportService,
) {

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// =====================================================
	// AUTH
	// FR-001
	// =====================================================
	auth := v1.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", authService.Logout)
	auth.Get("/profile", authService.Profile)

	// =====================================================
	// USERS (ADMIN ONLY)
	// FR-009: Manage Users
	// =====================================================
	users := v1.Group("/users")
	users.Get("/", adminService.GetAll)
	users.Get("/:id", adminService.GetByID)
	users.Post("/", adminService.Create)
	users.Put("/:id", adminService.Update)
	users.Delete("/:id", adminService.Delete)
	users.Put("/:id/password", adminService.UpdatePassword)

	// =====================================================
	// ACHIEVEMENTS
	// FR-003 s/d FR-008, FR-012 (+ DETAIL, HISTORY, UPDATE)
	// =====================================================
	ach := v1.Group("/achievements")

	// LIST (ROLE BASED)
	ach.Get("/", achievementService.List)

	// DETAIL & HISTORY
	ach.Get("/:id", achievementService.Detail)
	ach.Get("/:id/history", achievementService.History)

	// MAHASISWA
	ach.Post("/", achievementService.Create)
	ach.Put("/:id", achievementService.Update)
	ach.Post("/:id/submit", achievementService.Submit)
	ach.Delete("/:id", achievementService.Delete)
	ach.Post("/:id/attachments", achievementService.UploadAttachment)

	// DOSEN WALI
	ach.Post("/:id/verify", achievementService.Verify)
	ach.Post("/:id/reject", achievementService.Reject)

	// =====================================================
	// STUDENTS (ADMIN ONLY)
	// FR-009
	// =====================================================
	students := v1.Group("/students")
	students.Get("/", adminService.GetAllStudents)
	students.Get("/:id", adminService.GetStudentByID)
	students.Get("/:id/achievements", adminService.GetStudentAchievements)
	students.Put("/:id/advisor", adminService.SetAdvisor)

	// =====================================================
	// LECTURERS (ADMIN ONLY)
	// FR-006
	// =====================================================
	lecturers := v1.Group("/lecturers")
	lecturers.Get("/", adminService.GetAllLecturers)
	lecturers.Get("/:id/advisees", adminService.GetLecturerAdvisees)
	// =====================================================
	// REPORTS & ANALYTICS
	// FR-011
	// =====================================================
	reports := v1.Group("/reports")
	reports.Get("/statistics", reportService.Statistics)
	reports.Get("/student/:id", reportService.StudentReport)
}

