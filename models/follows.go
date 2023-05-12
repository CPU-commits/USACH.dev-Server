package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const FOLLOWS_COLLECTION = "follows"

// Model
type Follows struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id"`
	FollowedUser primitive.ObjectID `json:"followed_user" bson:"followed_user"`
	FollowerUser primitive.ObjectID `json:"follower_user" bson:"follower_user"`
	Date         primitive.DateTime `json:"date" bson:"date"`
}

type FollowsModel struct{}

func (follow *FollowsModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(FOLLOWS_COLLECTION)
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == FOLLOWS_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"followed_user",
			"follower_user",
			"date",
		},
		"properties": bson.M{
			"followed_user": bson.M{"bsonType": "objectId"},
			"follower_user": bson.M{"bsonType": "objectId"},
			"date":          bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(FOLLOWS_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}
