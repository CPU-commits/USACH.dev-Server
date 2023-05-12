package services

import (
	"github.com/CPU-commits/USACH.dev-Server/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User
type User struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	FullName string             `json:"full_name" bson:"full_name"`
	Username string             `json:"username" bson:"username"`
	Email    string             `json:"email" bson:"email"`
	Role     string             `json:"role" bson:"role"`
	Date     primitive.DateTime `json:"date,omitempty" bson:"date,omitempty"`
	Profile  *models.Profile    `json:"profile,omitempty"`
}

type SimpleUser struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id"`
	FullName string             `json:"full_name" bson:"full_name"`
	Username string             `json:"username" bson:"username"`
}

// Repository
type Repository struct {
	ID           primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	Owner        SimpleUser           `json:"owner" bson:"owner"`
	Name         string               `json:"name" bson:"name"`
	Content      string               `json:"content,omitempty" bson:"content,omitempty"`
	Stars        int                  `json:"stars"`
	SystemFile   []*models.SystemFile `json:"system_file,omitempty" bson:"system_file,omitempty"`
	Description  string               `json:"description,omitempty" bson:"description,omitempty"`
	Access       string               `json:"access" bson:"access"`
	Views        int                  `json:"views" bson:"views"`
	Links        []*models.Link       `json:"links,omitempty" bson:"links,omitempty"`
	Downloads    int                  `json:"downloads" bson:"downloads"`
	CustomAccess []primitive.ObjectID `json:"custom_access,omitempty" bson:"custom_access,omitempty"`
	UpdatedDate  primitive.DateTime   `json:"updated_date" bson:"updated_date"`
	CreatedDate  primitive.DateTime   `json:"created_date" bson:"created_date"`
	Tags         []string             `json:"tags" bson:"tags"`
}

// System File
type SystemFile struct {
	ID          primitive.ObjectID   `json:"_id" bson:"_id,omitempty"`
	FileType    string               `json:"file_type,omitempty" bson:"file_type,omitempty"`
	Name        string               `json:"name" bson:"name"`
	Childrens   []*models.SystemFile `json:"childrens,omitempty" bson:"childrens,omitempty"`
	Content     string               `json:"content,omitempty" bson:"content,omitempty"`
	IsDirectory bool                 `json:"is_directory" bson:"is_directory"`
	Date        primitive.DateTime   `json:"date" bson:"date"`
}
