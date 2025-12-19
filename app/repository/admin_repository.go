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

// CreateStudentWithUser: Menyimpan ke tabel users dan students sekaligus
func (r *AdminRepository) CreateStudent(u *models.User, st *models.Student) error {
    tx, err := r.DB.Begin()
    if err != nil { return err }

    // 1. Simpan ke tabel users [cite: 33-36]
    qUser := `INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
    _, err = tx.Exec(qUser, u.ID, u.Username, u.Email, u.PasswordHash, u.FullName, u.RoleID, u.IsActive, u.CreatedAt, u.UpdatedAt)
    if err != nil { tx.Rollback(); return err }

    // 2. Simpan ke tabel students [cite: 85-91]
    // advisor_id dikosongkan ($6 = NULL) karena diatur nanti oleh Admin
    qStudent := `INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id, created_at)
                 VALUES ($1, $2, $3, $4, $5, NULL, $6)`
    _, err = tx.Exec(qStudent, st.ID, u.ID, st.StudentID, st.ProgramStudy, st.AcademicYear, st.CreatedAt)
    if err != nil { tx.Rollback(); return err }

    return tx.Commit()
}

// CreateLecturerWithUser: Menyimpan ke tabel users dan lecturers sekaligus
func (r *AdminRepository) CreateLecturer(u *models.User, l *models.Lecturer) error {
    tx, err := r.DB.Begin()
    if err != nil { return err }

    // 1. Simpan ke tabel users [cite: 33-36]
    qUser := `INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
    _, err = tx.Exec(qUser, u.ID, u.Username, u.Email, u.PasswordHash, u.FullName, u.RoleID, u.IsActive, u.CreatedAt, u.UpdatedAt)
    if err != nil { tx.Rollback(); return err }

    // 2. Simpan ke tabel lecturers [cite: 98-103]
    qLecturer := `INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
                  VALUES ($1, $2, $3, $4, $5)`
    _, err = tx.Exec(qLecturer, l.ID, u.ID, l.LecturerID, l.Department, l.CreatedAt)
    if err != nil { tx.Rollback(); return err }

    return tx.Commit()
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
// GET STUDENT BY USER ID (Digunakan saat Create Achievement)
func (r *AdminRepository) GetStudentByUserID(userID string) (*models.Student, error) {
    q := `
        SELECT 
            s.id, s.user_id, s.student_id, s.program_study, s.academic_year, 
            s.advisor_id, s.created_at, u.full_name, u.email
        FROM students s
        JOIN users u ON s.user_id = u.id
        WHERE s.user_id=$1
    `
    s := &models.Student{}
    var advisorID sql.NullString 

    err := r.DB.QueryRow(q, userID).Scan(
        &s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear,
        &advisorID, &s.CreatedAt, &s.FullName, &s.Email,
    )

    if err != nil {
        return nil, err
    }

    if advisorID.Valid {
        s.AdvisorID = advisorID.String
    }

    return s, nil
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

func (r *AdminRepository) DeleteUser(id string) error {
    tx, err := r.DB.Begin()
    if err != nil {
        return err
    }

    // 1. Hapus dari tabel students (jika ada) [cite: 85-88]
    _, err = tx.Exec(`DELETE FROM students WHERE user_id=$1`, id)
    if err != nil {
        tx.Rollback()
        return err
    }

    // 2. Hapus dari tabel lecturers (jika ada) [cite: 98-101]
    _, err = tx.Exec(`DELETE FROM lecturers WHERE user_id=$1`, id)
    if err != nil {
        tx.Rollback()
        return err
    }

    // 3. Baru hapus dari tabel users [cite: 33-34]
    _, err = tx.Exec(`DELETE FROM users WHERE id=$1`, id)
    if err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
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

// // CREATE STUDENT
// func (r *AdminRepository) CreateStudent(st *models.Student) error {
// 	q := `
// 		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id, created_at)
// 		VALUES ($1,$2,$3,$4,$5,$6,NOW())
// 	`
// 	_, err := r.DB.Exec(q,
// 		st.ID, st.UserID, st.StudentID,
// 		st.ProgramStudy, st.AcademicYear, st.AdvisorID,
// 	)
// 	return err
// }

func (r *AdminRepository) GetAllStudents() ([]models.Student, error) {
	// 1. Tulis query secara eksplisit (sebutkan nama kolom satu per satu)
	// Query ini menggabungkan tabel students dan users untuk mendapatkan data lengkap
	query := `
		SELECT 
			s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.created_at,
			u.full_name, u.email
		FROM students s
		JOIN users u ON s.user_id = u.id
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student

	for rows.Next() {
		var s models.Student
		// Pastikan Scan menerima alamat (&) variabel sesuai urutan SELECT di atas
		err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.CreatedAt,
			&s.FullName, &s.Email, // Pastikan struct models.Student punya field ini
		)
		if err != nil {
			return nil, err
		}
		students = append(students, s)
	}

	// Cek jika ada error selama iterasi
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

// GET STUDENT BY ID
func (r *AdminRepository) GetStudentByID(id string) (*models.Student, error) {
    // Kita tambahkan u.full_name dan u.email
    q := `
        SELECT 
            s.id, s.user_id, s.student_id, s.program_study, s.academic_year, 
            s.advisor_id, s.created_at, u.full_name, u.email
        FROM students s
        JOIN users u ON s.user_id = u.id
        WHERE s.id=$1
    `
    s := &models.Student{}
    
    // Gunakan sql.NullString untuk advisor_id karena mahasiswa baru mungkin belum punya pembimbing
    var advisorID sql.NullString 

    err := r.DB.QueryRow(q, id).Scan(
        &s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear,
        &advisorID, &s.CreatedAt, &s.FullName, &s.Email,
    )

    if err != nil {
        return nil, err
    }

    // Pindahkan nilai dari NullString ke struct
    if advisorID.Valid {
        s.AdvisorID = advisorID.String
    }

    return s, nil
}

// SET ADVISOR
func (r *AdminRepository) SetAdvisor(studentID, advisorID string) error {
    // 1. Pastikan kolom advisor_id bisa menerima string/UUID yang dikirim
    // 2. Tambahkan log untuk melihat data yang masuk ke query
    query := `UPDATE students SET advisor_id=$2 WHERE id=$1`
    
    _, err := r.DB.Exec(query, studentID, advisorID)
    if err != nil {
        // Ini akan memberitahu jika ada error Foreign Key atau Tipe Data
        return err 
    }
    return nil
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

// // CREATE LECTURER
// func (r *AdminRepository) CreateLecturer(l *models.Lecturer) error {
// 	q := `
// 		INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
// 		VALUES ($1,$2,$3,$4,NOW())
// 	`
// 	_, err := r.DB.Exec(q,
// 		l.ID, l.UserID, l.LecturerID, l.Department,
// 	)
// 	return err
// }

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
func (r *AdminRepository) CheckIsAdvisee(studentID, lecturerID string) (bool, error) {
    var exists bool
    // Sesuaikan nama kolom 'advisor_id' dengan nama kolom di tabel students milikmu
    query := `SELECT EXISTS(SELECT 1 FROM students WHERE id = $1 AND advisor_id = $2)`
    
    err := r.DB.QueryRow(query, studentID, lecturerID).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}
// GET LECTURER ADVISEES
func (r *AdminRepository) GetLecturerAdvisees(lecturerID string) ([]models.Student, error) {
    q := `
        SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
               u.full_name, u.email
        FROM students s
        JOIN users u ON s.user_id = u.id
        WHERE s.advisor_id = $1
    `
    rows, err := r.DB.Query(q, lecturerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var list []models.Student
    for rows.Next() {
        var s models.Student
        // Pastikan struct models.Student sudah punya field FullName dan Email
        if err := rows.Scan(
            &s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear,
            &s.AdvisorID, &s.CreatedAt, &s.FullName, &s.Email,
        ); err != nil {
            return nil, err
        }
        list = append(list, s)
    }
    return list, nil

	
}
