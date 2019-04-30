// All the database requests related to a Products go here
package datastore

import (
	"encoding/hex"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	c "rob/lib/common/constants"
	lh "rob/lib/common/loghelper"
	"rob/lib/common/types"
	"time"
)

/*
Purpose :  Adds a product listing in the mongo collection
Input : a Product object
Outputs : error if any
Remark :
*/

func AddProduct(newProduct types.Product) error {

	var funcName = "datastore/product.go:AddProduct"
	log.WithFields(log.Fields{
		"sku":         newProduct.Sku,
		"title":       newProduct.Title,
		"brand":       newProduct.Brand,
		"quantity":    newProduct.Quantity,
		"description": newProduct.Description,
		"unitprice":   newProduct.UnitPrice,
		"summary":     newProduct.Summary,
		"image":       newProduct.Image,
		"thumb":       newProduct.ThumbNail,
		"color":       newProduct.Color,
		"size":        newProduct.Size,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return err
	}
	defer session.Close()
	newProduct.TimeOfCreation = time.Now().UTC().UnixNano()

	c := session.DB(c.DbName).C(c.ProductCollection)

	if err := c.Insert(&newProduct); err != nil {
		lh.Mongo.WriteError(err)
		return err
	}

	log.Debugf("Product successfully Listed. productId: %s", newProduct.Id.Hex())
	return nil

}

/*
Purpose : Retrives product details from mongo
Input : ProductId
Outputs : a product object pointer
Remark :
*/

func GetProduct(productId string) (*types.Product, error) {

	var funcName = "datastore/product.go:GetProduct"
	log.WithFields(log.Fields{
		"productId": productId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var result types.Product

	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return nil, err
	}
	defer session.Close()

	// Calling bson.ObjectIdHex will panic if it's invalid.
	// To avoid that I am making validation checks here
	d, err := hex.DecodeString(productId)
	if err != nil || len(d) != 12 {
		return nil, errors.New("Invalid productId")
	}

	c := session.DB(c.DbName).C(c.ProductCollection)
	if err := c.Find(bson.M{"_id": bson.ObjectIdHex(productId)}).One(&result); err != nil {
		lh.Mongo.ReadError(err)
		return nil, err
	}
	log.Debug("Product successfully retrieved")
	return &result, nil

}

func IsProductInStock(productId string) (bool, error) {
	var funcName = "datastore/common.go:IsProductInStock"
	log.WithFields(log.Fields{
		"productId": productId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return false, err
	}
	defer session.Close()

	c := session.DB(c.DbName).C(c.ProductCollection)
	var result types.Product
	if err := c.Find(bson.M{"_id": bson.ObjectIdHex(productId)}).One(&result); err != nil {
		lh.Mongo.ReadError(err)
		return false, err
	}

	if result.Quantity > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func IncrementStock(productId string) error {
	var funcName = "datastore/common.go:IncrementStock"
	log.WithFields(log.Fields{
		"productId": productId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return err
	}
	defer session.Close()
	c := session.DB(c.DbName).C(c.ProductCollection)
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"quantity": 1}},
		ReturnNew: true,
	}
	var doc types.Product
	_, err = c.Find(bson.M{"_id": bson.ObjectIdHex(productId)}).Apply(change, &doc)
	if err != nil {
		lh.Mongo.UpdateError(err)
		return err
	}
	return nil

}

func DecrementStock(productId string) error {
	var funcName = "datastore/common.go:DecrementStock"
	log.WithFields(log.Fields{
		"productId": productId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)
	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return err
	}
	defer session.Close()

	c := session.DB(c.DbName).C(c.ProductCollection)
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"quantity": -1}},
		ReturnNew: true,
	}
	var result types.Product
	if err := c.Find(bson.M{"_id": bson.ObjectIdHex(productId)}).One(&result); err != nil {
		lh.Mongo.ReadError(err)
		return err
	}

	if result.Quantity < 1 {
		return errors.New("Cannot Decrement Out of stock product")
	}

	var doc types.Product
	_, err = c.Find(bson.M{"_id": bson.ObjectIdHex(productId)}).Apply(change, &doc)
	if err != nil {
		lh.Mongo.UpdateError(err)
		return err
	}
	return nil

}
