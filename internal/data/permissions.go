package data

import (
	"context"
	"database/sql"
	"time"
)

type Permissions []string

func (permissions Permissions) Includes(code string) bool {
	for i := range permissions {
		if code == permissions[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *sql.DB
}

func (model PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	sqlQuery := `
SELECT permissions.code
FROM permissions
INNER JOIN users_permissions ON users_permissions.permission_id = permissions.permission_id
INNER JOIN users ON users.id = users_permissions.user_id
WHERE users.id = $1
  `

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := model.DB.QueryContext(ctx, sqlQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
