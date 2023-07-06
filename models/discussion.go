package models

import (
	"errors"
	"mime/multipart"
	"time"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"github.com/thanhpk/randstr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DISCUSSION_COLLECTION = "discussions"

// Model
type Discussion struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Title      string             `json:"title" bson:"title"`
	Image      string             `json:"image,omitempty" bson:"image,omitempty"`
	Repository primitive.ObjectID `json:"repository,omitempty" bson:"repository,omitempty"`
	Text       string             `json:"text" bson:"text"`
	Tags       []string           `json:"tags,omitempty" bson:"tags,omitempty"`
	Owner      primitive.ObjectID `json:"owner" bson:"owner"`
	Code       string             `json:"code" bson:"code"`
	Snippet    string             `json:"snippet" bson:"snippet"`
	CreatedAt  primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt  primitive.DateTime `json:"updated_at" bson:"updated_at"`
}

// Responses
// Reactions
type DiscussionRes struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Title        string             `json:"title" bson:"title"`
	Image        string             `json:"image,omitempty" bson:"image,omitempty"`
	Repository   primitive.ObjectID `json:"repository,omitempty" bson:"repository,omitempty"`
	Text         string             `json:"text" bson:"text"`
	Tags         []string           `json:"tags,omitempty" bson:"tags,omitempty"`
	Owner        primitive.ObjectID `json:"owner" bson:"owner"`
	Code         string             `json:"code" bson:"code"`
	Snippet      string             `json:"snippet" bson:"snippet"`
	CreatedAt    primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt    primitive.DateTime `json:"updated_at" bson:"updated_at"`
	Reactions    []ReactionRes      `json:"reactions,omitempty" bson:"reactions,omitempty"`
	UserReaction Reaction           `json:"user_reaction,omitempty" bson:"user_reaction,omitempty"`
}

type DiscussionModel struct{}

func (*DiscussionModel) NewModel(
	discussion *forms.DiscussionForm,
	idUser primitive.ObjectID,
	image *multipart.FileHeader,
) (*Discussion, error) {
	now := primitive.NewDateTimeFromTime(time.Now())
	discussionModel := &Discussion{
		Title:     discussion.Title,
		Text:      discussion.Text,
		Tags:      discussion.Tags,
		Owner:     idUser,
		Code:      randstr.Hex(16),
		Snippet:   discussion.Snippet,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if discussion.Repository != "" {
		idObjRepository, _ := primitive.ObjectIDFromHex(discussion.Repository)
		discussionModel.Repository = idObjRepository
	}
	if image != nil {
		route, err := utils.UploadFile(image)
		if err != nil {
			return nil, err
		}
		discussionModel.Image = route
	}

	return discussionModel, nil
}

func (d *DiscussionModel) Exists(filter bson.D) (bool, error) {
	var discussion *Discussion

	options := options.FindOne().SetProjection(bson.D{{
		Key:   "_id",
		Value: 1,
	}})
	cursor := d.Use().FindOne(db.Ctx, filter, options)

	if err := cursor.Decode(&discussion); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (d *DiscussionModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(DISCUSSION_COLLECTION)
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == DISCUSSION_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"title",
			"text",
			"owner",
		},
		"properties": bson.M{
			"title":      bson.M{"bsonType": "string", "maxLength": 100},
			"owner":      bson.M{"bsonType": "objectId"},
			"image":      bson.M{"bsonType": "string"},
			"repository": bson.M{"bsonType": "objectId"},
			"text":       bson.M{"bsonType": "string"},
			"tags": bson.M{
				"bsonType": "array",
				"items": bson.M{
					"bsonType":  "string",
					"maxLength": 100,
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
	err := DbConnect.CreateCollection(DISCUSSION_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func NewDiscussionModel() *DiscussionModel {
	return &DiscussionModel{}
}
