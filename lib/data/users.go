package data

import (
	"rob/lib/common/types"
	"rob/lib/datastore"
)

func GetUserByEmail(email string) (*types.User, error) {
	return datastore.GetUserByEmail(email)
}
