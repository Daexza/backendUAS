package services

import (
	"context"
	"fmt"
	"time"
	"log"
	"achievements-uas/app/models"
	"achievements-uas/app/repository"
	"achievements-uas/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService struct {
	MongoRepo *repository.AchievementMongoRepository
	PgRepo    *repository.AchievementPostgresRepository
	AdminRepo *repository.AdminRepository
}

// GET /api/v1/achievements
// FR-006, FR-010, & Mahasiswa List
func (s *AchievementService) List(c *fiber.Ctx) error {
    ctx := context.Background()
    claims := c.Locals("claims").(*utils.JWTClaims)

    // CASE 1: ADMIN (Lihat Semua)
    if claims.Role == "Admin" {
        refs, total, err := s.PgRepo.GetAllWithCount(ctx, c.Query("status"), "", c.QueryInt("limit", 10), c.QueryInt("offset", 0))
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch references"})
        }
        return s.fetchFullDataFromMongo(c, refs, total)
    }

    // CASE 2: DOSEN WALI (Hanya bimbingan sendiri)
    if claims.Role == "Dosen Wali" {
        // Ambil bimbingan berdasarkan UUID Dosen (advisor_id di DB)
        advisees, err := s.AdminRepo.GetLecturerAdvisees(claims.ID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data mahasiswa bimbingan"})
        }

        if len(advisees) == 0 {
            return c.JSON(fiber.Map{"total": 0, "data": []interface{}{}})
        }

        var studentUUIDs []string
        for _, st := range advisees {
            studentUUIDs = append(studentUUIDs, st.ID) // Menggunakan UUID Mahasiswa
        }

        refs, err := s.PgRepo.FindByStudentIDs(ctx, studentUUIDs)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Gagal sinkronisasi data referensi"})
        }
        return s.fetchFullDataFromMongo(c, refs, len(refs))
    }

    // CASE 3: MAHASISWA (Milik sendiri)
    if claims.Role == "Mahasiswa" {
        // Ambil profil untuk dapat NIM-nya agar bisa query ke Mongo
        student, err := s.AdminRepo.GetStudentByUserID(claims.ID)
        if err != nil {
            return c.Status(404).JSON(fiber.Map{"error": "Profil tidak ditemukan"})
        }
        data, err := s.MongoRepo.FindByStudentID(ctx, student.StudentID)
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data"})
        }
        return c.JSON(data)
    }

    return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
}

// GET /api/v1/achievements/:id
func (s *AchievementService) Detail(c *fiber.Ctx) error {
    ctx := context.Background()
    claims := c.Locals("claims").(*utils.JWTClaims)
    mongoID := c.Params("id")

    // 1. Ambil data dari MongoDB berdasarkan ID Prestasi
    data, err := s.MongoRepo.GetByID(ctx, mongoID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Prestasi tidak ditemukan"})
    }

    // 2. CEK STUDENT ID (Cross-check Kepemilikan)
    if claims.Role == "Mahasiswa" {
        // Ambil profil mahasiswa dari Postgres berdasarkan UserID di JWT
        me, err := s.AdminRepo.GetStudentByUserID(claims.ID)
        if err != nil {
            return c.Status(403).JSON(fiber.Map{"error": "Profil mahasiswa tidak valid"})
        }

        // VALIDASI: Apakah StudentID di MongoDB sama dengan NIM mahasiswa yang login?
        if data.StudentID != me.StudentID {
            // Jika beda, berarti dia mencoba melihat prestasi orang lain
            return c.Status(403).JSON(fiber.Map{
                "error": "Akses Ditolak: Anda hanya boleh melihat detail prestasi milik sendiri!",
            })
        }
    }

    // 3. CEK UNTUK DOSEN WALI (Filter Bimbingan)
    if claims.Role == "Dosen Wali" {
        // Pastikan mahasiswa pemilik prestasi ini adalah anak bimbingannya
        isAdvisee, err := s.AdminRepo.CheckIsAdvisee(data.StudentID, claims.ID)
        if err != nil || !isAdvisee {
            return c.Status(403).JSON(fiber.Map{"error": "Akses Ditolak: Ini bukan mahasiswa bimbingan anda"})
        }
    }

    return c.JSON(data)
}

