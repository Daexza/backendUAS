package main

import (
	"log"

	"achievements-uas/database"
	"achievements-uas/app/repository"
	"achievements-uas/routes"
	"achievements-uas/services"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// ===============================
	// LOAD ENV
	// ===============================
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env file not found, using system env")
	}

	// ===============================
	// CONNECT DATABASES
	// ===============================
	if err := database.ConnectPostgres(); err != nil {
		log.Fatal("[FATAL] PostgreSQL error:", err)
	}

	if err := database.ConnectMongo(); err != nil {
		log.Fatal("[FATAL] MongoDB error:", err)
	}

	// ===============================
	// INIT REPOSITORIES
	// ===============================
	adminRepo := repository.NewAdminRepository(database.Postgres)
	roleRepo := repository.NewRoleRepository(database.Postgres)
	rolePermRepo := repository.NewRolePermissionRepository(database.Postgres)
	authRepo := repository.NewAuthRepository(database.Postgres)

	studentRepo := repository.NewStudentRepository(database.Postgres)

	achPgRepo := repository.NewAchievementPostgresRepository(database.Postgres)
	achMongoRepo := repository.NewAchievementMongoRepository(database.MongoDB)

	// ===============================
	// INIT SERVICES
	// ===============================
	authService := services.NewAuthService(authRepo, rolePermRepo)

	adminService := services.NewAdminService(
		adminRepo,
		roleRepo,
		rolePermRepo,
		achPgRepo,
		achMongoRepo,
	)

	achievementService := &services.AchievementService{
		MongoRepo:   achMongoRepo,
		PgRepo:      achPgRepo,
		StudentRepo: studentRepo,
	}

	reportService := &services.ReportService{
		MongoRepo:   achMongoRepo,
		StudentRepo: studentRepo,
	}

	// ===============================
	// INIT APP & ROUTES
	// ===============================
	app := fiber.New()

	routes.SetupRoutes(
		app,
		authService,
		adminService,
		achievementService,
		reportService, // ← WAJIB
	)

	// ===============================
	// START SERVER
	// ===============================
	log.Println("🚀 Server running on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("[FATAL] Server error:", err)
	}
}
