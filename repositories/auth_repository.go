package repository

import (
	"database/sql"
	"achievements-uas/models"
)

type AuthRepository struct {
	DB *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{DB: db}
}

func (r *AuthRepository) FindByEmail(email string) (*models.User, error) {
	row := r.DB.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users WHERE email=$1
	`, email)

	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepository) GetByUsernameOrEmail(input string) (*models.User, error) {
	row := r.DB.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users WHERE username=$1 OR email=$1 LIMIT 1
	`, input)

	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash,
		&u.FullName, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *AuthRepository) GetProfile(id string) (*models.User, error) {
	row := r.DB.QueryRow(`
		SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at
		FROM users WHERE id=$1
	`, id)

	var u models.User
	if err := row.Scan(&u.ID, &u.Username, &u.Email,
		&u.FullName, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
