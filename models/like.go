package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const LIKES_COLLECTION = "likes"

// Model
type Like struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	User       primitive.ObjectID `json:"user" bson:"user"`
	Repository primitive.ObjectID `json:"repository" bson:"repository"`
	Plus       bool               `json:"plus" bson:"plus"`
	Date       primitive.DateTime `json:"date" bson:"date"`
}

type LikeModel struct{}

func (like *LikeModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(LIKES_COLLECTION)
}

func (l *LikeModel) NewLikeModel(idUser, idRepository primitive.ObjectID, plus bool) *Like {
	return &Like{
		User:       idUser,
		Repository: idRepository,
		Plus:       plus,
		Date:       primitive.NewDateTimeFromTime(time.Now()),
	}
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == LIKES_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"user",
			"repository",
			"plus",
			"date",
		},
		"properties": bson.M{
			"user":       bson.M{"bsonType": "objectId"},
			"repository": bson.M{"bsonType": "objectId"},
			"plus":       bson.M{"bsonType": "bool"},
			"date":       bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(LIKES_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func NewLikesModel() *LikeModel {
	return &LikeModel{}
}
