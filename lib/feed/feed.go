package feed

import (
	"encoding/json"
	c "rob/lib/common/constants"
	"rob/lib/common/types"
	"rob/lib/data"
)

func Get(lastSync int64, mascotId int, flag int) ([]types.Post, error) {
	var mascotFeedList []types.PostLink
	var err error
	err = nil
	if flag == c.TopPost {
		mascotFeedList, err = data.GetTopPosts(c.NumOfPosts, mascotId)
	} else if flag == c.PostBefore {
		mascotFeedList, err = data.GetPostsBefore(lastSync, mascotId, c.NumOfPosts)
	} else if flag == c.PostAfter {
		mascotFeedList, err = data.GetPostsAfter(lastSync, mascotId, c.NumOfPosts)
	}
	if err != nil {
		return nil, err
	}
	feed, err := data.GetPostsMetaData(mascotFeedList)
	if err != nil {
		return nil, err
	}

	feed, err = expand(feed)

	return feed, err
}

func expand(posts []types.Post) ([]types.Post, error) {
	// Expand List type cards
	for i, post := range posts {
		if post.CardType == c.CardTypeList {
			d, err := data.GetPosts(post.ChildPosts)
			if err != nil {
				return nil, err
			}
			for j := range d {
				d[j].TimeOfLink = post.TimeOfLink
			}
			m, err := json.Marshal(d)
			if err != nil {
				return nil, err
			}
			posts[i].ChildPostsJson = string(m)
		}
	}
	return posts, nil
}
