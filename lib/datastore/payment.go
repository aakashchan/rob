// All the database requests related to a Payments and Transactions go here
package datastore

import (
	"errors"
	"fmt"
	c "rob/lib/common/constants"
	lh "rob/lib/common/loghelper"
	"rob/lib/common/types"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	//	"rob/lib/datastore"
	"time"
)

/*
Purpose : Processes the payment through cc avenue
Input : transaction object
Outputs : Transation response object pointer and error if any
Remark : mocked version , will be changes upon ccavnue setup

func PaymentGateway(trans types.Transaction) (*types.TransactionResponse, error) {
	var funcName = "datastore/payment.go:PaymentGateway"
	log.WithFields(log.Fields{
		"Transaction": trans,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)
	var err error
	trans.PaymentMethod, trans.PaymentId, trans.PaymentStatus, trans.PaymentReturnCode, trans.PaymentMessage, err = ccavenue(c.ApiKey, c.ApiSecret, trans.Amount)
	if err != nil {
		return nil, errors.New("Error occured while processing with CC Avenue")
	}
	trans.Id, err = CreateTransaction(trans)
	if err != nil {
		return nil, errors.New("Failed to Create Transaction")
	}

	var res types.TransactionResponse
	res.TransId = trans.Id
	res.PaymentId = trans.PaymentId
	res.PaymentStatus = trans.PaymentStatus
	res.PaymentMessage = trans.PaymentMessage

	return &res, nil

}

/*
Purpose : Mock
Input : Mock
Outputs : Mock
Remark : Mock

//mock function for cc avenue . Return error only if request has timed out,etc
func ccavenue(key string, secret string, amount int) (string, string, string, string, string, error) {
	return "DEBITCARD", "319082381hsdjahsad", "Completed", "200OK", "Transaction successfully completed", nil
}

/*
Purpose : Creates a tansaction entry into the transaction table
Input : a Transaction object and saleId to decrement stock
Outputs : transactionId and error if any
Remark : Removed prepared statements as since they are likely to be reprepared multiple times on different connections when connections are busy.
*/
func InitiateTransaction(trans types.Transaction, saleId int) (int, error) {
	var funcName = "datastore/payment.go:InitiateTransaction"
	log.WithFields(log.Fields{
		"transaction": trans,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	tx, err := db.Begin()
	if err != nil {
		log.Error(err.Error())
	}
	var value = -1
	var query string
	query = fmt.Sprintf("UPDATE %s SET %s = %s + %d WHERE %s = %d ;", c.SaleTable, c.StockUnits, c.StockUnits, value, c.Id, saleId)
	lh.Mysql.Query(query)
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		log.Error(err.Error())
		return -1, err
	}
	trans.TimeOfCreation = time.Now().UTC().UnixNano()
	query = fmt.Sprintf("INSERT INTO %s(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s) VALUES(%d,%d,%d,%d,'%s','%s','%s','%s','%s','%s','%s');", c.TransactionTable, c.Amount, c.OrderId, c.Phone, c.TimeOfCreation, c.ProductInfo, c.Email, c.PaymentMethod, c.PaymentId, c.PaymentStatus, c.FirstName, c.Hash, trans.Amount, trans.OrderId, trans.Phone, trans.TimeOfCreation, trans.ProductInfo, trans.Email, trans.PaymentMethod, trans.PaymentId, trans.PaymentStatus, trans.FirstName, trans.Hash)
	lh.Mysql.Query(query)

	res, err := tx.Exec(query)
	if err != nil {
		tx.Rollback()
		log.Error(err.Error())
		return -1, err

	}
	tx.Commit()
	lastId, err := res.LastInsertId()
	if err != nil {
		return -1, errors.New("Coudnt retirve last inserted id ")
	}
	return int(lastId), nil

}

/*
Purpose : Update Successful Transaction information in the Order, Transaction table and initiate shipping
Input : Transactionid, transaction status , and the orderId
Outputs : error if any
Remark :
*/
func UpdateSuccessTransaction(trans types.Transaction) error {
	var funcName = "datastore/payment.go:UpdateOrderTransaction"
	log.WithFields(log.Fields{
		"transaction": trans,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	// Update Transaction Table
	tx, err := db.Begin()
	query := fmt.Sprintf("UPDATE %s SET %s = %d,%s=%d,%s=%d,%s = '%s',%s = '%s',%s = '%s',%s = '%s',%s = '%s',%s = '%s'%s = '%s' WHERE %s = %d", c.TransactionTable, c.Amount, trans.Amount, c.Phone, trans.Phone, c.TimeOfCreation, trans.TimeOfCreation, c.ProductInfo, trans.ProductInfo, c.Email, trans.Email, c.PaymentMethod, trans.PaymentMethod, c.PaymentId, trans.PaymentId, c.PaymentStatus, trans.PaymentStatus, c.FirstName, trans.FirstName, c.Hash, trans.Hash, c.Id, trans.Id)
	lh.Mysql.Query(query)
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		log.Error(err.Error())
		return err
	}

	// Update Order Table
	query = fmt.Sprintf("UPDATE %s SET %s=%d,%s=%s WHERE %s=%d", c.OrderTable, c.TransId, trans.Id, c.TransStatus, trans.PaymentStatus, c.Id, trans.OrderId)
	lh.Mysql.Query(query)
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		log.Error(err.Error())
		return err
	}
	tx.Commit()

	return nil
}

func GetTransaction(transId int) (*types.Transaction, error) {
	var funcName = "datastore/payment.go:GetTransaction"
	log.WithFields(log.Fields{
		"transactionId": transId,
	}).Debugf("Enter: %s", funcName)

	var query string
	query = fmt.Sprintf("Select * from %s where Id = %d", c.TransactionTable, transId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()
	var trans types.Transaction
	err = stmt.QueryRow().Scan(&trans.Id, &trans.Amount, &trans.OrderId, &trans.Phone, &trans.TimeOfCreation, &trans.ProductInfo, &trans.Email, &trans.PaymentMethod, &trans.PaymentId, &trans.PaymentStatus, &trans.FirstName, &trans.Hash)
	if err != nil {
		lh.Mysql.ScanError(err)
		return nil, err
	}
	return &trans, nil
}
