package cache

import (
	"rob/lib/common/types"
	"rob/lib/datastore"

	"github.com/golang/groupcache/lru"
	log "github.com/sirupsen/logrus"
)

var cacheData = lru.New(1000)

func GetMetaData(postId string) (*types.Post, error) {
	var funcName = "lib/datacache/cache.go:GetMetaData"
	log.WithFields(log.Fields{
		"postId": postId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	value, hit := cacheData.Get(postId)

	if hit == false {
		post, err := datastore.GetPostMetaData(postId)

		if err != nil {
			log.WithField("ErrMsg", err.Error()).Error("Failed to get post data from db")
			return nil, err
		}
		go cacheData.Add(postId, post)
		return post, nil
	}

	v := value.(*types.Post)
	log.Debugf("Fetched metadata for postId=%s, value=%+v", postId, v)
	return v, nil
}
