package repository

import (
	"database/sql"
)

type RolePermissionRepository struct {
	DB *sql.DB
}

func NewRolePermissionRepository(db *sql.DB) *RolePermissionRepository {
	return &RolePermissionRepository{DB: db}
}

func (r *RolePermissionRepository) GetPermissionsByRole(roleID string) ([]string, error) {
	rows, err := r.DB.Query(`
		SELECT p.name
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var p string
		rows.Scan(&p)
		perms = append(perms, p)
	}
	return perms, nil
}
