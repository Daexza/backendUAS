package routes

import (
	"achievements-uas/database"
	"achievements-uas/middleware"
	"achievements-uas/repositories"
	"achievements-uas/services"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {

	// ================================================
	// INIT REPOS & SERVICES
	// ================================================
	userRepo := repository.NewUserRepository(database.Postgres)
	authService := services.NewAuthService(userRepo)

	adminRepo := repository.NewUserAdminRepository(database.Postgres)
	adminService := services.NewUserAdminService(adminRepo)

	// achievementRepo := repository.NewAchievementRepository(database.Postgres, database.Mongo)
	// achievementService := services.NewAchievementService(achievementRepo)

	// studentRepo := repository.NewStudentRepository(database.Postgres)
	// studentService := services.NewStudentService(studentRepo, achievementRepo)

	// lecturerRepo := repositories.NewLecturerRepository(database.Postgres)
	// lecturerService := services.NewLecturerService(lecturerRepo, studentRepo, achievementRepo)

	// ================================================
	// AUTH ROUTES
	// ================================================
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", middleware.AuthRequired(), authService.Logout)
	auth.Get("/profile", middleware.AuthRequired(), authService.Profile)

	// ================================================
	// USERS ADMIN ROUTES (RBAC)
	// ================================================
	users := app.Group("/api/v1/users",
		middleware.AuthRequired(),
		middleware.RequirePermission("user:manage"),
	)

	users.Get("/", adminService.GetAll)
	users.Get("/:id", adminService.GetByID)
	users.Post("/", adminService.Create)
	users.Put("/:id", adminService.Update)
	users.Delete("/:id", adminService.Delete)

	// Admin reset password
	users.Put("/:id/password", adminService.UpdatePassword)

	// Admin create student/lecturer profile
	adminProfiles := app.Group("/api/v1/admin",
		middleware.AuthRequired(),
		middleware.RequirePermission("user:manage"),
	)
	adminProfiles.Post("/students", adminService.CreateStudentProfile)
	adminProfiles.Post("/lecturers", adminService.CreateLecturerProfile)

// 	// ================================================
// 	// STUDENT ROUTES
// 	// ================================================
// 	students := app.Group("/api/v1/students", middleware.AuthRequired())
// 	students.Get("/", studentService.GetAll)
// 	students.Get("/:id", studentService.GetByID)
// 	students.Get("/:id/achievements", studentService.GetAchievements)
// 	students.Put("/:id/advisor", middleware.RequirePermission("advisor:assign"), studentService.SetAdvisor)

// 	// ================================================
// 	// LECTURER ROUTES
// 	// ================================================
// 	lecturers := app.Group("/api/v1/lecturers", middleware.AuthRequired())
// 	lecturers.Get("/", lecturerService.GetAll)
// 	lecturers.Get("/:id/advisees", lecturerService.GetAdvisees)

// 	// ================================================
// 	// ACHIEVEMENT ROUTES
// 	// ================================================
// 	achievements := app.Group("/api/v1/achievements", middleware.AuthRequired())
// 	achievements.Get("/", achievementService.GetAll)
// 	achievements.Get("/:id", achievementService.GetByID)
// 	achievements.Post("/", achievementService.Create) // Mahasiswa
// 	achievements.Put("/:id", achievementService.Update) // Mahasiswa
// 	achievements.Delete("/:id", achievementService.Delete) // Mahasiswa

// 	achievements.Post("/:id/submit", achievementService.Submit) // Mahasiswa submit draft
// 	achievements.Post("/:id/verify", middleware.RequirePermission("advisor:verify"), achievementService.Verify)
// 	achievements.Post("/:id/reject", middleware.RequirePermission("advisor:verify"), achievementService.Reject)
// 	achievements.Get("/:id/history", achievementService.GetHistory)
// 	achievements.Post("/:id/attachments", achievementService.UploadAttachment)
}
