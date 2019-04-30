package cache

import (
	"errors"
	"rob/lib/common/types"

	"github.com/golang/groupcache/lru"
	log "github.com/sirupsen/logrus"
)

var postsCache = lru.New(4096)

var ErrCacheMiss = errors.New("Not found in cache")
var ErrCacheWrongType = errors.New("Type conversion error")

func GetPostMetaData(postId string) (*types.Post, error) {
	var funcName = "lib/datacache/posts.go:GetPostMetaData"
	log.WithFields(log.Fields{
		"postId": postId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	value, hit := postsCache.Get(postId)

	if hit == false {
		return nil, ErrCacheMiss
	}

	if v, ok := value.(*types.Post); !ok {
		return nil, ErrCacheWrongType
	} else {
		return v, nil
	}
}

func AddPostMetaData(post types.Post) {
	postsCache.Add(post.Id.Hex(), post)
}
