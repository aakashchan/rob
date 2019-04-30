package datastore

import (
	"fmt"

	c "rob/lib/common/constants"
	lh "rob/lib/common/loghelper"

	log "github.com/sirupsen/logrus"
)

func GetCacheDetails() ([]string, error) {
	var funcName = "datastore/cache.go:GetCacheDetails"
	log.Debugf("Enter : %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var tableName string
	tableName = c.UrlCacheTable
	query := fmt.Sprintf(`
	Select %s 
	FROM %s
	`, c.Url, tableName)

	lh.Mysql.Query(query)

	rows, err := db.Query(query)
	if err != nil {
		lh.Mysql.ExecError(err)
		return nil, err
	}
	defer rows.Close()

	log.Debugf("Returned UrlCache rows: %v", rows)

	var urls []string

	for rows.Next() {
		var url string

		err = rows.Scan(&url)

		log.Debugf("UrlCache url:%s", url)
		if err != nil {
			log.WithField("ErrMsg", err.Error()).Error("Failed to scan one row of UrlCache. Continuing to next row")
			continue
		}

		urls = append(urls, url)
	}

	return urls, nil

}

func InsertCacheUrl(url string) error {
	var funcName = "datastore/cache.go:GetCacheDetails"
	log.Debugf("Enter : %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
	INSERT INTO %s(%s)
	VALUES ('%s');
	`, c.UrlCacheTable, c.Url, url)

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

func DeleteCacheUrl(url string) error {
	var funcName = "datastore/cache.go:DeletetCacheUrl"
	log.Debugf("Enter %s", funcName)
	defer log.Debugf("Exit %s", funcName)

	query := fmt.Sprintf(`
	DELETE FROM %s
	WHERE %s = '%s';
	`, c.UrlCacheTable, c.Url, url)

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
