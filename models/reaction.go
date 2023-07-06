package models

import (
	"errors"
	"time"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const REACTION_COLLECTION = "reactions"

// Model
type Reaction struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	User       primitive.ObjectID `json:"user" bson:"user"`
	Discussion primitive.ObjectID `json:"discussion" bson:"discussion"`
	Reaction   string             `json:"reaction" bson:"reaction"`
	Date       primitive.DateTime `json:"date" bson:"date"`
}

// Responses
type ReactionRes struct {
	Reaction string `json:"reaction" bson:"reaction"`
	Count    int    `json:"count" bson:"count"`
}

type ReactionModel struct{}

func (reaction *ReactionModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(REACTION_COLLECTION)
}

func (r *ReactionModel) Exists(filter bson.D) (bool, error) {
	options := options.FindOne().SetProjection(bson.D{{
		Key:   "_id",
		Value: true,
	}})

	// Find
	var reaction *Reaction
	cursor := r.Use().FindOne(db.Ctx, filter, options)
	if err := cursor.Decode(&reaction); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *ReactionModel) NewReactionModel(idUser, idDiss primitive.ObjectID, reaction string) *Reaction {
	return &Reaction{
		User:       idUser,
		Discussion: idDiss,
		Reaction:   reaction,
		Date:       primitive.NewDateTimeFromTime(time.Now()),
	}
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == REACTION_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"user",
			"discussion",
			"reaction",
			"date",
		},
		"properties": bson.M{
			"user":       bson.M{"bsonType": "objectId"},
			"discussion": bson.M{"bsonType": "objectId"},
			"reaction":   bson.M{"bsonType": "string"},
			"date":       bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(REACTION_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func NewReactionModel() *ReactionModel {
	return &ReactionModel{}
}
