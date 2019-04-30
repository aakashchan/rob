// common database requests go here
package datastore

import (
	"database/sql"
	"fmt"
	c "rob/lib/common/constants"

	mgo "gopkg.in/mgo.v2"

	lh "rob/lib/common/loghelper"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

var db *sql.DB

func InitMySql() {
	var err error

	log.Info("Checking Mysql instance")

	dbb, err := sql.Open("mysql", c.DbUriBase)
	if err != nil {
		panic(err)
	}
	defer dbb.Close()

	var query string

	// Create database if needed
	query = fmt.Sprintf(`
		CREATE DATABASE IF NOT EXISTS %s`,
		c.DbName)

	if err := PrepareAndExec(query, dbb); err != nil {
		panic(err)
	}

	db, err = sql.Open("mysql", c.DbUri)
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}

	log.Info("Init of mysql done")
}

func CloseMySql() {
	if db != nil {
		db.Close()
	}
}

func InitDb() error {
	var funcName = "datastore/common.go:InitDb"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	log.Info("Checking Mongo instance")
	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return err
	}
	defer session.Close()
	log.Info("Mongo instance running OK")

	var query string
	// Create Users table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		(
			%s int(11) NOT NULL AUTO_INCREMENT,
			%s varchar(100),
			%s varchar(100),
			%s varchar(6),
			%s varchar(100),
			%s varchar(100),
			%s varchar(12) NOT NULL UNIQUE,
			%s bigint,
			%s int,
			%s varchar(400),
			%s varchar(400),
			%s varchar(400),
			PRIMARY KEY(%s)
		);`,
		c.UsersTable, c.Id, c.Email, c.Password, c.Gender, c.FirstName, c.LastName, c.Phone, c.TimeOfCreation, c.Verified, c.Code, c.Token, c.ResetPasswordToken, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Create Roles table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		(
			%s int(11) NOT NULL AUTO_INCREMENT,
			%s varchar(50) NOT NULL,
			PRIMARY KEY(%s)
		);`,
		c.RolesTable, c.Id, c.Name, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Add/update admin role to Roles table
	query = fmt.Sprintf(`
		INSERT INTO %s 
		VALUES(%d,'%s')
		ON DUPLICATE KEY
		UPDATE %s = %d;`,
		c.RolesTable, c.AdminRole, c.AdminRoleName, c.Id, c.AdminRole)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Add/update user role to Roles table
	query = fmt.Sprintf(`
		INSERT INTO %s 
		VALUES(%d,'%s')
		ON DUPLICATE KEY
		UPDATE %s = %d;`,
		c.RolesTable, c.UserRole, c.UserRoleName, c.Id, c.UserRole)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Add/update writer role to Roles table
	query = fmt.Sprintf(`
		INSERT INTO %s 
		VALUES(%d,'%s')
		ON DUPLICATE KEY
		UPDATE %s = %d;`,
		c.RolesTable, c.WriterRole, c.WriterRoleName, c.Id, c.WriterRole)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Create UserRole table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		(
			%s int NOT NULL AUTO_INCREMENT,
			%s int(11) NOT NULL,
			%s int(11) NOT NULL,
			FOREIGN KEY (%s) REFERENCES %s(%s),
			PRIMARY KEY (%s)
		);`,
		c.UserRoleTable, c.Id, c.UserId, c.RoleId, c.RoleId, c.RolesTable, c.Id, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Create Mascot table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		(
			%s int NOT NULL AUTO_INCREMENT,
			%s varchar(50),
			%s varchar(255),
			PRIMARY KEY(%s)
		);`,
		c.MascotTable, c.Id, c.Name, c.Description, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Add/Update default mascot details
	query = fmt.Sprintf(`
		INSERT INTO %s
		VALUES(%d,'%s','%s')
		ON DUPLICATE KEY
		UPDATE %s = %d;`,
		c.MascotTable, c.DefaultMascotId, c.DefaultMascotName,
		c.DefaultMascotDescription, c.Id, c.DefaultMascotId)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Create PostQueue table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		(
			%s bigint,
			%s varchar(255) NOT NULL,
			%s int NOT NULL,
			FOREIGN KEY (%s) REFERENCES %s(%s),
			PRIMARY KEY(%s,%s)
		);`,
		c.PostQueueTable, c.TimeOfCreation, c.PostId, c.MascotId,
		c.MascotId, c.MascotTable, c.Id, c.MascotId, c.PostId)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	//Create Sale table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS
		Sale(
			%s int NOT NULL AUTO_INCREMENT,
			%s varchar(500),
			%s varchar(200),
			%s varchar(50),
			%s varchar(500),
			%s varchar(500),
			%s int,
			%s bigint,
			%s bigint,
			%s bigint,
			PRIMARY KEY(%s)
		);`,
		c.Id, c.Title, c.Brand, c.ProductSku, c.Description, c.ThumbNail, c.StockUnits, c.SaleStartTime, c.SaleEndTime, c.TimeOfCreation, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	//Create Order table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS
		%s(
			%s int NOT NULL AUTO_INCREMENT,
			%s varchar(400),
			%s varchar(400),
			%s varchar(400),
			%s int(11),
			%s bigint,
			%s int(11),
			%s int(11),
			%s int(11),
			%s int(11),
			%s int(11),
			%s varchar(400),
			%s int(11),
			%s int(11),
			%s int(11),
			%s varchar(400),
			%s varchar(40),
			%s bigint,
			PRIMARY KEY(%s)
		);`, c.OrderTable, c.Id, c.ProductId, c.ProductTitle, c.ProductThumb, c.UserId, c.OrderDate, c.Price, c.Tax, c.ShippingCost, c.Amount, c.TransId, c.TransStatus, c.SaleId, c.AddressId, c.ShippingId, c.ShippingStatus, c.TrackingId, c.TimeOfCreation, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	//Create Address table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS
		%s(
			%s int NOT NULL AUTO_INCREMENT,
			%s int,
			%s varchar(1000),
			%s varchar(200),
			%s varchar(400),
			%s varchar(400),
			%s int(6),
			%s varchar(10),
			%s bigint,
			PRIMARY KEY(%s)
		);`, c.AddressTable, c.Id, c.UserId, c.Address, c.AddressType, c.City, c.State, c.PostalCode, c.Phone, c.TimeOfCreation, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	//Create Transaction table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS
		%s(
			%s int NOT NULL AUTO_INCREMENT,
			%s int,
			%s int,
			%s int,
			%s bigint,
			%s varchar(500),
			%s varchar(100),			
			%s varchar(100),
			%s varchar(100),
			%s varchar(100),
			%s varchar(100),
			%s varchar(500),
			PRIMARY KEY(%s)
		);`, c.TransactionTable, c.Id, c.Amount, c.OrderId, c.Phone, c.TimeOfCreation, c.ProductInfo, c.Email, c.PaymentMethod, c.PaymentId, c.PaymentStatus, c.FirstName, c.Hash, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	//Create Shipping table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS
		%s(
			%s int NOT NULL AUTO_INCREMENT,
			%s int,
			%s int,
			%s varchar(100),
			%s int,
			%s varchar(200),
			%s bigint,
			PRIMARY KEY(%s)
		);`, c.ShippingTable, c.Id, c.OrderId, c.UserId, c.TrackingId, c.AddressId, c.ShippingStatus, c.TimeOfCreation, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Create Feedback table if needed
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS
		%s(
			%s int NOT NULL AUTO_INCREMENT,
			%s varchar(100),
			%s varchar(100),
			%s varchar(2000),
			PRIMARY KEY(%s)
		);`, c.FeedbackTable, c.Id, c.Phone, c.Type, c.Description, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	// Create UrlCache table if needed
	query = fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS 
	%s(
		%s int NOT NULL AUTO_INCREMENT,
		%s varchar(100),
		PRIMARY KEY(%s)
	);`, c.UrlCacheTable, c.Id, c.Url, c.Id)

	if err := PrepareAndExec(query, db); err != nil {
		return err
	}

	log.Info("Mysql running OK")

	return nil
}

func AddFeedback(ph, typ, desc string) error {
	var funcName = "datastore/common.go:AddFeedback"
	log.WithFields(log.Fields{
		"type": typ,
		"desc": desc,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		INSERT INTO %s(%s, %s, %s)
		VALUES('%s', '%s', '%s')`,
		c.FeedbackTable,
		c.Phone, c.Type, c.Description,
		ph, typ, desc)

	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}

// Intended behaviour to not add common Enter/Exit statements to this func
func PrepareAndExec(query string, ldb *sql.DB) error {
	lh.Mysql.Query(query)
	stmt, err := ldb.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		lh.Mysql.ExecError(err)
		return err
	}
	return nil
}
