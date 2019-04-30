// All the database requests related to a user go here
package datastore

import (
	"fmt"
	c "rob/lib/common/constants"
	"rob/lib/common/types"
	"time"

	lh "rob/lib/common/loghelper"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

// Accepts a user object instantiated with form data and inserts the data into Users table
func AddUser(newUser types.User) error {
	var funcName = "datastore/user.go:AddUser"
	log.WithFields(log.Fields{
		"newUser": newUser,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var timeOfCreation = time.Now().UTC().UnixNano()

	query := fmt.Sprintf(`
		INSERT INTO %s(%s,%s,%s,%s,%s,%s,%s,%s,%s)
		VALUES('%s','%s','%s','%s','%s','%s',%d,%d,%s)`,
		c.UsersTable, c.Email, c.Password, c.Gender, c.FirstName,
		c.LastName, c.Phone, c.TimeOfCreation, c.Verified, c.Code, newUser.Email,
		newUser.Password, newUser.Gender.String, newUser.FirstName.String,
		newUser.LastName.String, newUser.Phone.String, timeOfCreation, newUser.Verified, newUser.Code)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}

func UpdateUserDetails(user types.User) error {
	var funcName = "datastore/user.go:UpdateUserProfile"
	log.WithFields(log.Fields{
		"user": user,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s='%s',%s='%s',%s='%s',%s='%s', %s='%d'
		WHERE %s = %d`,
		c.UsersTable,
		// Key, value paris
		c.FirstName, user.FirstName.String,
		c.LastName, user.LastName.String,
		c.Gender, user.Gender.String,
		c.Password, user.Password,
		c.Verified, user.Verified,
		c.Id, user.Id)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}
func UpdateUserProfile(user types.User) error {
	var funcName = "datastore/user.go:UpdateUserProfile"
	log.WithFields(log.Fields{
		"user": user,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s='%s',%s='%s',%s='%s'
		WHERE %s = %d`,
		c.UsersTable,
		// Key, value paris
		c.FirstName, user.FirstName.String,
		c.LastName, user.LastName.String,
		c.Gender, user.Gender.String,
		c.Id, user.Id)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}

func GetUserByPhone(phone string) (*types.User, error) {
	var funcName = "datastore/user.go:GetUserByPhone"
	log.WithFields(log.Fields{
		"phone": phone,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		SELECT %s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s
		FROM %s
		WHERE %s = '%s'`,
		c.Id, c.Email, c.Password, c.Gender, c.FirstName, c.LastName, c.Phone, c.TimeOfCreation, c.Verified, c.Code, c.ResetPasswordToken,
		c.UsersTable,
		c.Phone, phone)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	var u types.User
	err = stmt.QueryRow().Scan(&u.Id, &u.Email, &u.Password, &u.Gender, &u.FirstName, &u.LastName, &u.Phone, &u.TimeOfCreation, &u.Verified, &u.Code, &u.ResetPasswordToken)

	if err != nil {
		// Removing this statement as it will repeat everytime a user signup happens
		// lh.Mysql.ScanError(err)
		return nil, err
	}

	return &u, nil
}

func GetUserByEmail(email string) (*types.User, error) {
	var funcName = "datastore/user.go:GetUserByEmail"
	log.WithFields(log.Fields{
		"email": email,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		SELECT %s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s
		FROM %s
		WHERE %s = '%s'`,
		c.Id, c.Email, c.Password, c.Gender, c.FirstName, c.LastName, c.Phone, c.TimeOfCreation, c.Verified, c.Code, c.ResetPasswordToken,
		c.UsersTable,
		c.Email, email)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	var u types.User
	err = stmt.QueryRow().Scan(&u.Id, &u.Email, &u.Password, &u.Gender, &u.FirstName, &u.LastName, &u.Phone, &u.TimeOfCreation, &u.Verified, &u.Code, &u.ResetPasswordToken)

	if err != nil {
		// Removing this statement as it will repeat everytime a user signup happens
		// lh.Mysql.ScanError(err)
		return nil, err
	}

	return &u, nil
}

func GetRole(userId int) (*int, error) {
	var funcName = "datastore/user.go:GetRole"
	log.WithFields(log.Fields{
		"userId": userId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?`,
		c.RoleId,
		c.UserRoleTable,
		c.UserId)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	var roleId int
	err = stmt.QueryRow(userId).Scan(&roleId)
	if err != nil {
		lh.Mysql.ScanError(err)
		return nil, err
	}

	return &roleId, nil
}

func InsertRole(userId int, roleId int) error {
	var funcName = "datastore/user.go:InsertRole"
	log.WithFields(log.Fields{
		"userId": userId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		INSERT INTO %s(%s,%s)
		VALUES(%d,%d);`,
		c.UserRoleTable, c.UserId, c.RoleId,
		userId, roleId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		lh.Mysql.ExecError(err)
		return err
	}
	defer stmt.Close()

	return nil
}

func UpdateRole(userId int, roleId int) error {
	var funcName = "datastore/user.go:UpdateRole"
	log.WithFields(log.Fields{
		"userId": userId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s=%d
		WHERE %s=%d`,
		c.UserRoleTable,
		c.RoleId, roleId,
		c.UserId, userId)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		lh.Mysql.ExecError(err)
		return err
	}
	defer stmt.Close()

	return nil
}

func UpdatePasswordResetToken(phone, token string) error {
	var funcName = "datastore/user.go:UpdatePasswordResetToken"
	log.WithFields(log.Fields{
		"phone": phone,
		"token": token,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s='%s'
		WHERE %s='%s'`,
		c.UsersTable,
		c.ResetPasswordToken, token,
		c.Phone, phone)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		lh.Mysql.ExecError(err)
		return err
	}

	return nil
}

func ResetPassword(phone, newPassword string) error {
	var funcName = "datastore/user.go:ResetPassword"
	log.WithFields(log.Fields{
		"phone":        phone,
		"new Password": newPassword,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s='%s', %s='%s'
		WHERE %s='%s'`,
		c.UsersTable,
		c.Password, newPassword,
		c.ResetPasswordToken, "",
		c.Phone, phone)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		lh.Mysql.ExecError(err)
		return err
	}

	return nil
}

func DeleteUserByEmail(email string) error {
	var funcName = "datastore/user.go:DeleteUserByEmail"
	log.WithFields(log.Fields{
		"email": email,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE %s = '%s'`,
		c.UsersTable,
		c.Email,
		email)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		lh.Mysql.ExecError(err)
		return err
	}
	defer stmt.Close()

	return nil
}
