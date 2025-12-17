package repository

import (
	"database/sql"

	"achievements-uas/app/models"
)

type LecturerRepository struct {
	DB *sql.DB
}

func NewLecturerRepository(db *sql.DB) *LecturerRepository {
	return &LecturerRepository{DB: db}
}

// ======================================================
// GET LECTURER BY USER ID
// ======================================================
func (r *LecturerRepository) FindByUserID(userID string) (*models.Lecturer, error) {
	row := r.DB.QueryRow(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE user_id=$1
	`, userID)

	var l models.Lecturer
	if err := row.Scan(
		&l.ID,
		&l.UserID,
		&l.LecturerID,
		&l.Department,
		&l.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &l, nil
}

// ======================================================
// GET ADVISEES
// ======================================================
func (r *LecturerRepository) GetAdvisees(
	lecturerID string,
) ([]models.Student, error) {

	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study,
		       academic_year, advisor_id, created_at
		FROM students
		WHERE advisor_id=$1
	`, lecturerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		var s models.Student
		_ = rows.Scan(
			&s.ID,
			&s.UserID,
			&s.StudentID,
			&s.ProgramStudy,
			&s.AcademicYear,
			&s.AdvisorID,
			&s.CreatedAt,
		)
		list = append(list, s)
	}

	return list, nil
}

// ======================================================
// CHECK IS ADVISEE
// ======================================================
func (r *LecturerRepository) IsAdvisee(
	lecturerID string,
	studentID string,
) (bool, error) {

	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM students
		WHERE advisor_id=$1 AND id=$2
	`, lecturerID, studentID).Scan(&count)

	return count > 0, err
}

// ======================================================
// GET ACHIEVEMENT REFERENCES
// ======================================================
func (r *LecturerRepository) GetAchievementsByStudentIDs(
	studentIDs []string,
) ([]models.AchievementReference, error) {

	if len(studentIDs) == 0 {
		return []models.AchievementReference{}, nil
	}

	rows, err := r.DB.Query(`
		SELECT id, student_id, mongo_achievement_id, status,
		       submitted_at, verified_at, verified_by,
		       rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE student_id = ANY($1)
		ORDER BY created_at DESC
	`, studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		_ = rows.Scan(
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
// UPDATE STATUS
// ======================================================
func (r *LecturerRepository) UpdateStatus(
	id string,
	status string,
	verifiedBy string,
	rejectionNote *string,
) error {

	_, err := r.DB.Exec(`
		UPDATE achievement_references
		SET status=$2,
		    verified_by=$3,
		    rejection_note=$4,
		    updated_at=NOW()
		WHERE id=$1
	`, id, status, verifiedBy, rejectionNote)

	return err
}
