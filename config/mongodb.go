package config

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var LwDB *mongo.Database

func InitMongoDB() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(EnvManager.MongoURI))
	if err != nil {
		log.Fatal(err)
	}
	Client = client
	LwDB = client.Database(EnvManager.MongoDBLearnWiz)
}

func GetLwDB() *mongo.Database {
	return LwDB
}
