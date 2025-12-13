package repository

import (
	"database/sql"
	"achievements-uas/app/models"
)

type StudentRepository struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{DB: db}
}

// ======================================================
// FIND STUDENT BY USER ID (Mahasiswa ambil data diri sendiri)
// ======================================================
func (r *StudentRepository) FindByUserID(userID string) (*models.Student, error) {
	row := r.DB.QueryRow(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students 
		WHERE user_id=$1
	`, userID)

	var s models.Student
	err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&s.AdvisorID,
		&s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// ======================================================
// GET ALL ACHIEVEMENTS REFERENCE MILIK STUDENT
// ======================================================
func (r *StudentRepository) GetAchievements(studentID string) ([]models.AchievementReference, error) {
	rows, err := r.DB.Query(`
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at,
		       verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE student_id=$1
		ORDER BY created_at DESC
	`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		results = append(results, ref)
	}

	return results, nil
}

// ======================================================
// GET SINGLE ACHIEVEMENT REFERENCE BY ID (student only)
// ======================================================
func (r *StudentRepository) FindAchievementByID(id, studentID string) (*models.AchievementReference, error) {
	row := r.DB.QueryRow(`
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at,
		       verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE id=$1 AND student_id=$2
		LIMIT 1
	`, id, studentID)

	var ref models.AchievementReference
	err := row.Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &ref, nil
}

// ======================================================
// UPDATE STATUS ACHIEVEMENT (submit / verify / reject)
// ======================================================
func (r *StudentRepository) UpdateStatus(id string, status string, verifiedBy *string, rejectionNote *string) error {
	q := `
		UPDATE achievement_references
		SET status=$2, verified_by=$3, rejection_note=$4, updated_at=NOW()
		WHERE id=$1
	`
	_, err := r.DB.Exec(q, id, status, verifiedBy, rejectionNote)
	return err
}

// ======================================================
// SOFT DELETE ACHIEVEMENT REFERENCE
// ======================================================
func (r *StudentRepository) SoftDelete(id string) error {
	q := `
		UPDATE achievement_references
		SET status='deleted', updated_at=NOW()
		WHERE id=$1
	`
	_, err := r.DB.Exec(q, id)
	return err
}
