package data

import (
	"rob/lib/datastore"
)

func IsProductInStock(productId string) (bool, error) {
	return datastore.IsProductInStock(productId)
}

func DecrementStock(productId string) error {
	return datastore.DecrementStock(productId)
}

func IncrementStock(productId string) error {
	return datastore.IncrementStock(productId)
}
