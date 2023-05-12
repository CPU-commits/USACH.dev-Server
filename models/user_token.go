package models

import (
	"encoding/hex"
	"math/rand"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const USERS_TOKEN_COLLECTION = "users_token"

// Permissions
const (
	PERMISSION_CONFIRM_USER = "confirm_user"
)

// Model
type UserToken struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	User        primitive.ObjectID `json:"user" bson:"user"`
	FinishDate  primitive.DateTime `json:"finish_date" bson:"finish_date"`
	Token       string             `json:"token" bson:"token"`
	Permissions []string           `json:"permissions" bson:"permissions"`
}

type UsersTokenModel struct{}

func (users *UsersTokenModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(USERS_TOKEN_COLLECTION)
}

func (users *UsersTokenModel) NewModel(
	user primitive.ObjectID,
	finishDate primitive.DateTime,
	permissions []string,
) (*UserToken, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	userModel := &UserToken{
		User:        user,
		FinishDate:  finishDate,
		Token:       hex.EncodeToString(b),
		Permissions: permissions,
	}
	return userModel, nil
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == USERS_TOKEN_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"user",
			"finish_date",
			"token",
			"permissions",
		},
		"properties": bson.M{
			"user":        bson.M{"bsonType": "objectId"},
			"finish_date": bson.M{"bsonType": "date"},
			"token":       bson.M{"bsonType": "string"},
			"permissions": bson.M{
				"bsonType": bson.A{"array"},
				"items": bson.M{
					"bsonType": "string",
				},
			},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(USERS_TOKEN_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func NewUsersTokenModel() *UsersTokenModel {
	return &UsersTokenModel{}
}
