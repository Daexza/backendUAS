package repository

import (
	"database/sql"
	"achievements-uas/models"
)

type PermissionRepository struct {
	DB *sql.DB
}

func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{DB: db}
}

func (r *PermissionRepository) FindAll() ([]models.Permission, error) {
	rows, err := r.DB.Query(`
		SELECT id, name, resource, action, description FROM permissions ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.Permission{}
	for rows.Next() {
		var p models.Permission
		rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description)
		list = append(list, p)
	}
	return list, nil
}

func (r *PermissionRepository) FindByID(id string) (*models.Permission, error) {
	row := r.DB.QueryRow(`
		SELECT id, name, resource, action, description 
		FROM permissions WHERE id=$1
	`, id)

	var p models.Permission
	if err := row.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description); err != nil {
		return nil, err
	}
	return &p, nil
}
