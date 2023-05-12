package models

import (
	"mime/multipart"

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
}

type DiscussionModel struct{}

func (d *DiscussionModel) NewModel(
	discussion *forms.DiscussionForm,
	idUser primitive.ObjectID,
	image *multipart.FileHeader,
) (*Discussion, error) {
	discussionModel := &Discussion{
		Title: discussion.Title,
		Text:  discussion.Text,
		Tags:  discussion.Tags,
		Owner: idUser,
		Code:  randstr.Hex(16),
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
