package models

import (
	"errors"
	"mime"
	"mime/multipart"
	"strings"
	"time"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const SYSTEM_FILE_COLLECTION = "system_files"

// Model
type SystemFile struct {
	ID          primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	FileType    string               `json:"file_type,omitempty" bson:"file_type,omitempty"`
	Name        string               `json:"name" bson:"name"`
	Childrens   []primitive.ObjectID `json:"childrens,omitempty" bson:"childrens,omitempty"`
	Content     string               `json:"content,omitempty" bson:"content,omitempty"`
	IsDirectory bool                 `json:"is_directory" bson:"is_directory"`
	Date        primitive.DateTime   `json:"date" bson:"date"`
}

// Responses
// System File
type SystemFileRes struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	FileType    string             `json:"file_type,omitempty" bson:"file_type,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Childrens   []*SystemFile      `json:"childrens,omitempty" bson:"childrens,omitempty"`
	Content     string             `json:"content,omitempty" bson:"content,omitempty"`
	IsDirectory bool               `json:"is_directory" bson:"is_directory"`
	Date        primitive.DateTime `json:"date" bson:"date"`
}

type SystemFileModel struct{}

func (repo *SystemFileModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(SYSTEM_FILE_COLLECTION)
}

func (repo *SystemFileModel) Exists(filter bson.D) (bool, error) {
	var systemFile *SystemFile

	options := options.FindOne().SetProjection(bson.D{{
		Key:   "_id",
		Value: 1,
	}})
	cursor := repo.Use().FindOne(db.Ctx, filter, options)

	if err := cursor.Decode(&systemFile); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (repo *SystemFileModel) NewModel(
	element *forms.SystemFileForm,
	file *multipart.FileHeader,
) (*SystemFile, error) {
	now := primitive.NewDateTimeFromTime(time.Now())
	if !*element.IsDirectory {
		ext := strings.Split(file.Filename, ".")
		uploadedFile, err := utils.UploadFile(file)
		if err != nil {
			return nil, err
		}

		return &SystemFile{
			Name:        file.Filename,
			IsDirectory: false,
			Date:        now,
			FileType:    mime.TypeByExtension(ext[len(ext)-1]),
			Content:     uploadedFile,
		}, nil
	}
	return &SystemFile{
		Name:        element.Name,
		IsDirectory: true,
		Date:        now,
	}, nil
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"name",
			"is_directory",
			"date",
		},
		"properties": bson.M{
			"file_type": bson.M{"bsonType": "string"},
			"name": bson.M{
				"bsonType":  "string",
				"maxLength": 100,
			},
			"stars": bson.M{"bsonType": "int"},
			"childrens": bson.M{
				"bsonType": "array",
				"items":    bson.M{"bsonType": "objectId"},
			},
			"content":      bson.M{"bsonType": "string"},
			"is_directory": bson.M{"bsonType": "bool"},
			"date":         bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	for _, collection := range collections {
		if collection == SYSTEM_FILE_COLLECTION {
			err := DbConnect.UpdateCollection(SYSTEM_FILE_COLLECTION, bson.D{{
				Key:   "validator",
				Value: validators,
			}})
			if err != nil {
				panic(err)
			}
			return
		}
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(SYSTEM_FILE_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func NewSystemFileModel() *SystemFileModel {
	return &SystemFileModel{}
}
