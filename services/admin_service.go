package services

import (
	"context"
	"strconv"
	"time"
	"achievements-uas/app/models"
	"achievements-uas/app/repository"
	"achievements-uas/utils"
	"fmt"
	

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserAdminService struct {
	AdminRepo    *repository.AdminRepository
	RoleRepo     *repository.RoleRepository
	RolePermRepo *repository.RolePermissionRepository

	// ===== FR-010 =====
	AchPgRepo    *repository.AchievementPostgresRepository
	AchMongoRepo *repository.AchievementMongoRepository
}

// ==============================================
// CONSTRUCTOR
// ==============================================
func NewAdminService(
	adminRepo *repository.AdminRepository,
	roleRepo *repository.RoleRepository,
	rolePermRepo *repository.RolePermissionRepository,
	achPgRepo *repository.AchievementPostgresRepository,
	achMongoRepo *repository.AchievementMongoRepository,
) *UserAdminService {
	return &UserAdminService{
		AdminRepo:    adminRepo,
		RoleRepo:     roleRepo,
		RolePermRepo: rolePermRepo,
		AchPgRepo:    achPgRepo,
		AchMongoRepo: achMongoRepo,
	}
}

func (s *UserAdminService) Create(c *fiber.Ctx) error {
    var body struct {
        Username     string `json:"username"`
        Email        string `json:"email"`
        Password     string `json:"password"`
        FullName     string `json:"full_name"`
        RoleID       string `json:"role_id"`
        // Field Mahasiswa [cite: 88, 89]
        StudentID    string `json:"student_id"` 
        ProgramStudy string `json:"program_study"`
        AcademicYear string `json:"academic_year"`
        // Field Dosen [cite: 101]
        LecturerID   string `json:"lecturer_id"`
        Department   string `json:"department"`
    }

    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
    }

    role, err := s.RoleRepo.FindByID(body.RoleID)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid role_id"})
    }

    hash, _ := utils.HashPassword(body.Password)
    userID := uuid.New().String()

    // Data User Dasar [cite: 33-36]
    user := &models.User{
        ID:           userID,
        Username:     body.Username,
        Email:        body.Email,
        PasswordHash: hash,
        FullName:     body.FullName,
        RoleID:       body.RoleID,
        IsActive:     true,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    // Eksekusi berdasarkan Role [cite: 223-228]
    switch role.Name {
    case "Admin":
        if err := s.AdminRepo.CreateUser(user); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to create admin"})
        }
        return c.Status(201).JSON(fiber.Map{"status": "success", "data": user})

    case "Mahasiswa":
        student := &models.Student{
            ID:           uuid.New().String(),
            UserID:       userID,
            StudentID:    body.StudentID,
            ProgramStudy: body.ProgramStudy,
            AcademicYear: body.AcademicYear,
            CreatedAt:    time.Now(),
        }
        if err := s.AdminRepo.CreateStudent(user, student); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to create student and user"})
        }
        return c.Status(201).JSON(fiber.Map{"status": "success", "data": student})

    case "Dosen Wali":
        lecturer := &models.Lecturer{
            ID:         uuid.New().String(),
            UserID:     userID,
            LecturerID: body.LecturerID,
            Department: body.Department,
            CreatedAt:  time.Now(),
        }
        if err := s.AdminRepo.CreateLecturer(user, lecturer); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "failed to create lecturer and user"})
        }
        return c.Status(201).JSON(fiber.Map{"status": "success", "data": lecturer})

    default:
        return c.Status(400).JSON(fiber.Map{"error": "role not supported"})
    }
}

// ==============================================
// GET ALL USERS
// ==============================================
func (s *UserAdminService) GetAll(c *fiber.Ctx) error {
	users, err := s.AdminRepo.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get users"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": users})
}

