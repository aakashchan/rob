// All the database requests related to a Sales go here
package datastore

import (
	"errors"
	"fmt"
	c "rob/lib/common/constants"
	lh "rob/lib/common/loghelper"
	"rob/lib/common/types"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

/*
Purpose : To add a sale entry
Input : Sale object
Outputs : saleId. error
Remark :
*/
func AddSale(newSale types.Sale) (int, error) {
	var funcName = "datastore/sale.go:AddSale"
	log.WithFields(log.Fields{
		"newSale": newSale,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	timeofCreation := time.Now().UTC().UnixNano()
	query = fmt.Sprintf("INSERT INTO %s(%s,%s,%s,%s,%s,%s,%s,%s,%s) VALUES('%s','%s','%s','%s','%s',%d,%d,%d,%d);", c.SaleTable, c.Title, c.Brand, c.ProductSku, c.Description, c.ThumbNail, c.StockUnits, c.SaleStartTime, c.SaleEndTime, c.TimeOfCreation, newSale.Title, newSale.Brand, newSale.ProductSku, newSale.Description, newSale.ThumbNail, newSale.StockUnits, newSale.SaleStartTime, newSale.SaleEndTime, timeofCreation)
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

/*
Purpose : To retrieves all details from a sale entry
Input : saleId
Outputs : a sale object and error if any
Remark : Return a single sale entry
*/
func GetSale(sId int) (*types.Sale, error) {
	var funcName = "datastore/sale.go:GetSale"
	log.WithFields(log.Fields{
		"salesId": sId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select Id,Title,Brand,ProductSku,Description,ThumbNail,StockUnits,SaleStartTime,SaleEndTime from Sale where Id = %d", sId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	var sale types.Sale
	err = stmt.QueryRow().Scan(&sale.Id, &sale.Title, &sale.Brand, &sale.ProductSku, &sale.Description, &sale.ThumbNail, &sale.StockUnits, &sale.SaleStartTime, &sale.SaleEndTime)

	if err != nil {
		lh.Mysql.ScanError(err)
		return nil, err
	}

	return &sale, nil

}

/*
Purpose : Returns all sales listed in Twiq
Input : null
Outputs : SaleList object
Remark : wraps all sales into salelist
*/
func GetSales() (*types.SalesList, error) {
	var funcName = "datastore/sale.go:GetSales"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select Title,Brand,ProductSku,Description,ThumbNail,StockUnits,SaleStartTime,SaleEndTime from %s", c.SaleTable)
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
	var sales types.SalesList

	for rows.Next() {
		var sale types.Sale
		if err = rows.Scan(&sale.Title, &sale.Brand, &sale.ProductSku, &sale.Description, &sale.ThumbNail, &sale.StockUnits, &sale.SaleStartTime, &sale.SaleEndTime); err != nil {
			lh.Mysql.ScanError(err)
			continue
		}
		sales.Data = append(sales.Data, sale)
	}
	if err = rows.Err(); err != nil {
		log.Error(err)
		return nil, err
	}
	return &sales, nil
}

/*
Purpose : Updates the Stock attribute in sale entry
Input : saleid and the value to be incremeted
Outputs : error if any
Remark : To increment pass +ve value , to decrement pass -ve
*/
func UpdateSaleStock(saleId int, value int) error {
	var funcName = "datastore/sale.go:UpdateSaleStock"
	log.WithFields(log.Fields{
		"saleId": saleId,
		"value":  value,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select %s from %s where Id = %d", c.StockUnits, c.SaleTable, saleId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()
	var stock int
	err = stmt.QueryRow().Scan(&stock)

	if err != nil {
		lh.Mysql.ScanError(err)
		return err
	}
	if stock < 1 {
		return errors.New("Product Out Of Stock")
	}
	query = fmt.Sprintf("UPDATE %s SET %s = %s + %d WHERE %s = %d ;", c.SaleTable, c.StockUnits, c.StockUnits, value, c.Id, saleId)
	lh.Mysql.Query(query)

	stmt, err = db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()
	var mutex = &sync.Mutex{}
	mutex.Lock()
	_, err = stmt.Exec()
	mutex.Unlock()
	if err != nil {
		lh.Mysql.ExecError(err)
		return err
	}
	return nil
}

/*
Purpose : Fetches the current status of the sale
Input : saleid
Outputs : Object of StatusResponse and error if any
Remark : Return the  difference between the current time and the saleStartTime
*/
func GetStatus(saleId int) (*types.StatusResponse, error) {
	var funcName = "datastore/sale.go:GetStatus"
	log.WithFields(log.Fields{
		"saleId": saleId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select %s,%s from %s where Id = %d", c.SaleStartTime, c.StockUnits, c.SaleTable, saleId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	var status types.StatusResponse
	var saleStartTime int64
	err = stmt.QueryRow().Scan(&saleStartTime, &status.StockLeft)
	if err != nil {
		lh.Mysql.ScanError(err)
		return nil, err
	}
	now := time.Now()
	then := time.Unix(0, saleStartTime)
	status.TimeToStart = (then.Sub(now).Nanoseconds())

	return &status, nil

}
