package repository

import (
	"database/sql"
	"achievements-uas/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// =============================
// FIND BY EMAIL
// =============================
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `
	SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
	FROM users WHERE email = $1 LIMIT 1;
	`
	u := &models.User{}
	err := r.DB.QueryRow(query, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// =============================
// GET PROFILE BY ID
// =============================
func (r *UserRepository) GetProfile(id string) (*models.User, error) {
	query := `
	SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at 
	FROM users 
	WHERE id = $1 LIMIT 1;
	`
	u := &models.User{}
	err := r.DB.QueryRow(query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// =============================
// GET PERMISSIONS BY ROLE
// =============================
func (r *UserRepository) GetPermissionsByRole(roleID string) ([]string, error) {
	query := `
	SELECT p.name
	FROM permissions p
	JOIN role_permissions rp ON rp.permission_id = p.id
	WHERE rp.role_id = $1;
	`
	rows, err := r.DB.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		perms = append(perms, name)
	}
	return perms, nil
}

// =============================
// GET USER BY USERNAME OR EMAIL
// =============================
func (r *UserRepository) GetByUsernameOrEmail(input string) (*models.User, error) {
	query := `
	SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
	FROM users 
	WHERE username = $1 OR email = $1
	LIMIT 1;
	`
	u := &models.User{}
	err := r.DB.QueryRow(query, input).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}