// POST /api/v1/achievements
// FR-003: Submit Prestasi (Initial Draft)
func (s *AchievementService) Create(c *fiber.Ctx) error {
    ctx := context.Background()
    claims := c.Locals("claims").(*utils.JWTClaims)

    var ach models.Achievement
    if err := c.BodyParser(&ach); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
    }

    student, err := s.AdminRepo.GetStudentByUserID(claims.ID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student profile not found"})
    }

    ach.StudentID = student.StudentID
    ach.Status = "draft"
    ach.Points = 0 
    ach.CreatedAt = time.Now()
    ach.UpdatedAt = time.Now()
    
    // Isi history awal secara internal
    ach.History = []models.AchievementHistory{
        {
            StudentID: student.StudentID,
            Status:    "draft",
            ChangedBy: claims.Username, 
            ChangedAt: time.Now(),
            Notes:     "Initial draft created",
        },
    }

    mongoData, err := s.MongoRepo.Create(ctx, &ach)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save to MongoDB"})
    }

    ref := models.AchievementReference{
        ID:                 uuid.New().String(),
        StudentID:          student.ID,
        MongoAchievementID: mongoData.ID.Hex(),
        Status:             "draft",
        CreatedAt:          time.Now(),
        UpdatedAt:          time.Now(),
    }
    s.PgRepo.Create(ctx, ref) 

    // --- PERBAIKAN: Sembunyikan History dari Response ---
    // Kita buat map manual atau set History ke nil sebelum JSON
    mongoData.History = nil 

    return c.Status(201).JSON(mongoData)
}

func (s *AchievementService) Update(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    claims := c.Locals("claims").(*utils.JWTClaims)

    idParam := c.Params("id")
    oid, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Format ID tidak valid"})
    }

    // 1. Ambil data lama dari MongoDB
    oldData, err := s.MongoRepo.GetByID(ctx, idParam)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Prestasi tidak ditemukan"})
    }

    // --- PENGUNCI STATUS: HANYA DRAFT ---
    if oldData.Status != "draft" {
        return c.Status(403).JSON(fiber.Map{
            "error": "Akses ditolak",
            "message": "Hanya prestasi dengan status 'draft' yang dapat diubah",
        })
    }

    var input models.Achievement
    if err := c.BodyParser(&input); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Data input tidak valid"})
    }

    // 2. Update di MongoDB
    now := time.Now()
    updateQuery := bson.M{
        "$set": bson.M{
            "achievementType": input.AchievementType,
            "title":           input.Title,
            "description":     input.Description,
            "details":         input.Details,
            "tags":            input.Tags,
            "updatedAt":       now,
        },
        "$push": bson.M{
            "history": models.AchievementHistory{
                Status:    "draft",
                ChangedBy: claims.Username,
                ChangedAt: now,
                Notes:     "Melakukan perubahan data prestasi",
            },
        },
    }

    err = s.MongoRepo.Update(ctx, oid, updateQuery)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Gagal update di MongoDB"})
    }

    // 3. Update di PostgreSQL (Sinkronisasi Timestamp)
    // Supaya di dashboard dosen, data ini naik ke urutan paling atas karena baru saja diupdate
    err = s.PgRepo.UpdateTimestamp(ctx, idParam)
    if err != nil {
        // Kita log saja, jangan gagalkan response karena data utama sudah aman di Mongo
        log.Printf("Warning: Gagal sinkronisasi timestamp ke Postgres untuk ID %s: %v", idParam, err)
    }

    return c.JSON(fiber.Map{
        "message": "Prestasi berhasil diperbarui",
        "id":      idParam,
    })
}

// DELETE /api/v1/achievements/:id
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)
	idParam := c.Params("id")

	// 1. Cari data di MongoDB untuk cek status & kepemilikan
	achievement, err := s.MongoRepo.GetByID(ctx, idParam)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
	}

	if achievement.Status == "deleted" {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot update a deleted achievement"})
    }

	// 2. Precondition: Hanya status 'draft' yang boleh dihapus
	if achievement.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{"error": "Only draft achievements can be deleted"})
	}

	// 3. Pastikan yang menghapus adalah pemiliknya (Mahasiswa yang login)
	student, _ := s.AdminRepo.GetStudentByUserID(claims.ID)
	if achievement.StudentID != student.StudentID {
		return c.Status(403).JSON(fiber.Map{"error": "You are not authorized to delete this data"})
	}

	// 4. FLOW 1: Soft delete data di MongoDB (Ubah status jadi 'deleted')
	oid, _ := primitive.ObjectIDFromHex(idParam)
	err = s.MongoRepo.SoftDelete(ctx, oid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to soft delete in MongoDB"})
	}

	// 5. FLOW 2: Update reference di PostgreSQL (Ubah status jadi 'deleted')
	// Pastikan PgRepo memiliki fungsi UpdateStatus
	// Di dalam AchievementService.Delete
	err = s.PgRepo.UpdateStatus(ctx, idParam, "deleted")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update Postgres reference"})
		}

	// 6. FLOW 3: Return success message
	return c.JSON(fiber.Map{
		"message": "Achievement draft deleted successfully",
	})
}

