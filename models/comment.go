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

const COMMENTS_COLLECTION = "comments"

// Type
type Comment struct {
	ID         primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Author     primitive.ObjectID   `json:"author" bson:"author"`
	Discussion primitive.ObjectID   `json:"discussion" bson:"discussion"`
	Responses  []primitive.ObjectID `json:"responses,omitempty" bson:"responses,omitempty"`
	Comment    string               `json:"comment" bson:"comment"`
	IsRes      bool                 `json:"is_res" bson:"is_res"`
	CreatedAt  primitive.DateTime   `json:"created_at" bson:"created_at"`
	UpdatedAt  primitive.DateTime   `json:"updated_at" bson:"updated_at"`
}

// Responses
type CommentRes struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Author     SimpleUser         `json:"author" bson:"author"`
	Discussion primitive.ObjectID `json:"discussion" bson:"discussion"`
	Responses  []CommentRes       `json:"responses,omitempty" bson:"responses,omitempty"`
	Comment    string             `json:"comment" bson:"comment"`
	IsRes      bool               `json:"is_res" bson:"is_res"`
	CreatedAt  primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt  primitive.DateTime `json:"updated_at" bson:"updated_at"`
}

type CommentQueue struct {
	ID         string     `json:"_id"`
	Author     SimpleUser `json:"author"`
	Discussion string     `json:"discussion"`
	Comment    string     `json:"comment"`
	Date       string     `json:"date"`
	IsRes      bool       `json:"is_res"`
	ResTo      string     `json:"res_to,omitempty"`
}

func (c *Comment) ToQueue(idComment, resTo string) (*CommentQueue, error) {
	var author User

	cursor := userModel.Use().FindOne(db.Ctx, bson.D{{
		Key:   "_id",
		Value: c.Author,
	}})
	if err := cursor.Decode(&author); err != nil {
		return nil, err
	}

	return &CommentQueue{
		ID:         idComment,
		Discussion: c.Discussion.Hex(),
		Comment:    c.Comment,
		Date:       c.CreatedAt.Time().String(),
		Author: SimpleUser{
			ID:       author.ID,
			FullName: author.FullName,
			Username: author.Username,
		},
		IsRes: c.IsRes,
		ResTo: resTo,
	}, nil
}

// Struct
type CommentModel struct{}

func (*CommentModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(COMMENTS_COLLECTION)
}

func (cM *CommentModel) Exists(filter bson.D) (bool, error) {
	var comment Comment

	opts := options.FindOne().SetProjection(bson.D{{
		Key:   "_id",
		Value: 1,
	}})

	cursor := cM.Use().FindOne(db.Ctx, filter, opts)
	if err := cursor.Decode(&comment); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (*CommentModel) NewModel(
	comment string,
	idDiscussion,
	idUser primitive.ObjectID,
	isRes bool,
) *Comment {
	now := primitive.NewDateTimeFromTime(time.Now())

	return &Comment{
		Author:     idUser,
		Discussion: idDiscussion,
		Comment:    comment,
		CreatedAt:  now,
		UpdatedAt:  now,
		IsRes:      isRes,
	}
}

func (cm *CommentModel) Upload(comment *Comment) (primitive.ObjectID, error) {
	insertedId, err := cm.Use().InsertOne(db.Ctx, comment)
	if err == nil {
		return insertedId.InsertedID.(primitive.ObjectID), nil
	}
	return primitive.NilObjectID, err
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == COMMENTS_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"author",
			"comment",
			"created_at",
			"updated_at",
		},
		"properties": bson.M{
			"author":  bson.M{"bsonType": "objectId"},
			"comment": bson.M{"bsonType": "string", "maxLength": 325},
			"responses": bson.M{
				"bsonType": "array",
				"items": bson.M{
					"bsonType": "objectId",
				},
			},
			"created_at": bson.M{"bsonType": "date"},
			"updated_at": bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(COMMENTS_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func NewCommentModel() *CommentModel {
	return &CommentModel{}
}
