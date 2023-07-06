package models

import (
	"errors"
	"time"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/forms"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const REPOSITORY_COLLECTION = "repositories"

// Model
// Link
type Link struct {
	ID    primitive.ObjectID `json:"_id" bson:"_id"`
	Type  string             `json:"type" bson:"type"`
	Title string             `json:"title" bson:"title"`
	Link  string             `json:"link" bson:"link"`
}

// Repository
type Repository struct {
	ID           primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Owner        primitive.ObjectID   `json:"owner" bson:"owner"`
	Name         string               `json:"name" bson:"name"`
	Stars        int                  `json:"stars"`
	Tags         []string             `json:"tags,omitempty" bson:"tags,omitempty"`
	Content      string               `json:"content" bson:"content"`
	SystemFile   []primitive.ObjectID `json:"system_file,omitempty" bson:"system_file,omitempty"`
	Description  string               `json:"description,omitempty" bson:"description,omitempty"`
	Access       string               `json:"access" bson:"access"`
	Views        int                  `json:"views" bson:"views"`
	Links        []Link               `json:"links" bson:"links,omitempty"`
	Downloads    int                  `json:"downloads" bson:"downloads"`
	CustomAccess []primitive.ObjectID `json:"custom_access,omitempty" bson:"custom_access,omitempty"`
	UpdatedDate  primitive.DateTime   `json:"updated_date" bson:"updated_date"`
	CreatedDate  primitive.DateTime   `json:"created_date" bson:"created_date"`
}

// Responses
type RepositoryRes struct {
	ID           primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Owner        SimpleUser           `json:"owner" bson:"owner"`
	Name         string               `json:"name" bson:"name"`
	Content      string               `json:"content,omitempty" bson:"content,omitempty"`
	Stars        int                  `json:"stars"`
	SystemFile   []*SystemFile        `json:"system_file,omitempty" bson:"system_file,omitempty"`
	Description  string               `json:"description,omitempty" bson:"description,omitempty"`
	Access       string               `json:"access" bson:"access"`
	Views        int                  `json:"views" bson:"views"`
	Links        []*Link              `json:"links,omitempty" bson:"links,omitempty"`
	Downloads    int                  `json:"downloads" bson:"downloads"`
	CustomAccess []primitive.ObjectID `json:"custom_access,omitempty" bson:"custom_access,omitempty"`
	UpdatedDate  primitive.DateTime   `json:"updated_date" bson:"updated_date"`
	CreatedDate  primitive.DateTime   `json:"created_date" bson:"created_date"`
	Tags         []string             `json:"tags" bson:"tags"`
}

type RepositoryModel struct{}

func (repo *RepositoryModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(REPOSITORY_COLLECTION)
}

func (repo *RepositoryModel) Exists(filter bson.D) (bool, error) {
	var repository *Repository

	options := options.FindOne().SetProjection(bson.D{{
		Key:   "_id",
		Value: 1,
	}})
	cursor := repo.Use().FindOne(db.Ctx, filter, options)

	if err := cursor.Decode(&repository); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *RepositoryModel) NewModel(
	repo *forms.RepositoryForm,
	owner primitive.ObjectID,
) *Repository {
	now := primitive.NewDateTimeFromTime(time.Now())
	return &Repository{
		Owner:       owner,
		Name:        repo.Name,
		Stars:       0,
		Description: repo.Description,
		Access:      repo.Access,
		Views:       0,
		Downloads:   0,
		UpdatedDate: now,
		CreatedDate: now,
	}
}

func (r *RepositoryModel) NewLinkModel(link *forms.LinkForm) *Link {
	return &Link{
		Type:  link.Type,
		Title: link.Title,
		Link:  link.Link,
		ID:    primitive.NewObjectID(),
	}
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"owner",
			"name",
			"stars",
			"access",
			"views",
			"downloads",
			"updated_date",
			"created_date",
		},
		"properties": bson.M{
			"owner": bson.M{"bsonType": "objectId"},
			"name": bson.M{
				"bsonType":  "string",
				"maxLength": 100,
			},
			"stars": bson.M{"bsonType": "int"},
			"system_file": bson.M{
				"bsonType": "array",
				"items":    bson.M{"bsonType": "objectId"},
			},
			"description": bson.M{
				"bsonType":  "string",
				"maxLength": 300,
			},
			"content": bson.M{"bsonType": "string"},
			"access": bson.M{
				"bsonType": "string",
				"enum":     bson.A{"private", "private-group", "public"},
			},
			"tags": bson.M{
				"bsonType": "array",
				"items": bson.M{
					"bsonType": "string",
				},
			},
			"links": bson.M{
				"bsonType": "array",
				"items": bson.M{
					"bsonType": "object",
					"required": bson.A{
						"_id",
						"title",
						"type",
						"link",
					},
					"properties": bson.M{
						"_id": bson.M{"bsonType": "objectId"},
						"title": bson.M{
							"bsonType":  "string",
							"maxLength": 30,
						},
						"type": bson.M{
							"bsonType": "string",
							"enum": bson.A{
								"drive",
								"github",
								"cloud",
								"pdf",
								"youtube",
								"other",
							},
						},
						"link": bson.M{
							"bsonType": "string",
						},
					},
				},
			},
			"views":        bson.M{"bsonType": "int"},
			"downloads":    bson.M{"bsonType": "int"},
			"updated_date": bson.M{"bsonType": "date"},
			"custom_access": bson.M{
				"bsonType": "array",
				"items": bson.M{
					"bsonType": "objectId",
				},
			},
			"created_date": bson.M{"bsonType": "date"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	for _, collection := range collections {
		if collection == REPOSITORY_COLLECTION {
			err := DbConnect.UpdateCollection(REPOSITORY_COLLECTION, bson.D{{
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
	err := DbConnect.CreateCollection(REPOSITORY_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}

func NewRepositoryModel() *RepositoryModel {
	return &RepositoryModel{}
}
