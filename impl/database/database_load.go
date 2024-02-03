package database

import "errors"

func (d *DatabaseImpl) loadStmt() error {
	q, err := d.db.Prepare("SELECT id, username, password FROM users JOIN tokens ON users.id = tokens.user_id WHERE tokens.token = $1 AND tokens.expires > DATETIME();")
	if err != nil {
		return errors.New("Error preparing statement userByToken: " + err.Error())
	}
	d.userByToken = q

	q, err = d.db.Prepare("SELECT id, name, description, value FROM roles JOIN user_roles ON roles.id = user_roles.role_id WHERE user_roles.user_id = $1;")
	if err != nil {
		return errors.New("Error preparing statement rolesByUser: " + err.Error())
	}
	d.rolesByUser = q

	q, err = d.db.Prepare("SELECT token FROM tokens JOIN users ON tokens.user_id = users.id WHERE users.username = $1 AND users.password = $2;")
	if err != nil {
		return errors.New("Error preparing statement tokenByUsernamePassword: " + err.Error())
	}
	d.tokenFromLogin = q

	return nil
}
