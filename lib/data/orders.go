package data

import (
	"rob/lib/common/types"
	"rob/lib/datastore"
)

func GetOrder(orderId int) (*types.Order, error) {
	return datastore.GetOrder(orderId)
}
