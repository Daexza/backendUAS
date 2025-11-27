package repository

import (
	"database/sql"
	"achievements-uas/models"
)

type LecturerRepository struct {
	DB *sql.DB
}

func NewLecturerRepository(db *sql.DB) *LecturerRepository {
	return &LecturerRepository{DB: db}
}

func (r *LecturerRepository) Create(l models.Lecturer) error {
	_, err := r.DB.Exec(`
		INSERT INTO lecturers (
			id, user_id, lecturer_id, department, created_at
		)
		VALUES ($1,$2,$3,$4,NOW())
	`,
		l.ID, l.UserID, l.LecturerID, l.Department,
	)
	return err
}

func (r *LecturerRepository) Update(l models.Lecturer) error {
	_, err := r.DB.Exec(`
		UPDATE lecturers 
		SET lecturer_id=$1, department=$2
		WHERE id=$3
	`,
		l.LecturerID, l.Department, l.ID,
	)
	return err
}

func (r *LecturerRepository) Delete(id string) error {
	_, err := r.DB.Exec(`DELETE FROM lecturers WHERE id=$1`, id)
	return err
}

func (r *LecturerRepository) GetAdvisees(lecturerID string) ([]string, error) {
	rows, err := r.DB.Query(`
		SELECT id 
		FROM students 
		WHERE advisor_id=$1
	`, lecturerID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string

	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}

	return ids, nil
}
