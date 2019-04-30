// All the database requests related to a Shipping go here
package datastore

import (
	"errors"
	"fmt"
	c "rob/lib/common/constants"
	lh "rob/lib/common/loghelper"
	"rob/lib/common/types"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

/*
Purpose : Places an order for shipping to be  initiated
Input : Shipping object
Outputs : ShippingId and error if any
Remark : To be included as a SQL Transaction along with other transactional functions
*/

func PlaceOrder(ship types.Shipping) (int, error) {
	var funcName = "datastore/shipping.go:PlaceOrder"
	log.WithFields(log.Fields{
		"shipping": ship,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	ship.TrackingId = c.Uninitiated
	ship.ShippingStatus = c.Uninitiated
	var query string
	timeofCreation := time.Now().UTC().UnixNano()
	query = fmt.Sprintf("INSERT INTO %s(%s,%s,%s,%s,%s,%s) VALUES(%d,%d,'%s',%d,'%s',%d);", c.ShippingTable, c.OrderId, c.UserId, c.TrackingId, c.AddressId, c.ShippingStatus, c.TimeOfCreation, ship.OrderId, ship.UserId, ship.TrackingId, ship.AddressId, ship.ShippingStatus, timeofCreation)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return -1, err
	}
	defer stmt.Close()
	res, err := stmt.Exec()
	if err != nil {
		lh.Mysql.ExecError(err)
		return -1, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return -1, errors.New("Coudnt retirve last inserted id ")
	}
	return int(lastId), nil

}
