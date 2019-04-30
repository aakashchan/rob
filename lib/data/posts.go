package data

import (
	"rob/lib/common/types"
	cache "rob/lib/datacache"
	"rob/lib/datastore"
)

func GetTopPosts(n, mascotId int) ([]types.PostLink, error) {
	return datastore.GetTopPosts(n, mascotId)
}

func GetPostsBefore(lastSync int64, mascotId, n int) ([]types.PostLink, error) {
	return datastore.GetPostsBefore(lastSync, mascotId, n)
}

func GetPostsAfter(lastSync int64, mascotId, n int) ([]types.PostLink, error) {
	return datastore.GetPostsAfter(lastSync, mascotId, n)
}

func GetPostsMetaData(links []types.PostLink) ([]types.Post, error) {
	return datastore.GetPostsMetaData(links)
}

func GetPostMetaData(postId string) (*types.Post, error) {
	// Try to return from cache
	v, err := cache.GetPostMetaData(postId)
	if err == nil {
		// cache hit
		return v, nil
	}

	// Try to get it from database
	v, err = datastore.GetPostMetaData(postId)
	if err != nil {
		return nil, err
	}

	// Add to cache
	go cache.AddPostMetaData(*v)
	return v, nil
}

func GetPosts(postIds []string) ([]types.Post, error) {
	var posts []types.Post

	for _, postId := range postIds {
		p, err := GetPostMetaData(postId)
		if err != nil {
			return nil, err
		}
		posts = append(posts, *p)
	}

	return posts, nil
}
