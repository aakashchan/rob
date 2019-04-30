// All the database requests related to a Address go here
package datastore

import (
	"errors"
	"fmt"
	c "rob/lib/common/constants"
	lh "rob/lib/common/loghelper"
	"rob/lib/common/types"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

/*
Purpose : Adds an address entry into the address table
Input : an Address object
Outputs : addressId and error if any
Remark :
*/

func AddAddress(address types.Address) (int, error) {
	var funcName = "datastore/address.go:AddAddress"
	log.WithFields(log.Fields{
		"Address": address,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	timeofCreation := time.Now().UTC().UnixNano()
	query = fmt.Sprintf("INSERT INTO %s(%s,%s,%s,%s,%s,%s,%s,%s) VALUES(%d,'%s','%s','%s','%s',%d,'%s',%d);", c.AddressTable, c.UserId, c.Address, c.AddressType, c.City, c.State, c.PostalCode, c.Phone, c.TimeOfCreation, address.UserId, address.Address, address.AddressType, address.City, address.State, address.PostalCode, address.Phone, timeofCreation)
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
Purpose : Retrieves an address entry from the address table
Input : addressId
Outputs : an address object pointer and error if any
Remark :
*/

func GetAddress(aId int) (*types.Address, error) {
	var funcName = "datastore/address.go:GetAddress"
	log.WithFields(log.Fields{
		"addressId": aId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select %s,%s,%s,%s,%s,%s,%s,%s from %s where Id = %d", c.Id, c.UserId, c.Address, c.AddressType, c.City, c.State, c.PostalCode, c.Phone, c.AddressTable, aId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	var address types.Address
	err = stmt.QueryRow().Scan(&address.Id, &address.UserId, &address.Address, &address.AddressType, &address.City, &address.State, &address.PostalCode, &address.Phone)

	if err != nil {
		lh.Mysql.ScanError(err)
		return nil, err
	}

	return &address, nil

}

/*
Purpose : Retrive all address added by a User
Input : UserId
Outputs : AddressList object pointer
Remark : Wraps all address in a addressList
*/

func GetUserAddresses(userId int) (*types.AddressList, error) {

	var funcName = "datastore/address.go:GetUserAdresses"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select * from %s where %s=%d", c.AddressTable, c.UserId, userId)
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
	var addresses types.AddressList

	for rows.Next() {
		var address types.Address
		if err = rows.Scan(&address.Id, &address.UserId, &address.Address, &address.AddressType, &address.City, &address.State, &address.PostalCode, &address.Phone, &address.TimeOfCreation); err != nil {
			lh.Mysql.ScanError(err)
			continue
		}
		addresses.Data = append(addresses.Data, address)
	}
	if err = rows.Err(); err != nil {
		log.Error(err)
		return nil, err
	}
	return &addresses, nil

}

/*
Purpose : Edits attributes of a address entry
Input : addressId for address to be updated and the new address object
Outputs : error if any
Remark : Frontend should auto fill the address to be edited else fields can be replaced by null
*/

func EditAddress(addressId int, newAddress types.Address) error {
	var funcName = "datastore/address.go:EditAddress"
	log.WithFields(log.Fields{
		"Address": newAddress,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	timeofCreation := time.Now().UTC().UnixNano()
	query = fmt.Sprintf("UPDATE %s SET %s='%s',%s='%s',%s='%s',%s='%s',%s=%d,%s='%s',%s=%d WHERE %s = %d;", c.AddressTable, c.Address, newAddress.Address, c.AddressType, newAddress.AddressType, c.City, newAddress.City, c.State, newAddress.State, c.PostalCode, newAddress.PostalCode, c.Phone, newAddress.Phone, c.TimeOfCreation, timeofCreation, c.Id, addressId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		lh.Mysql.ExecError(err)
		return err
	}
	return nil

}
