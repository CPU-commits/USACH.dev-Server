package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PROFILE_COLLECTION = "profiles"

// Model
// Media
type Media struct {
	Instagram string `json:"instagram,omitempty" bson:"instagram,omitempty"`
	Github    string `json:"github,omitempty" bson:"github,omitempty"`
	Twitter   string `json:"twitter,omitempty" bson:"twitter,omitempty"`
}

type Profile struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	User        primitive.ObjectID `json:"user" bson:"user"`
	Media       *Media             `json:"media,omitempty" bson:"media,omitempty"`
	Description string             `json:"description" bson:"description"`
	Date        primitive.DateTime `json:"date" bson:"date"`
}

type ProfileModel struct{}

func (profile *ProfileModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(PROFILE_COLLECTION)
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == PROFILE_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"user",
			"date",
		},
		"properties": bson.M{
			"user": bson.M{"bsonType": "objectId"},
			"description": bson.M{
				"bsonType":  "string",
				"maxLength": 300,
			},
			"media": bson.M{
				"bsonType": "object",
				"properties": bson.M{
					"instagram": bson.M{
						"bsonType":  "string",
						"maxLength": 100,
					},
					"github": bson.M{
						"bsonType":  "string",
						"maxLength": 100,
					},
					"twitter": bson.M{
						"bsonType":  "string",
						"maxLength": 100,
					},
				},
			},
			"date": bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(PROFILE_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}
