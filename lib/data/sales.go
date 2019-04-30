package data

import (
	"rob/lib/datastore"
)

func UpdateSalesStock(saleId int, value int) error {
	return datastore.UpdateSaleStock(saleId, value)
}
