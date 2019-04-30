package data

import (
	"rob/lib/datastore"

	"github.com/sirupsen/logrus"
)

var curls []string

func GetCacheDetails() ([]string, error) {

	var funcName = "data:cache.go/GetCacheDetails"
	logrus.Debugf("Enter :%s", funcName)
	defer logrus.Debugf("Exit %s", funcName)

	if curls == nil {
		urls, err := datastore.GetCacheDetails()

		if err != nil {
			return nil, err
		}

		curls = urls
		return curls, nil
	}

	return curls, nil
}

func DeleteCacheUrl(url string) error {
	var funcName = "data:cache.go/DeleteCacheUrl"
	logrus.Debugf("Enter %s", funcName)
	defer logrus.Debugf("Exit %s", funcName)

	curls = nil
	return datastore.DeleteCacheUrl(url)
}

func InsertCacheUrl(url string) error {
	var funcName = "data:cache.go/InsertCacheUrl"
	logrus.Debugf("Enter %s", funcName)
	defer logrus.Debugf("Exit %s", funcName)

	curls = nil
	return datastore.InsertCacheUrl(url)
}