// ==============================================
// GET USER BY ID
// ==============================================
func (s *UserAdminService) GetByID(c *fiber.Ctx) error {
	user, err := s.AdminRepo.GetByID(c.Params("id"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": user})
}

// ==============================================
// UPDATE USER
// ==============================================
func (s *UserAdminService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	if _, err := s.RoleRepo.FindByID(body.RoleID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid role_id"})
	}

	user := &models.User{
		ID:       id,
		Username: body.Username,
		Email:    body.Email,
		FullName: body.FullName,
		RoleID:   body.RoleID,
		IsActive: body.IsActive,
	}

	if err := s.AdminRepo.UpdateUser(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed update user"})
	}

	return c.JSON(fiber.Map{"status": "success", "data": user})
}

func (s *UserAdminService) Delete(c *fiber.Ctx) error {
    // 1. Ambil ID dari parameter URL [cite: 248-249]
    id := c.Params("id")

    // 2. Validasi sederhana jika ID tidak ada
    if id == "" {
        return c.Status(400).JSON(fiber.Map{
            "error":   "Bad Request",
            "message": "User ID is required", // Sesuai kode error 400 di SRS [cite: 323]
        })
    }

    // 3. Panggil repository yang menggunakan Transaksi
    // Fungsi ini akan otomatis menghapus di tabel students/lecturers dulu [cite: 85-108]
    if err := s.AdminRepo.DeleteUser(id); err != nil {
        // Jika gagal karena masalah database
        return c.Status(500).JSON(fiber.Map{
            "error":   "Internal Server Error",
            "message": "Failed to delete user and associated profiles", // Sesuai kode error 500 [cite: 323]
        })
    }

    // 4. Return success response
    return c.Status(200).JSON(fiber.Map{
        "status":  "success",
        "message": "User and related profiles deleted successfully",
    })
}

// ==============================================
// UPDATE USER PASSWORD (ADMIN)
// ==============================================
func (s *UserAdminService) UpdatePassword(c *fiber.Ctx) error {
	userID := c.Params("id")

	var body struct {
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&body); err != nil || body.NewPassword == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "new_password is required",
		})
	}

	// hash password baru
	hash, err := utils.HashPassword(body.NewPassword)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed hash password",
		})
	}

	// update ke DB
	if err := s.AdminRepo.UpdatePassword(userID, hash); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed update password",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "password updated",
	})
}

//
// =====================================================
// FR-010: VIEW ALL ACHIEVEMENTS (ADMIN)
// =====================================================
func (s *UserAdminService) GetAllAchievements(c *fiber.Ctx) error {
	ctx := context.Background()

	status := c.Query("status", "")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	offset := (page - 1) * limit

	refs, err := s.AchPgRepo.GetAll(ctx, status, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed load references"})
	}

	var ids []primitive.ObjectID
	for _, r := range refs {
		if oid, err := primitive.ObjectIDFromHex(r.MongoAchievementID); err == nil {
			ids = append(ids, oid)
		}
	}

	data, err := s.AchMongoRepo.FindByIDs(ctx, ids)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed load achievements"})
	}

	return c.JSON(fiber.Map{"status": "success", "data": data})
}

//
// ================= STUDENTS (ADMIN) =================
//

func (s *UserAdminService) GetAllStudents(c *fiber.Ctx) error {
    data, err := s.AdminRepo.GetAllStudents()
    if err != nil {
        // Tambahkan log di terminal agar mudah debug
        fmt.Printf("[ERROR] GetAllStudents: %v\n", err) 
        
        return c.Status(500).JSON(fiber.Map{
            "error": "failed get students",
            "message": err.Error(), // Kirim pesan asli ke Postman selama tahap dev
        })
    }
    
    // Jika data kosong, berikan array kosong [], bukan null
    if data == nil {
        data = []models.Student{}
    }

    return c.JSON(fiber.Map{
        "status": "success", 
        "data": data,
    })
}
func (s *UserAdminService) GetStudentByID(c *fiber.Ctx) error {
    id := c.Params("id")
    data, err := s.AdminRepo.GetStudentByID(id)
    
    if err != nil {
        // Cek terminal jika error, bisa jadi ID tidak ditemukan atau Scan error
        fmt.Printf("[DEBUG] Error GetStudentByID: %v\n", err)
        
        return c.Status(404).JSON(fiber.Map{
            "error": "student not found",
            "message": err.Error(), // Menampilkan pesan asli untuk mempermudah debug
        })
    }
    
    return c.JSON(fiber.Map{
        "status": "success", 
        "data": data,
    })
}

func (s *UserAdminService) GetStudentAchievements(c *fiber.Ctx) error {
	data, err := s.AdminRepo.GetStudentAchievements(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get achievements"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": data})
}

func (s *UserAdminService) SetAdvisor(c *fiber.Ctx) error {
    var body struct {
        AdvisorID string `json:"advisor_id"`
    }

    // 1. Cek apakah JSON bisa di-parse
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid json format"})
    }

    // 2. Cek apakah advisor_id benar-benar ada isinya
    if body.AdvisorID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "advisor_id cannot be empty"})
    }

    // 3. Eksekusi ke Repository
    studentID := c.Params("id")
    if err := s.AdminRepo.SetAdvisor(studentID, body.AdvisorID); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{"status": "success", "message": "Advisor assigned"})
}

//
// ================= LECTURERS (ADMIN) =================
//

func (s *UserAdminService) GetAllLecturers(c *fiber.Ctx) error {
	data, err := s.AdminRepo.GetAllLecturers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get lecturers"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": data})
}

func (s *UserAdminService) GetLecturerAdvisees(c *fiber.Ctx) error {
	data, err := s.AdminRepo.GetLecturerAdvisees(c.Params("id"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed get advisees"})
	}
	return c.JSON(fiber.Map{"status": "success", "data": data})
}
