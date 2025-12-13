package repository

import (
	"database/sql"
	"achievements-uas/app/models"
)

type RoleRepository struct {
	DB *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{DB: db}
}

func (r *RoleRepository) FindAll() ([]models.Role, error) {
	rows, err := r.DB.Query(`
		SELECT id, name, description, created_at FROM roles ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.Role{}
	for rows.Next() {
		var role models.Role
		rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
		list = append(list, role)
	}
	return list, nil
}

func (r *RoleRepository) FindByID(id string) (*models.Role, error) {
	row := r.DB.QueryRow(`
		SELECT id, name, description, created_at FROM roles WHERE id=$1
	`, id)

	var role models.Role
	if err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
		return nil, err
	}
	return &role, nil
}
