package data

import (
	"rob/lib/common/types"
	"rob/lib/datastore"
)

func GetAddress(addressId int) (*types.Address, error) {
	return datastore.GetAddress(addressId)
}
