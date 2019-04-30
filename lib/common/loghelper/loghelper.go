package loghelper

import (
	log "github.com/sirupsen/logrus"
)

// These structs are just to get nice syntax like lh.Mysql.ConnectError
// Instead of lh.MysqlConnectError

type mysqlLogs struct {
}

var Mysql mysqlLogs

type mongoLogs struct {
}

var Mongo mongoLogs

func (m mysqlLogs) ConnectError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed to connect mysql DB")
}

func (m mysqlLogs) PrepareError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed to prepare mysql query")
}

func (m mysqlLogs) ScanError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed to scan db response")
}

func (m mysqlLogs) ExecError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed to execute mysql stmt")
}

func (m mysqlLogs) Query(s string) {
	log.Debugf("Database query: %s", s)
}

func (m mongoLogs) ConnectError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed to connect to Mongo Server")
}

func (m mongoLogs) ReadError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed reading from Mongo")
}

func (m mongoLogs) RemoveError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed deleting from Mongo")
}

func (m mongoLogs) UpdateError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed updating the Mongo document")
}

func (m mongoLogs) WriteError(err error) {
	log.WithField("ErrMsg", err.Error()).Error("Failed writing to Mongo")
}
