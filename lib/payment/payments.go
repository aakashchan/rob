package hash

import (
	"crypto/sha512"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	c "rob/lib/common/constants"
	"rob/lib/common/types"
	"rob/lib/data"
	"strconv"
)

/*
Purpose : Business logic for initiating a transaction for a order
Input : orderId for which transaction has to be initiated and email to get user details
Output : all fields required to initiate transaction with PayU , including a hash
Remark :
*/
func InitiateTransaction(orderId int, email string) (*types.HashResponse, error) {
	var funcName = "payment/payments.go:InitiateTransaction"
	log.WithFields(log.Fields{
		"orderId": orderId,
		"email":   email,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)
	var response types.HashResponse
	var err error
	err = nil
	order, err := data.GetOrder(orderId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, errors.New("OrderId not found")
		}
		return nil, err
	}
	response.Amount = order.Amount
	response.ProductInfo = order.ProductTitle
	user, err := data.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	response.FirstName = user.FirstName.String
	response.Email = email
	response.Key = c.PayUKey
	address, err := data.GetAddress(order.AddressId)
	if err != nil {
		return nil, err
	}
	phone, _ := strconv.Atoi(address.Phone)
	response.Phone = phone
	response.Surl = c.Surl
	response.Furl = c.Furl
	inStock, err := data.IsProductInStock(order.ProductId)
	if err != nil {
		return nil, err
	}
	if !inStock {
		return nil, errors.New("Product out of stock")
	}
	err = data.DecrementStock(order.ProductId)
	if err != nil {
		return nil, err
	}
	tran := data.CreateDefaultTransaction()
	tran.OrderId = orderId
	response.TxnId, err = data.InitiateTransaction(tran, order.SaleId)
	if err != nil {
		err1 := data.IncrementStock(order.ProductId)
		if err1 != nil {
			log.Error(err1.Error())
		}
		log.Error(err.Error())
		return nil, err
	}
	var hashString = response.Key + "|" + strconv.Itoa(response.TxnId) + "|" + strconv.Itoa(response.Amount) + "|" + response.ProductInfo + "|" + response.FirstName + "|" + response.Email + "|||||||||||" + c.PayUSalt
	s512 := sha512.New()
	s512.Write([]byte(hashString))
	response.Hash = fmt.Sprintf("%s", fmt.Sprintf("%x", s512.Sum(nil)))
	return &response, nil
}

/*
Purpose : Business logic for making updates and initiating shipping upon successful payment
Input : transaction object
Output : shippingId and error if any
Remark : shipping is not included in SQL transaction so that shipping failure does not discard the update to transaction


func OnSuccess(tran *types.Transaction) (int, error) {
	var funcName = "payment/payments.go:OnSuccess"
	log.WithFields(log.Fields{
		"transaction response from PayU": tran,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

}
*/
