// All the database requests related to a post go here
// that include creating posts, post links, mascot queues, etc
package datastore

import (
	"encoding/hex"
	"errors"
	"fmt"
	c "rob/lib/common/constants"
	"rob/lib/common/types"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	lh "rob/lib/common/loghelper"

	log "github.com/sirupsen/logrus"
)

func GetTopPosts(numOfPosts int, mascotId int) ([]types.PostLink, error) {
	var funcName = "datastore/post.go:GetTopPosts"
	log.WithFields(log.Fields{
		"numOfPosts": numOfPosts,
		"mascotId":   mascotId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		SELECT %s, %s
		FROM %s
		WHERE %s = %d
		AND %s > %d
		ORDER BY %s 
		DESC LIMIT %d;`,
		c.PostId, c.TimeOfCreation,
		c.PostQueueTable,
		c.MascotId, mascotId,
		c.TimeOfCreation, c.DefaultTimestamp,
		c.TimeOfCreation,
		numOfPosts)

	lh.Mysql.Query(query)

	rows, err := db.Query(query)
	if err != nil {
		lh.Mysql.ExecError(err)
		return nil, err
	}
	defer rows.Close()
	log.Debugf("Returned PostQueue rows: %v", rows)

	var retItem []types.PostLink
	for rows.Next() {
		var cur types.PostLink

		err = rows.Scan(&cur.PostId, &cur.TimeOfCreation)
		log.Debugf("PostQueue row id: %s, timeOfCreation: %d", cur.PostId, cur.TimeOfCreation)
		if err != nil {
			log.WithField("ErrMsg", err.Error()).Error("Failed to scan one row of PostQueue. Continuing to next row")
			continue
		}

		retItem = append(retItem, cur)

	}

	return retItem, nil
}

func GetPostsAfter(timestamp int64, mascotId int, numOfPosts int) ([]types.PostLink, error) {
	var funcName = "datastore/post.go:GetPostsAfter"
	log.WithFields(log.Fields{
		"timestamp":  timestamp,
		"numOfPosts": numOfPosts,
		"mascotId":   mascotId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		SELECT %s, %s
		FROM  (SELECT * FROM %s 
		WHERE %s = %d
		AND %s > %d
		ORDER BY %s 
		LIMIT %d )
		AS T ORDER BY %s DESC;`,
		c.PostId, c.TimeOfCreation,
		c.PostQueueTable,
		c.MascotId, mascotId, c.TimeOfCreation, timestamp,
		c.TimeOfCreation, numOfPosts,
		c.TimeOfCreation,
	)

	lh.Mysql.Query(query)

	rows, err := db.Query(query)
	if err != nil {
		lh.Mysql.ExecError(err)
		return nil, err
	}
	defer rows.Close()
	log.Debugf("Returned PostQueue rows: %v", rows)

	var retItem []types.PostLink
	for rows.Next() {
		var cur types.PostLink

		err = rows.Scan(&cur.PostId, &cur.TimeOfCreation)
		log.Debugf("PostQueue row id: %s, timOfCreation: %d", cur.PostId, cur.TimeOfCreation)
		if err != nil {
			log.WithField("ErrMsg", err.Error()).Error("Failed to scan one row of PostQueue. Continuing to next row")
			continue
		}

		retItem = append(retItem, cur)

	}

	return retItem, nil
}

func GetPostsBefore(timestamp int64, mascotId int, numOfPosts int) ([]types.PostLink, error) {
	var funcName = "datastore/post.go:GetPostsBefore"
	log.WithFields(log.Fields{
		"timestamp":  timestamp,
		"numOfPosts": numOfPosts,
		"mascotId":   mascotId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	query := fmt.Sprintf(`
		SELECT %s, %s
		FROM (SELECT * FROM %s 
		WHERE %s = %d
		AND %s < %d
		ORDER BY %s 
		DESC LIMIT %d )
		AS T ORDER BY %s DESC;`,
		c.PostId, c.TimeOfCreation,
		c.PostQueueTable,
		c.MascotId, mascotId, c.TimeOfCreation, timestamp,
		c.TimeOfCreation, numOfPosts,
		c.TimeOfCreation,
	)

	lh.Mysql.Query(query)

	rows, err := db.Query(query)
	if err != nil {
		lh.Mysql.ExecError(err)
		return nil, err
	}
	defer rows.Close()
	log.Debugf("Returned PostQueue rows: %v", rows)

	var retItem []types.PostLink
	for rows.Next() {
		var cur types.PostLink

		err = rows.Scan(&cur.PostId, &cur.TimeOfCreation)
		log.Debugf("PostQueue row id: %s, timeOfCreation: %d", cur.PostId, cur.TimeOfCreation)
		if err != nil {
			log.WithField("ErrMsg", err.Error()).Error("Failed to scan one row of PostQueue. Continuing to next row")
			continue
		}

		retItem = append(retItem, cur)

	}

	return retItem, nil
}

func PostLink(mascotId int, postId string) error {
	var funcName = "datastore/post.go:PostLink"
	log.WithFields(log.Fields{
		"mascotId": mascotId,
		"postId":   postId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	timeOfCreation := time.Now().UTC().UnixNano()

	query := fmt.Sprintf(`
		INSERT INTO %s(%s, %s, %s)
		VALUES(%d,'%s',%d);`,
		c.PostQueueTable, c.TimeOfCreation, c.PostId, c.MascotId,
		timeOfCreation, postId, mascotId)
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

func GetPostLinks(mascotId int) ([]types.PostLink, error) {

	var funcName = "datastore/post.go:GetPostLinks"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var query string
	query = fmt.Sprintf("Select * from %s where %s=%d", c.PostQueueTable, c.MascotId, mascotId)
	lh.Mysql.Query(query)

	stmt, err := db.Prepare(query)
	if err != nil {
		lh.Mysql.PrepareError(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		lh.Mysql.ExecError(err)
		return nil, err
	}
	defer rows.Close()
	var postLinks []types.PostLink

	for rows.Next() {
		var postLink types.PostLink
		if err = rows.Scan(&postLink.TimeOfCreation, &postLink.PostId, &postLink.MascotId); err != nil {
			lh.Mysql.ScanError(err)
			continue
		}
		postLinks = append(postLinks, postLink)
	}
	if err = rows.Err(); err != nil {
		log.Error(err)
		return nil, err
	}
	return postLinks, nil

}

func AddPost(p types.Post) (string, error) {
	var funcName = "datastore/post.go:AddPost"
	log.WithFields(log.Fields{
		"cardType":   p.CardType,
		"dpSrc":      p.DpSrc,
		"title":      p.Title,
		"src":        p.Src,
		"desc":       p.Description,
		"buttonText": p.ButtonText,
		"url":        p.Url,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return "", err
	}
	defer session.Close()

	c := session.DB(c.DbName).C(c.Collection)

	p.Id = bson.NewObjectId()
	p.TimeOfCreation = time.Now().UTC().UnixNano()

	if err := c.Insert(&p); err != nil {
		lh.Mongo.WriteError(err)
		return "", err
	}

	log.Debugf("Post successfully written. postId: %s", p.Id.Hex())
	return p.Id.Hex(), nil
}

func GetPostMetaData(postId string) (*types.Post, error) {
	var funcName = "datastore/post.go:GetPostMetaData"
	log.WithFields(log.Fields{
		"postId": postId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var result types.Post

	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return nil, err
	}
	defer session.Close()

	// Calling bson.ObjectIdHex will panic if it's invalid.
	// To avoid that I am making validation checks here
	d, err := hex.DecodeString(postId)
	if err != nil || len(d) != 12 {
		return nil, errors.New("Invalid postId")
	}

	c := session.DB(c.DbName).C(c.Collection)
	if err := c.Find(bson.M{"_id": bson.ObjectIdHex(postId)}).One(&result); err != nil {
		lh.Mongo.ReadError(err)
		return nil, err
	}
	log.Debug("Post successfully retrieved")
	return &result, nil

}

func GetPosts() (*[]types.Post, error) {
	var funcName = "datastore/post.go:GetPosts"
	log.Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	var result []types.Post

	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return nil, err
	}
	defer session.Close()

	// Calling bson.ObjectIdHex will panic if it's invalid.
	// To avoid that I am making validation checks here

	c := session.DB(c.DbName).C(c.Collection)
	if err := c.Find(nil).All(&result); err != nil {
		lh.Mongo.ReadError(err)
		return nil, err
	}
	log.Debug("Posts successfully retrieved")
	return &result, nil

}

func DeletePost(postId string) error {
	var funcName = "datastore/post.go:DeletePost"
	log.WithFields(log.Fields{
		"postId": postId,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)

	d, err := hex.DecodeString(postId)
	if err != nil || len(d) != 12 {
		return errors.New("Invalid postId")
	}
	session, err := mgo.Dial(c.Server)
	if err != nil {
		lh.Mongo.ConnectError(err)
		return err
	}
	defer session.Close()

	c := session.DB(c.DbName).C(c.Collection)

	if err := c.Remove(bson.M{"_id": bson.ObjectIdHex(postId)}); err != nil {
		lh.Mongo.RemoveError(err)
		return err
	}
	return nil
}

func GetPostsMetaData(postLinks []types.PostLink) ([]types.Post, error) {
	var funcName = "datastore/post.go:GetPostsMetaData"
	log.WithFields(log.Fields{
		"postLinks": postLinks,
	}).Debugf("Enter: %s", funcName)
	defer log.Debugf("Exit: %s", funcName)
	allPostMetaData := []types.Post{}
	// For above reason, looping in reverse order will solve
	for _, postLink := range postLinks {
		log.Debugf("Looping over MascotPosts, current item: %v", postLink)
		postData, err := GetPostMetaData(postLink.PostId)
		if err != nil {
			log.WithField("ErrMsg", err.Error()).Error("Failed to get post data from DB. Continueing")
			continue
		}

		postData.TimeOfLink = postLink.TimeOfCreation
		allPostMetaData = append(allPostMetaData, *postData)
	}

	log.Debugf("Selected posts = %+v", allPostMetaData)
	return allPostMetaData, nil
}
