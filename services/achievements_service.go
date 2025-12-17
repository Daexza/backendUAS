package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"achievements-uas/app/models"
	"achievements-uas/app/repository"
	"achievements-uas/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService struct {
	MongoRepo   *repository.AchievementMongoRepository
	PgRepo      *repository.AchievementPostgresRepository
	StudentRepo *repository.StudentRepository
}

//
// ====================================================
// FR-003: CREATE ACHIEVEMENT (MAHASISWA)
// ====================================================
//
func (s *AchievementService) Create(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	if claims.Role != "MAHASISWA" {
		return fiber.ErrForbidden
	}

	var payload models.Achievement
	if err := c.BodyParser(&payload); err != nil {
		return fiber.ErrBadRequest
	}

	payload.ID = primitive.NilObjectID
	payload.StudentID = claims.ID
	payload.Status = "draft"
	payload.CreatedAt = time.Now()
	payload.UpdatedAt = time.Now()
	payload.History = []models.AchievementHistory{
		{
			Status:    "draft",
			ChangedBy: claims.ID,
			ChangedAt: time.Now(),
		},
	}

	created, err := s.MongoRepo.Create(ctx, &payload)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	ref := &models.AchievementReference{
		ID:                 primitive.NewObjectID().Hex(),
		StudentID:          claims.ID,
		MongoAchievementID: created.ID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.PgRepo.Create(ctx, ref); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   created,
	})
}

//
// ====================================================
// FR-004: SUBMIT ACHIEVEMENT
// ====================================================
//
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	if claims.Role != "MAHASISWA" {
		return fiber.ErrForbidden
	}

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.ErrBadRequest
	}

	ach, err := s.MongoRepo.FindByID(ctx, objID)
	if err != nil {
		return fiber.ErrNotFound
	}

	if ach.StudentID != claims.ID {
		return fiber.ErrForbidden
	}

	if ach.Status != "draft" {
		return errors.New("only draft achievement can be submitted")
	}

	ach.Status = "submitted"
	ach.UpdatedAt = time.Now()
	ach.History = append(ach.History, models.AchievementHistory{
		Status:    "submitted",
		ChangedBy: claims.ID,
		ChangedAt: time.Now(),
	})

	if err := s.MongoRepo.Update(ctx, ach); err != nil {
		return fiber.ErrInternalServerError
	}

	if err := s.PgRepo.UpdateStatus(ctx, id, "submitted"); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{"status": "submitted"})
}

//
// ====================================================
// FR-005: DELETE ACHIEVEMENT (SOFT)
// ====================================================
//
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	if claims.Role != "MAHASISWA" {
		return fiber.ErrForbidden
	}

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.ErrBadRequest
	}

	ach, err := s.MongoRepo.FindByID(ctx, objID)
	if err != nil {
		return fiber.ErrNotFound
	}

	if ach.StudentID != claims.ID {
		return fiber.ErrForbidden
	}

	if ach.Status != "draft" {
		return errors.New("only draft achievement can be deleted")
	}

	if err := s.MongoRepo.SoftDelete(ctx, objID); err != nil {
		return fiber.ErrInternalServerError
	}

	if err := s.PgRepo.UpdateStatus(ctx, id, "deleted"); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}

//
// ====================================================
// FR-006: LIST ACHIEVEMENTS (ROLE BASED)
// ====================================================
//
func (s *AchievementService) List(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	switch claims.Role {

	case "MAHASISWA":
		data, err := s.MongoRepo.FindByStudentID(ctx, claims.ID)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(data)

	case "DOSEN_WALI":
		students, err := s.StudentRepo.FindByAdvisorID(claims.ID)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		var ids []string
		for _, st := range students {
			ids = append(ids, st.UserID)
		}

		data, err := s.MongoRepo.FindByStudentIDs(ctx, ids)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(data)

	case "ADMIN":
		refs, err := s.PgRepo.GetAll(ctx, "", 100, 0)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		var mongoIDs []primitive.ObjectID
		for _, r := range refs {
			if oid, err := primitive.ObjectIDFromHex(r.MongoAchievementID); err == nil {
				mongoIDs = append(mongoIDs, oid)
			}
		}

		data, err := s.MongoRepo.FindByIDs(ctx, mongoIDs)
		if err != nil {
			return fiber.ErrInternalServerError
		}
		return c.JSON(data)
	}

	return fiber.ErrForbidden
}

//
// ====================================================
// FR-007: VERIFY ACHIEVEMENT
// ====================================================
//
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	if claims.Role != "DOSEN_WALI" {
		return fiber.ErrForbidden
	}

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.ErrBadRequest
	}

	ach, err := s.MongoRepo.FindByID(ctx, objID)
	if err != nil {
		return fiber.ErrNotFound
	}

	if ach.Status != "submitted" {
		return errors.New("only submitted achievement can be verified")
	}

	ach.Status = "verified"
	ach.UpdatedAt = time.Now()
	ach.History = append(ach.History, models.AchievementHistory{
		Status:    "verified",
		ChangedBy: claims.ID,
		ChangedAt: time.Now(),
	})

	if err := s.MongoRepo.Update(ctx, ach); err != nil {
		return fiber.ErrInternalServerError
	}

	if err := s.PgRepo.SetVerified(ctx, id, claims.ID); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{"status": "verified"})
}

