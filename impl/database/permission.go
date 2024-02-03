package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/Fesaa/Media-Provider/models"
)

type PermissionImpl struct {
	permisions []*models.Permission

	refreshQuery *sql.Stmt
}

func newPermission(pool *sql.DB) (*PermissionImpl, error) {
	q, err := pool.Prepare("SELECT key, description, value FROM permissions")
	if err != nil {
		return nil, err
	}
	p := &PermissionImpl{
		permisions:   make([]*models.Permission, 0),
		refreshQuery: q,
	}

	err = p.RefreshPermissions()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *PermissionImpl) GetAllPermissions() []*models.Permission {
	return p.permisions
}

func (p *PermissionImpl) GetPermissionByKey(key string) *models.Permission {
	index := slices.IndexFunc(p.permisions, func(p *models.Permission) bool {
		return p.Key() == key
	})
	if index == -1 {
		return nil
	}
	return p.permisions[index]
}

func (p *PermissionImpl) UserHasPermission(u *models.User, key string) bool {
	perm := p.GetPermissionByKey(key)
	if perm == nil {
		return false
	}
	return u.HasPermission(perm)
}

func (p *PermissionImpl) RoleHasPermission(r *models.Role, key string) bool {
	perm := p.GetPermissionByKey(key)
	if perm == nil {
		return false
	}
	return r.HasPermission(perm.Value())
}

func (p *PermissionImpl) RefreshPermissions() error {
	slog.Info("Reloading permissions from database")
	start := time.Now()
	rows, err := p.refreshQuery.Query()
	if err != nil {
		return err
	}

	p.permisions = make([]*models.Permission, 0)

	for rows.Next() {
		var key, description string
		var value int64
		err = rows.Scan(&key, &description, &value)
		if err != nil {
			return err
		}
		p.permisions = append(p.permisions, models.NewPermission(key, description, value))
	}

	slog.Info(fmt.Sprintf("Loaded %d permissions in %f s", len(p.permisions), time.Since(start).Seconds()))
	return nil
}
