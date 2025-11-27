package repository

import (
	"database/sql"
	"achievements-uas/models"
)

type RolePermissionRepository struct {
	DB *sql.DB
}

func NewRolePermissionRepository(db *sql.DB) *RolePermissionRepository {
	return &RolePermissionRepository{DB: db}
}

func (r *RolePermissionRepository) Assign(roleID, permissionID string) error {
	_, err := r.DB.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
	`, roleID, permissionID)
	return err
}

func (r *RolePermissionRepository) Remove(roleID, permissionID string) error {
	_, err := r.DB.Exec(`
		DELETE FROM role_permissions WHERE role_id=$1 AND permission_id=$2
	`, roleID, permissionID)
	return err
}

func (r *RolePermissionRepository) GetPermissionsByRole(roleID string) ([]models.Permission, error) {
	rows, err := r.DB.Query(`
		SELECT p.id, p.name, p.resource, p.action, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Permission

	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description); err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	return list, nil
}