//
// ====================================================
// FR-008: REJECT ACHIEVEMENT
// ====================================================
//
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	if claims.Role != "DOSEN_WALI" {
		return fiber.ErrForbidden
	}

	id := c.Params("id")
	note := c.FormValue("note")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.ErrBadRequest
	}

	ach, err := s.MongoRepo.FindByID(ctx, objID)
	if err != nil {
		return fiber.ErrNotFound
	}

	ach.Status = "rejected"
	ach.UpdatedAt = time.Now()
	ach.History = append(ach.History, models.AchievementHistory{
		Status:    "rejected",
		ChangedBy: claims.ID,
		ChangedAt: time.Now(),
		Notes:     note,
	})

	if err := s.MongoRepo.Update(ctx, ach); err != nil {
		return fiber.ErrInternalServerError
	}

	if err := s.PgRepo.SetRejected(ctx, id, note); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{"status": "rejected"})
}

//
// ====================================================
// FR-012: UPLOAD ATTACHMENT
// ====================================================
//
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	if claims.Role != "MAHASISWA" {
		return fiber.ErrForbidden
	}

	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.ErrBadRequest
	}

	ach, err := s.MongoRepo.FindByID(ctx, objID)
	if err != nil {
		return fiber.ErrNotFound
	}

	if ach.StudentID != claims.ID {
		return fiber.ErrForbidden
	}

	if ach.Status != "draft" {
		return errors.New("cannot upload attachment after submit")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.ErrBadRequest
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	path := "./uploads/" + filename

	if err := c.SaveFile(file, path); err != nil {
		return fiber.ErrInternalServerError
	}

	att := models.Attachment{
		FileName:   filename,
		FileURL:    "/uploads/" + filename,
		FileType:   file.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	if err := s.MongoRepo.AddAttachment(ctx, objID, att); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{"status": "attachment uploaded"})
}
// ====================================================
// FR-010 / DETAIL ACHIEVEMENT
// ====================================================
func (s *AchievementService) Detail(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	objID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	ach, err := s.MongoRepo.FindByID(ctx, objID)
	if err != nil {
		return fiber.ErrNotFound
	}

	switch claims.Role {

	case "MAHASISWA":
		if ach.StudentID != claims.ID {
			return fiber.ErrForbidden
		}

	case "DOSEN_WALI":
		students, err := s.StudentRepo.FindByAdvisorID(claims.ID)
		if err != nil {
			return fiber.ErrForbidden
		}

		allowed := false
		for _, s := range students {
			if s.UserID == ach.StudentID {
				allowed = true
				break
			}
		}
		if !allowed {
			return fiber.ErrForbidden
		}

	case "ADMIN":
		// bebas

	default:
		return fiber.ErrForbidden
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   ach,
	})
}
// ====================================================
// FR-003 EXTENSION: UPDATE ACHIEVEMENT (DRAFT)
// ====================================================
func (s *AchievementService) Update(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	if claims.Role != "MAHASISWA" {
		return fiber.ErrForbidden
	}

	objID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	ach, err := s.MongoRepo.FindByID(ctx, objID)
	if err != nil {
		return fiber.ErrNotFound
	}

	if ach.StudentID != claims.ID {
		return fiber.ErrForbidden
	}

	if ach.Status != "draft" {
		return errors.New("only draft achievement can be updated")
	}

	var payload models.Achievement
	if err := c.BodyParser(&payload); err != nil {
		return fiber.ErrBadRequest
	}

	// === FIELD YANG BOLEH DIUPDATE ===
	ach.Type = payload.Type
	ach.Title = payload.Title
	ach.Description = payload.Description
	ach.Details = payload.Details
	ach.Tags = payload.Tags
	ach.Points = payload.Points
	ach.UpdatedAt = time.Now()

	if err := s.MongoRepo.Update(ctx, ach); err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   ach,
	})
}
// ====================================================
// FR-010 / ACHIEVEMENT HISTORY
// ====================================================
func (s *AchievementService) History(c *fiber.Ctx) error {
	ctx := context.Background()
	claims := c.Locals("claims").(*utils.JWTClaims)

	objID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	ach, err := s.MongoRepo.FindByID(ctx, objID)
	if err != nil {
		return fiber.ErrNotFound
	}

	// === Akses sama seperti detail ===
	if claims.Role == "MAHASISWA" && ach.StudentID != claims.ID {
		return fiber.ErrForbidden
	}

	if claims.Role == "DOSEN_WALI" {
		students, _ := s.StudentRepo.FindByAdvisorID(claims.ID)
		allowed := false
		for _, s := range students {
			if s.UserID == ach.StudentID {
				allowed = true
				break
			}
		}
		if !allowed {
			return fiber.ErrForbidden
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   ach.History,
	})
}
