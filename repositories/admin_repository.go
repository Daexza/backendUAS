package repository

import (
	"database/sql"
	"achievements-uas/models"
)

type AdminRepository struct {
	DB *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{DB: db}
}

// ------------------------------------------------------
// CREATE USER
// ------------------------------------------------------
func (r *AdminRepository) CreateUser(u *models.User) error {
	q := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())
	`
	_, err := r.DB.Exec(q,
		u.ID, u.Username, u.Email, u.PasswordHash,
		u.FullName, u.RoleID, u.IsActive,
	)
	return err
}

// ------------------------------------------------------
// GET ALL USERS
// ------------------------------------------------------
func (r *AdminRepository) GetAllUsers() ([]models.User, error) {
	q := `
		SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email,
			&u.FullName, &u.RoleID, &u.IsActive,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

// ------------------------------------------------------
// GET USER BY ID
// ------------------------------------------------------
func (r *AdminRepository) GetByID(id string) (*models.User, error) {
	q := `
		SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE id=$1 LIMIT 1
	`
	u := &models.User{}
	err := r.DB.QueryRow(q, id).Scan(
		&u.ID, &u.Username, &u.Email,
		&u.FullName, &u.RoleID, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ------------------------------------------------------
// UPDATE USER INFO
// ------------------------------------------------------
func (r *AdminRepository) UpdateUser(u *models.User) error {
	q := `
		UPDATE users
		SET username=$2, email=$3, full_name=$4, role_id=$5, is_active=$6, updated_at=NOW()
		WHERE id=$1
	`
	_, err := r.DB.Exec(q,
		u.ID, u.Username, u.Email,
		u.FullName, u.RoleID, u.IsActive,
	)
	return err
}

// ------------------------------------------------------
// DELETE USER
// ------------------------------------------------------
func (r *AdminRepository) DeleteUser(id string) error {
	_, err := r.DB.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}

// ------------------------------------------------------
// UPDATE PASSWORD (RESET PASSWORD BY ADMIN)
// ------------------------------------------------------
func (r *AdminRepository) UpdatePassword(userID, hash string) error {
	q := `
		UPDATE users
		SET password_hash=$2, updated_at=NOW()
		WHERE id=$1
	`
	_, err := r.DB.Exec(q, userID, hash)
	return err
}

// ------------------------------------------------------
// ASSIGN ROLE
// ------------------------------------------------------
func (r *AdminRepository) AssignRole(userID, roleID string) error {
	_, err := r.DB.Exec(`
		UPDATE users SET role_id=$2, updated_at=NOW()
		WHERE id=$1
	`, userID, roleID)
	return err
}

// ------------------------------------------------------
// CREATE STUDENT PROFILE
// ------------------------------------------------------
func (r *AdminRepository) CreateStudent(st *models.Student) error {
	q := `
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,NOW())
	`
	_, err := r.DB.Exec(q,
		st.ID, st.UserID, st.StudentID,
		st.ProgramStudy, st.AcademicYear, st.AdvisorID,
	)
	return err
}

// ------------------------------------------------------
// CREATE LECTURER PROFILE
// ------------------------------------------------------
func (r *AdminRepository) CreateLecturer(l *models.Lecturer) error {
	q := `
		INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
		VALUES ($1,$2,$3,$4,NOW())
	`
	_, err := r.DB.Exec(q,
		l.ID, l.UserID, l.LecturerID, l.Department,
	)
	return err
}

// ------------------------------------------------------
// SET ADVISOR FOR STUDENT
// ------------------------------------------------------
func (r *AdminRepository) SetAdvisor(studentID, advisorID string) error {
	_, err := r.DB.Exec(`
		UPDATE students SET advisor_id=$2 WHERE id=$1
	`, studentID, advisorID)
	return err
}
