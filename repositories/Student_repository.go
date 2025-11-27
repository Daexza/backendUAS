package repository

import (
	"database/sql"
	"achievements-uas/models"
)

type StudentRepository struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{DB: db}
}

func (r *StudentRepository) Create(s models.Student) error {
	_, err := r.DB.Exec(`
		INSERT INTO students (
			id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,NOW())
	`,
		s.ID, s.UserID, s.StudentID, s.ProgramStudy, s.AcademicYear, s.AdvisorID,
	)
	return err
}

func (r *StudentRepository) Update(s models.Student) error {
	_, err := r.DB.Exec(`
		UPDATE students 
		SET student_id=$1, program_study=$2, academic_year=$3, advisor_id=$4
		WHERE id=$5
	`,
		s.StudentID, s.ProgramStudy, s.AcademicYear, s.AdvisorID, s.ID,
	)
	return err
}

func (r *StudentRepository) Delete(id string) error {
	_, err := r.DB.Exec(`DELETE FROM students WHERE id=$1`, id)
	return err
}

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
