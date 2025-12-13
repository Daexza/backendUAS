package routes

import (
	"achievements-uas/app/repositories"
	"achievements-uas/app/services"
	"achievements-uas/database"
	"achievements-uas/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {

	// ======================================
	// INIT REPOSITORIES
	// ======================================
	authRepo := repositories.NewAuthRepository(database.Postgres)
	userRepo := repositories.NewAdminRepository(database.Postgres)
	roleRepo := repositories.NewRoleRepository(database.Postgres)
	rolePermRepo := repositories.NewRolePermissionRepository(database.Postgres)

	studentRepo := repositories.NewStudentRepository(database.Postgres)
	lecturerRepo := repositories.NewLecturerRepository(database.Postgres)
	achMongoRepo := repositories.NewAchievementMongoRepository(database.Mongo)
	achPgRepo := repositories.NewAchievementPGRepository(database.Postgres)

	// ======================================
	// INIT SERVICES
	// ======================================
	authService := services.NewAuthService(authRepo, rolePermRepo)
	userService := services.NewUserAdminService(userRepo, roleRepo, rolePermRepo)

	studentAchievementService := services.NewStudentAchievementService(studentRepo, achMongoRepo, achPgRepo)
	lecturerAchievementService := services.NewLecturerAchievementService(lecturerRepo, achMongoRepo, achPgRepo)
	achievementQueryService := services.NewAchievementQueryService(achMongoRepo, achPgRepo)

	studentService := services.NewStudentService(studentRepo, achPgRepo)
	lecturerService := services.NewLecturerService(lecturerRepo, studentRepo)

	// ============================================================
	//  AUTHENTICATION 
	// ============================================================
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", middleware.AuthRequired(), authService.Logout)
	auth.Get("/profile", middleware.AuthRequired(), authService.Profile)

	// ============================================================
	// USERS (ADMIN) 
	// ============================================================
	users := app.Group("/api/v1/users", middleware.AuthRequired())
	users.Get("/", userService.GetAll)
	users.Get("/:id", userService.GetByID)
	users.Post("/", userService.Create)
	users.Put("/:id", userService.Update)
	users.Delete("/:id", userService.Delete)
	users.Put("/:id/role", userService.UpdateRole)

	// ============================================================
	// ACHIEVEMENTS 
	// ============================================================
	ach := app.Group("/api/v1/achievements", middleware.AuthRequired())

	// GENERAL
	ach.Get("/", achievementQueryService.List)
	ach.Get("/:id", achievementQueryService.Detail)
	ach.Get("/:id/history", achievementQueryService.History)

	// STUDENT ACHIEVEMENTS
	ach.Post("/", studentAchievementService.Create)
	ach.Put("/:id", studentAchievementService.Update)
	ach.Delete("/:id", studentAchievementService.Delete)
	ach.Post("/:id/submit", studentAchievementService.Submit)
	ach.Post("/:id/attachments", studentAchievementService.UploadAttachment)

	// LECTURER ACHIEVEMENTS
	ach.Post("/:id/verify", lecturerAchievementService.Verify)
	ach.Post("/:id/reject", lecturerAchievementService.Reject)

	// ============================================================
	// STUDENTS 
	// ============================================================
	st := app.Group("/api/v1/students", middleware.AuthRequired())
	st.Get("/", studentService.GetProfileOrList)
	st.Get("/:id", studentService.GetByID)
	st.Get("/:id/achievements", studentService.GetAchievements)
	st.Put("/:id/advisor", studentService.SetAdvisor)

	// ============================================================
	// LECTURERS 
	// ============================================================
	lc := app.Group("/api/v1/lecturers", middleware.AuthRequired())
	lc.Get("/", lecturerService.GetAll)
	lc.Get("/:id/advisees", lecturerService.GetAdvisees)
}
