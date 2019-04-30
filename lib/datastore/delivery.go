// All the database requests related to a Delivery go here
package datastore

import (
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"regexp"
)

/*
Purpose : To check whether a pincode is deliverable by Twiq or not
Input : postal code
Outputs : True or false
Remark :
*/

func CheckDelivery(pc string) bool {
	var funcName = "datastore/delivery.go:GetAddress"
	log.WithFields(log.Fields{
		"postalcode": pc,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	match, _ := regexp.MatchString("560[0-9][0-9][0-9]", pc)
	if match {
		return true
	} else {
		return false
	}
}