// POST /api/v1/achievements/:id/submit
// FR-004: Submit for verification
func (s *AchievementService) Submit(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    claims := c.Locals("claims").(*utils.JWTClaims)
    mongoID := c.Params("id")

    // 1. Validasi: Pastikan data ada dan statusnya masih 'draft' atau 'rejected'
    oldData, err := s.MongoRepo.GetByID(ctx, mongoID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }
    if oldData.Status != "draft" && oldData.Status != "rejected" {
        return c.Status(400).JSON(fiber.Map{"error": "Hanya status draft atau rejected yang bisa di-submit"})
    }

    now := time.Now()
    
    // 2. Update MongoDB (Status & History)
    newHistory := models.AchievementHistory{
        Status:    "submitted",
        ChangedBy: claims.Username,
        ChangedAt: now,
        Notes:     "Mahasiswa mengajukan verifikasi prestasi",
    }

    updateQuery := bson.M{
        "$set":  bson.M{"status": "submitted", "updatedAt": now},
        "$push": bson.M{"history": newHistory},
    }

    oid, _ := primitive.ObjectIDFromHex(mongoID)
    if err := s.MongoRepo.Update(ctx, oid, updateQuery); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Gagal update MongoDB"})
    }

    // 3. SINKRONISASI KE POSTGRESQL (Mengisi submitted_at)
    // Kita panggil fungsi khusus agar submitted_at tidak [null] lagi
    if err := s.PgRepo.UpdateToSubmitted(ctx, mongoID); err != nil {
        log.Printf("Postgres Sync Error: %v", err)
    }

    return c.JSON(fiber.Map{
        "message": "Achievement submitted successfully",
        "status":  "submitted",
    })
}

// POST /api/v1/achievements/:id/verify
// FR-007: Verify (Dosen Wali)
func (s *AchievementService) Verify(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Ambil data dosen dari JWT (pastikan middleware JWT sudah terpasang)
    claims := c.Locals("claims").(*utils.JWTClaims)
    dosenUUID := claims.ID // Ini adalah UUID dosen yang sedang login
    mongoID := c.Params("id")

    // 1. Ambil data dari Mongo untuk validasi status
    oldData, err := s.MongoRepo.GetByID(ctx, mongoID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Prestasi tidak ditemukan"})
    }

    // Hanya bisa verifikasi jika statusnya 'submitted'
    if oldData.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{"error": "Hanya prestasi dengan status 'submitted' yang bisa diverifikasi"})
    }

    now := time.Now()

    // 2. Update status di MongoDB & Tambah History
    newHistory := models.AchievementHistory{
        Status:    "verified",
        ChangedBy: claims.Username, // Nama dosen
        ChangedAt: now,
        Notes:     "Prestasi telah diverifikasi dan disetujui oleh Dosen Wali",
    }

    updateQuery := bson.M{
        "$set":  bson.M{"status": "verified", "updatedAt": now},
        "$push": bson.M{"history": newHistory},
    }

    oid, _ := primitive.ObjectIDFromHex(mongoID)
    if err := s.MongoRepo.Update(ctx, oid, updateQuery); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Gagal update data di MongoDB"})
    }

    // 3. SINKRONISASI KE POSTGRESQL
    // Mengupdate status, verified_at, dan verified_by (ID Dosen)
    if err := s.PgRepo.UpdateToVerified(ctx, mongoID, dosenUUID); err != nil {
        log.Printf("Postgres Sync Error: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Gagal sinkronisasi data ke PostgreSQL"})
    }

    return c.JSON(fiber.Map{
        "message": "Prestasi berhasil diverifikasi",
        "status":  "verified",
        "verified_by": dosenUUID,
    })
}

