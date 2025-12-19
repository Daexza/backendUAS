package routes

import (
	"achievements-uas/middleware"
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
	// 1. PUBLIC ROUTES (Tanpa Token)
	// =====================================================
	authPublic := v1.Group("/auth")
	authPublic.Post("/login", authService.Login)
	authPublic.Post("/refresh", authService.Refresh)

	// =====================================================
	// 2. PROTECTED ROUTES (Wajib Login & Cek Blacklist)
	// =====================================================
	// Semua yang menggunakan 'protected' akan dicek oleh middleware AuthRequired
	protected := v1.Group("/", middleware.AuthRequired())

	// AUTH - Profile & Logout
	protected.Post("/auth/logout", authService.Logout)
	protected.Get("/auth/profile", authService.Profile)

	// USERS (ADMIN ONLY) - FR-009
	// Kita tambahkan RoleRequired agar hanya Admin yang bisa akses
	users := protected.Group("/users", middleware.RoleRequired("Admin"))
	users.Get("/", adminService.GetAll)
	users.Get("/:id", adminService.GetByID)
	users.Post("/", adminService.Create)
	users.Put("/:id", adminService.Update)
	users.Delete("/:id", adminService.Delete)
	users.Put("/:id/password", adminService.UpdatePassword)

	// ACHIEVEMENTS - FR-003 s/d FR-008
	ach := protected.Group("/achievements")
	ach.Get("/", achievementService.List)
	ach.Get("/:id", achievementService.Detail)
	ach.Get("/:id/history", achievementService.History)
	ach.Post("/", achievementService.Create)
	ach.Put("/:id", achievementService.Update)
	ach.Post("/:id/submit", achievementService.Submit)
	ach.Delete("/:id", achievementService.Delete)
	ach.Post("/:id/attachments", achievementService.UploadAttachment)
	// Verifikasi (Biasanya oleh Dosen/Admin)
	ach.Post("/:id/verify", achievementService.Verify)
	ach.Post("/:id/reject", achievementService.Reject)

	// STUDENTS (ADMIN ONLY) - FR-009
	students := protected.Group("/students", middleware.RoleRequired("Admin"))
	students.Get("/", adminService.GetAllStudents)
	students.Get("/:id", adminService.GetStudentByID)
	students.Get("/:id/achievements", adminService.GetStudentAchievements)
	students.Put("/:id/advisor", adminService.SetAdvisor)

	// LECTURERS (ADMIN ONLY) - FR-006
	lecturers := protected.Group("/lecturers", middleware.RoleRequired("Admin"))
	lecturers.Get("/", adminService.GetAllLecturers)
	lecturers.Get("/:id/advisees", adminService.GetLecturerAdvisees)

	// REPORTS & ANALYTICS - FR-011
// Tetap gunakan AuthRequired agar sistem tahu "Siapa" yang memanggil
reportGroup := protected.Group("/reports", middleware.AuthRequired())
reportGroup.Get("/statistics", reportService.Statistics)
reportGroup.Get("/student/:id", reportService.StudentReport)
}