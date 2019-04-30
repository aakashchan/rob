package data

import (
	c "rob/lib/common/constants"
	"rob/lib/common/types"
	"rob/lib/datastore"
)

func InitiateTransaction(trans types.Transaction, saleId int) (int, error) {
	return datastore.InitiateTransaction(trans, saleId)
}

func CreateDefaultTransaction() types.Transaction {
	var tran types.Transaction
	tran.Amount = c.DefaultInt
	tran.OrderId = c.DefaultInt
	tran.Phone = c.DefaultInt
	tran.ProductInfo = c.DefaultString
	tran.Email = c.DefaultString
	tran.PaymentMethod = c.DefaultString
	tran.PaymentId = c.DefaultString
	tran.PaymentStatus = c.DefaultString
	tran.FirstName = c.DefaultString
	tran.Hash = c.DefaultString
	return tran
}

func UpdateSuccessTransaction(tran types.Transaction) error {
	return datastore.UpdateSuccessTransaction(tran)
}