// POST /api/v1/achievements/:id/reject
// FR-008: Reject (Dosen Wali)
func (s *AchievementService) Reject(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    claims := c.Locals("claims").(*utils.JWTClaims)
    mongoID := c.Params("id")

    // 1. Ambil alasan penolakan dari body JSON
    var input struct {
        Reason string `json:"reason"`
    }
    if err := c.BodyParser(&input); err != nil || input.Reason == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Alasan penolakan (reason) wajib diisi"})
    }

    // 2. Validasi: Cek keberadaan data di MongoDB
    oldData, err := s.MongoRepo.GetByID(ctx, mongoID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Prestasi tidak ditemukan"})
    }

    // Hanya bisa reject jika statusnya 'submitted' (sedang diajukan)
    if oldData.Status != "submitted" {
        return c.Status(400).JSON(fiber.Map{"error": "Hanya prestasi dengan status 'submitted' yang bisa ditolak"})
    }

    now := time.Now()

    // 3. Update MongoDB: Set status ke 'rejected' dan tambah History
    newHistory := models.AchievementHistory{
        Status:    "rejected",
        ChangedBy: claims.Username, // Nama Dosen/Admin
        ChangedAt: now,
        Notes:     "Ditolak: " + input.Reason,
    }

    updateQuery := bson.M{
        "$set":  bson.M{"status": "rejected", "updatedAt": now},
        "$push": bson.M{"history": newHistory},
    }

    oid, _ := primitive.ObjectIDFromHex(mongoID)
    if err := s.MongoRepo.Update(ctx, oid, updateQuery); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Gagal update status di MongoDB"})
    }

    // 4. Update PostgreSQL: Mengisi rejection_note, verified_by, dan status
    // Query ini memastikan kolom [null] di gambar kamu terisi sesuai data dosen
    err = s.PgRepo.UpdateToRejected(ctx, mongoID, claims.ID, input.Reason)
    if err != nil {
        log.Printf("Postgres Sync Error: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Gagal sinkronisasi data ke PostgreSQL"})
    }

    return c.JSON(fiber.Map{
        "message":        "Prestasi berhasil ditolak",
        "status":         "rejected",
        "rejection_note": input.Reason,
        "rejected_by":    claims.ID,
    })
}

// GET /api/v1/achievements/:id/history
func (s *AchievementService) History(c *fiber.Ctx) error {
    ctx := context.Background()
    
    // Ambil data utuh
    data, err := s.MongoRepo.GetByID(ctx, c.Params("id"))
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }

    // Pastikan jika history masih kosong (nil), return array kosong [] bukan null
    historyList := data.History
    if historyList == nil {
        historyList = []models.AchievementHistory{}
    }

    return c.JSON(fiber.Map{
        "id":      c.Params("id"),
        "history": historyList,
    })
}

// POST /api/v1/achievements/:id/attachments
// FR-003 step 2: Upload dokumen pendukung
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    idParam := c.Params("id")
    oid, err := primitive.ObjectIDFromHex(idParam)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // 1. CEK STATUS TERLEBIH DAHULU
    // Pastikan hanya status 'draft' yang bisa tambah lampiran
    oldData, err := s.MongoRepo.GetByID(ctx, idParam)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }
    if oldData.Status != "draft" {
        return c.Status(403).JSON(fiber.Map{"error": "Cannot add attachment to locked achievement"})
    }

    // 2. AMBIL FILE
    file, err := c.FormFile("document")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
    }

    // 3. SIMPAN FILE KE FOLDER
    // Gunakan penamaan yang aman dari karakter aneh
    fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
    path := "./uploads/" + fileName
    
    if err := c.SaveFile(file, path); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
    }

    // 4. BUAT URL PUBLIK (Asumsi server jalan di localhost:8080)
    // Di produksi, base URL ini harus diambil dari Env Variable
    fileURL := fmt.Sprintf("/uploads/%s", fileName)

    att := models.Attachment{
        FileName:   file.Filename,
        FileURL:    fileURL, // Simpan URL-nya, bukan path sistem lokal
        FileType:   file.Header.Get("Content-Type"),
        UploadedAt: time.Now(),
    }

    // 5. UPDATE MONGODB
    if err := s.MongoRepo.AddAttachment(ctx, oid, att); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to update database"})
    }

    // 6. UPDATE POSTGRES (Sinkronkan UpdatedAt)
    s.PgRepo.UpdateTimestamp(ctx, idParam)

    return c.Status(201).JSON(att)
}

// Helper: Fetch details from Mongo based on Postgres References
func (s *AchievementService) fetchFullDataFromMongo(c *fiber.Ctx, refs []models.AchievementReference, total int) error {
	ctx := context.Background()
	var ids []primitive.ObjectID
	for _, r := range refs {
		if id, err := primitive.ObjectIDFromHex(r.MongoAchievementID); err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return c.JSON(fiber.Map{"total": total, "data": []interface{}{}})
	}
	data, _ := s.MongoRepo.FindByIDs(ctx, ids)
	return c.JSON(fiber.Map{"total": total, "data": data})
}
