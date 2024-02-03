package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/Fesaa/Media-Provider/cerrors"
	"github.com/Fesaa/Media-Provider/models"
)

type DatabaseImpl struct {
	db *sql.DB

	permissionProvider *PermissionImpl

	userByToken    *sql.Stmt
	rolesByUser    *sql.Stmt
	tokenFromLogin *sql.Stmt
}

func NewDatabase(pool *sql.DB) (models.DatabaseProvider, error) {
	d := &DatabaseImpl{
		db: pool,
	}
	err := d.loadTables()
	if err != nil {
		return nil, err
	}

	err = d.loadStmt()
	if err != nil {
		return nil, err
	}

	p, err := newPermission(pool)
	if err != nil {
		return nil, err
	}

	d.permissionProvider = p
	return d, nil
}

func (d *DatabaseImpl) GetPool() *sql.DB {
	return d.db
}

func (d *DatabaseImpl) GetUser(token string) (*models.User, error) {
	row := d.userByToken.QueryRow(token)

	var id int
	var username, password string
	err := row.Scan(&id, &username, &password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cerrors.InvalidCredentials
		}
		return nil, err
	}

	roles, err := d.getRoles(id)
	if err != nil {
		return nil, err
	}

	user := models.NewUser(id, username, password, roles)
	return user, nil
}

func (d *DatabaseImpl) getRoles(userId int) ([]models.Role, error) {
	rows, err := d.rolesByUser.Query(userId)
	if err != nil {
		return nil, err
	}

	roles := make([]models.Role, 0)
	for rows.Next() {
		var id int
		var name string
		var description string
		var value int64
		err = rows.Scan(&id, &name, &description, &value)
		if err != nil {
			return nil, err
		}
		roles = append(roles, models.NewRole(id, name, description, value))
	}

	return roles, nil
}

func (d *DatabaseImpl) GetToken(username, password string) (*string, error) {
	rows := d.tokenFromLogin.QueryRow(username, password)

	var t string
	err := rows.Scan(&t)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cerrors.InvalidCredentials
		}
		return nil, err
	}

	return &t, nil
}

// TODO: Cleanup this mess
func (d *DatabaseImpl) CreateUser(username, password string) (*models.User, *string, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, nil, err
	}

	stmt, err := tx.Prepare("INSERT INTO users (username, password) VALUES ($1, $2)")
	if err != nil {
		return nil, nil, err
	}
	defer stmt.Close()

	r, err := stmt.Exec(username, password)
	if err != nil {
		return nil, nil, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return nil, nil, err
	}

	stmt, err = tx.Prepare("INSERT INTO tokens (user_id, token, expires) VALUES ($1, $2, $3)")
	if err != nil {
		return nil, nil, err
	}
	defer stmt.Close()

	token := generateSecureToken(32)
	_, err = stmt.Exec(id, token, time.Now().Add(time.Hour*24*7))
	if err != nil {
		return nil, nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, err
	}

	roles := make([]models.Role, 0)
	user := models.NewUser(int(id), username, password, roles)
	return user, &token, nil
}

func (d *DatabaseImpl) GetPermissionProvider() models.PermissionProvider {
	return d.permissionProvider
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
