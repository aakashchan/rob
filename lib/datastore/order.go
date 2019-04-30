// All the database requests related to a Order go here
package datastore

import (
	"fmt"
	c "rob/lib/common/constants"
	lh "rob/lib/common/loghelper"
	"rob/lib/common/types"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

/*
Purpose : Creates an order entry in the Orders table
Input : an order object
Outputs : orderId and errro if any
Remark :
*/
func CreateOrder(newOrder types.Order) (int, error) {
	var funcName = "datastore/order.go:CreateOrder"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	product, err := GetProduct(newOrder.ProductId)
	if err != nil {
		log.Error("Could not fetch Product details", err)
		return -1, err
	}
	newOrder.ProductTitle = product.Title
	newOrder.ProductThumb = product.ThumbNail
	newOrder.TransId = c.UninitiatedId
	newOrder.TransStatus = c.Uninitiated
	newOrder.ShippingStatus = c.Uninitiated
	newOrder.TrackingId = c.Uninitiated
	newOrder.ShippingId = c.UninitiatedId
	newOrder.TimeOfCreation = time.Now().UTC().UnixNano()

	var query string
	query = fmt.Sprintf("INSERT INTO %s(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s) VALUES('%s','%s','%s',%d,%d,%d,%d,%d,%d,%d,'%s',%d,%d,%d,'%s','%s',%d);", c.OrderTable, c.ProductId, c.ProductTitle, c.ProductThumb, c.UserId, c.OrderDate, c.Price, c.Tax, c.ShippingCost, c.Amount, c.TransId, c.TransStatus, c.SaleId, c.AddressId, c.ShippingId, c.ShippingStatus, c.TrackingId, c.TimeOfCreation, newOrder.ProductId, newOrder.ProductTitle, newOrder.ProductThumb, newOrder.UserId, newOrder.OrderDate, newOrder.Price, newOrder.Tax, newOrder.ShippingCost, newOrder.Amount, newOrder.TransId, newOrder.TransStatus, newOrder.SaleId, newOrder.AddressId, newOrder.ShippingId, newOrder.ShippingStatus, newOrder.TrackingId, newOrder.TimeOfCreation)
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

	id, err := res.LastInsertId()
	if err != nil {
		lh.Mysql.ScanError(err)
		return -1, err
	}
	return int(id), err

}

/*
Purpose : Retrieves Order details from the orders table
Input : OrderId
Outputs : Order Object pointer
Remark :
*/

func GetOrder(oId int) (*types.Order, error) {
	var funcName = "datastore/order.go:GetOrder"
	log.WithFields(log.Fields{
		"orderId": oId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select * from %s where Id = %d", c.OrderTable, oId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	var order types.Order
	err = stmt.QueryRow().Scan(&order.Id, &order.ProductId, &order.ProductTitle, &order.ProductThumb, &order.UserId, &order.OrderDate, &order.Price, &order.Tax, &order.ShippingCost, &order.Amount, &order.TransId, &order.TransStatus, &order.SaleId, &order.AddressId, &order.ShippingId, &order.ShippingStatus, &order.TrackingId, &order.TimeOfCreation)

	if err != nil {
		lh.Mysql.ScanError(err)
		return nil, err
	}

	return &order, nil

}

/*
Purpose : Retrieves all orders created by a User
Input : UserId
Outputs : OrderList object pointer and error if any
Remark : Wraps all order under OrderList
*/

func GetUserOrders(userId int) (*types.OrdersList, error) {

	var funcName = "datastore/order.go:MyOrders"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select * from %s where %s=%d", c.OrderTable, c.UserId, userId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		lh.Mysql.ExecError(err)
		return nil, err
	}
	defer rows.Close()
	var orders types.OrdersList

	for rows.Next() {
		var order types.Order
		if err = rows.Scan(&order.Id, &order.ProductId, &order.ProductTitle, &order.ProductThumb, &order.UserId, &order.OrderDate, &order.Price, &order.Tax, &order.ShippingCost, &order.Amount, &order.TransId, &order.TransStatus, &order.SaleId, &order.AddressId, &order.ShippingId, &order.ShippingStatus, &order.TrackingId, &order.TimeOfCreation); err != nil {
			lh.Mysql.ScanError(err)
			continue
		}
		orders.Data = append(orders.Data, order)
	}
	if err = rows.Err(); err != nil {
		log.Error(err)
		return nil, err
	}
	return &orders, nil

}
