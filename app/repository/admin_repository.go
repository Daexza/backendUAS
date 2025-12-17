package repository

import (
	"database/sql"
	"achievements-uas/app/models"
)

type AdminRepository struct {
	DB *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{DB: db}
}

//
// ================= USER =================
//

// CREATE USER
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

// GET ALL USERS
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

// GET USER BY ID
func (r *AdminRepository) GetByID(id string) (*models.User, error) {
	q := `
		SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at
		FROM users WHERE id=$1
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

// UPDATE USER
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

// DELETE USER
func (r *AdminRepository) DeleteUser(id string) error {
	_, err := r.DB.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}

// UPDATE PASSWORD
func (r *AdminRepository) UpdatePassword(userID, hash string) error {
	_, err := r.DB.Exec(`
		UPDATE users SET password_hash=$2, updated_at=NOW()
		WHERE id=$1
	`, userID, hash)
	return err
}

// ASSIGN ROLE
func (r *AdminRepository) AssignRole(userID, roleID string) error {
	_, err := r.DB.Exec(`
		UPDATE users SET role_id=$2, updated_at=NOW()
		WHERE id=$1
	`, userID, roleID)
	return err
}

//
// ================= STUDENT =================
//

// CREATE STUDENT
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

// GET ALL STUDENTS
func (r *AdminRepository) GetAllStudents() ([]models.Student, error) {
	q := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID,
			&s.ProgramStudy, &s.AcademicYear,
			&s.AdvisorID, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

// GET STUDENT BY ID
func (r *AdminRepository) GetStudentByID(id string) (*models.Student, error) {
	q := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students WHERE id=$1
	`
	s := &models.Student{}
	err := r.DB.QueryRow(q, id).Scan(
		&s.ID, &s.UserID, &s.StudentID,
		&s.ProgramStudy, &s.AcademicYear,
		&s.AdvisorID, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SET ADVISOR
func (r *AdminRepository) SetAdvisor(studentID, advisorID string) error {
	_, err := r.DB.Exec(`
		UPDATE students SET advisor_id=$2 WHERE id=$1
	`, studentID, advisorID)
	return err
}

// GET STUDENT ACHIEVEMENTS (REFERENCE TABLE)
func (r *AdminRepository) GetStudentAchievements(studentID string) ([]map[string]interface{}, error) {
	q := `
		SELECT id, student_id, achievement_id, status, created_at
		FROM achievement_references
		WHERE student_id=$1
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(q, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []map[string]interface{}
	for rows.Next() {
		var (
			id, stID, achID, status string
			createdAt              string
		)
		if err := rows.Scan(&id, &stID, &achID, &status, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]interface{}{
			"id":             id,
			"student_id":     stID,
			"achievement_id": achID,
			"status":         status,
			"created_at":     createdAt,
		})
	}
	return out, nil
}

//
// ================= LECTURER =================
//

// CREATE LECTURER
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

// GET ALL LECTURERS
func (r *AdminRepository) GetAllLecturers() ([]models.Lecturer, error) {
	q := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Lecturer
	for rows.Next() {
		var l models.Lecturer
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.LecturerID,
			&l.Department, &l.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, nil
}

// GET LECTURER ADVISEES
func (r *AdminRepository) GetLecturerAdvisees(lecturerID string) ([]models.Student, error) {
	q := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE advisor_id=$1
	`
	rows, err := r.DB.Query(q, lecturerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID,
			&s.ProgramStudy, &s.AcademicYear,
			&s.AdvisorID, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}
