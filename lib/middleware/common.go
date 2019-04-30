package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"regexp"
	c "rob/lib/common/constants"
	"rob/lib/common/httperr"
	lh "rob/lib/common/loghelper"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var Re = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
var PhRe = regexp.MustCompile("^[789]\\d{9}$")

func Common(h http.Handler) http.Handler {
	h = handlers.LoggingHandler(os.Stdout, h)

	var enableCSRF = false

	CSRF := csrf.Protect([]byte("AJSD@*JD82!#$S@@DS(*HJDnasd2*@#"), csrf.Secure(false))

	if enableCSRF {
		h = CSRF(h)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		h.ServeHTTP(w, r)
	})

}

func ValidateEmail(f http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var funcName = "middleware/common.go:ValidateEmail"
		log.Debugf("Enter: %s", funcName)
		defer log.Debugf("Exit: %s", funcName)

		email := r.FormValue(c.Email)

		if !validEmail(email) {
			httperr.E(w, http.StatusBadRequest, fmt.Sprintf("Email %q is not valid", email), nil)
			return
		}

		f.ServeHTTP(w, r)
	})
}

func validEmail(em string) bool {
	em = strings.ToLower(em)
	return Re.MatchString(em)
}

func ValidatePhone(f http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var funcName = "middleware/common.go:ValidatePhone"
		log.Debugf("Enter: %s", funcName)
		defer log.Debugf("Exit: %s", funcName)

		phone := r.FormValue(c.Phone)
		if !validPhoneNumber(phone) {
			httperr.E(w, http.StatusBadRequest, fmt.Sprintf("Phone Number %q is not valid", phone), nil)
			return
		}

		f.ServeHTTP(w, r)
	})
}

func validPhoneNumber(p string) bool {
	return PhRe.MatchString(p)
}

func CheckExistanceMysql(tableName string, fieldName string, fieldValue string, isInt bool) (bool, error) {
	var funcName = "middleware/common.go:checkExistanceMysql"
	log.WithFields(log.Fields{
		"tableName":  tableName,
		"fieldName":  fieldName,
		"fieldValue": fieldValue,
		"isInt":      isInt,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	db, err := sql.Open("mysql", c.DbUri)
	if err != nil {
		lh.Mysql.ConnectError(err)
		return false, err
	}
	defer db.Close()
	var exists bool
	queryString := fmt.Sprintf("SELECT exists (SELECT * FROM %s WHERE %s='%s')", tableName, fieldName, fieldValue)
	queryInt := fmt.Sprintf("SELECT exists (SELECT * FROM %s WHERE %s=%s)", tableName, fieldName, fieldValue)

	query := queryString
	if isInt {
		query = queryInt
	}
	err = db.QueryRow(query).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Error("error checking if row exists", err)
		return false, err
	}
	return exists, nil
}

func CheckExistanceMongo(collection string, id string) (bool, error) {
	var funcName = "middleware/common.go:checkExistanceMongo"
	log.WithFields(log.Fields{
		"collection": collection,
		"id":         id,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return false, err
	}
	defer session.Close()

	c := session.DB(c.DbName).C(collection)
	count, err := c.Find(bson.M{"_id": bson.ObjectIdHex(id)}).Count()
	if err != nil {
		lh.Mongo.ReadError(err)
		return false, err
	}

	log.Debugf("Product exists with Productid : %s", id)
	return (count > 0), nil

}
