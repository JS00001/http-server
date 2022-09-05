package db

import (
	"context"
	"http-server/config"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var VerificationCodes *mongo.Collection
var Users *mongo.Collection
var Apps *mongo.Collection

func Connect() error {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.GetConfig("MONGODB_URL")))

	// If there is an error connection to the MongoDB, return the error
	if err != nil {
		return err
	}

	// Check the connection
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return err
	}

	VerificationCodes = client.Database("app").Collection("verificationCodes")
	Users = client.Database("app").Collection("users")
	Apps = client.Database("app").Collection("apps")

	log.Println("Connected to MongoDB in mode:", config.GetConfig("ENV"))

	// Return nil if no error
	return nil
}
