package datastore

import (
	"fmt"
	c "rob/lib/common/constants"
	lh "rob/lib/common/loghelper"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

func UpdateToken(userId int, token string) error {
	var funcName = "datastore/firebase.go:UpdateToken"
	log.WithFields(log.Fields{
		"userId": userId,
		"token":  token,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s='%s'
		WHERE %s = %d`,
		c.UsersTable,
		// Key, value paris
		c.Token, token, c.Id, userId)

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
