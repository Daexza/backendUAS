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
	// INIT REPOSITORIES
	// ================================================
	authRepo := repository.NewAuthRepository(database.Postgres)
	userAdminRepo := repository.NewAdminRepository(database.Postgres)
	roleRepo := repository.NewRoleRepository(database.Postgres)
	rolePermRepo := repository.NewRolePermissionRepository(database.Postgres)

	// ================================================
	// INIT SERVICES
	// ================================================
	authService := services.NewAuthService(authRepo, rolePermRepo)
	adminService := services.NewAdminService(userAdminRepo, roleRepo, rolePermRepo)

	// ================================================
	// AUTH ROUTES
	// ================================================
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", authService.Login)
	auth.Post("/refresh", authService.Refresh)
	auth.Post("/logout", middleware.AuthRequired(), authService.Logout)
	auth.Get("/profile", middleware.AuthRequired(), authService.Profile)

	// ================================================
	// USERS ADMIN ROUTES (PROTECTED BY PERMISSION)
	// ================================================
	users := app.Group("/api/v1/users",middleware.AuthRequired(),middleware.RequirePermission("user:manage"),)

	users.Get("/", adminService.GetAll)
	users.Get("/:id", adminService.GetByID)
	users.Post("/", adminService.Create)
	users.Put("/:id", adminService.Update)
	users.Delete("/:id", adminService.Delete)

	// RESET PASSWORD (ADMIN)
	users.Put("/:id/password", adminService.UpdatePassword)

	// ================================================
	// ADMIN PROFILES (STUDENT & LECTURER)
	// ================================================
	adminProfiles := app.Group("/api/v1/admin",middleware.AuthRequired(),middleware.RequirePermission("user:manage"),)
	
	adminProfiles.Post("/students", adminService.CreateStudentProfile)
	adminProfiles.Post("/lecturers", adminService.CreateLecturerProfile)
}
