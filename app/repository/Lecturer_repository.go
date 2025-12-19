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