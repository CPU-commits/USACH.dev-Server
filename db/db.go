package db

import (
	"context"
	"fmt"

	"github.com/CPU-commits/USACH.dev-Server/settings"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var settingsData = settings.GetSettings()
var Ctx = context.TODO()

type MongoClient struct {
	client   *mongo.Client
	database string
}

func newMongoClient(client *mongo.Client, database string) *MongoClient {
	return &MongoClient{
		client:   client,
		database: database,
	}
}

func (mongo *MongoClient) GetCollection(collectionName string) *mongo.Collection {
	collection := mongo.client.Database(mongo.database).Collection(collectionName)
	return collection
}

func (mongo *MongoClient) GetCollections() ([]string, error) {
	filter := bson.D{}
	return mongo.client.Database(mongo.database).ListCollectionNames(Ctx, filter)
}

func (mongo *MongoClient) CreateCollection(collectionName string, opts *options.CreateCollectionOptions) error {
	db := mongo.client.Database(mongo.database)
	return db.CreateCollection(Ctx, collectionName, opts)
}

func (mongo *MongoClient) UpdateCollection(collectionName string, commands bson.D) error {
	db := mongo.client.Database(mongo.database)
	command := bson.D{{
		Key:   "collMod",
		Value: collectionName,
	}}
	for _, v := range commands {
		command = append(command, v)
	}
	var result bson.M
	return db.RunCommand(Ctx, command).Decode(&result)
}

func NewMongoClient(host string) *mongo.Client {
	uri := fmt.Sprintf(
		"%s://%s:%s@%s",
		settingsData.MONGO_CONNECTION,
		settingsData.MONGO_ROOT_USERNAME,
		settingsData.MONGO_ROOT_PASSWORD,
		host,
	)
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(Ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	return client
}

func NewConnection(host string, dbName string) *MongoClient {
	client := NewMongoClient(host)
	err := client.Ping(Ctx, nil)
	if err != nil {
		panic(err)
	}
	return newMongoClient(client, dbName)
}
