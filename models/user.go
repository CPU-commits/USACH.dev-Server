package models

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserTypes string

const USERS_COLLECTION = "users"

// User Roles
const (
	USER  = "a"
	ADMIN = "b"
)

// Model
type User struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	FullName string             `json:"full_name" bson:"full_name"`
	Username string             `json:"username" bson:"username"`
	Email    string             `json:"email" bson:"email"`
	Profile  primitive.ObjectID `json:"profile,omitempty" bson:"profile,omitempty"`
	Password string             `json:"password,omitempty" bson:"password"`
	Status   bool               `bson:"status"`
	Role     string             `json:"role" bson:"role"`
	Date     primitive.DateTime `json:"date,omitempty" bson:"date,omitempty"`
}

type UsersModel struct{}

func (users *UsersModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(USERS_COLLECTION)
}

func (users *UsersModel) NewModel(
	fullName,
	email,
	password,
	role string,
) *User {
	username := strings.ReplaceAll(email, "@usach.cl", "")
	user := &User{
		FullName: fullName,
		Username: username,
		Email:    email,
		Password: password,
		Status:   false,
		Role:     role,
		Date:     primitive.NewDateTimeFromTime(time.Now()),
	}
	return user
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == USERS_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"email",
			"password",
			"date",
			"full_name",
			"username",
		},
		"properties": bson.M{
			"email": bson.M{
				"bsonType":  "string",
				"maxLength": 100,
			},
			"full_name": bson.M{
				"bsonType":  "string",
				"maxLength": 100,
			},
			"profile":  bson.M{"bsonType": "objectId"},
			"password": bson.M{"bsonType": "string"},
			"username": bson.M{"bsonType": "string"},
			"role":     bson.M{"enum": bson.A{"a", "b"}},
			"status":   bson.M{"bsonType": "bool"},
			"date":     bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(USERS_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func NewUsersModel() *UsersModel {
	return &UsersModel{}
}
