package client

import (
	"context"
	"fmt"
	"librus/helper"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Db *mongo.Database

func init() {
	mongoHost := helper.GetEnv("MONGO_HOST", "localhost")
	mongoPort := "27017"
	mongoURI := fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort)

	clientOptions := options.Client().ApplyURI(mongoURI)

	// Set a timeout context for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client for %s: %v", mongoURI, err)
	}

	// Ping the MongoDB server to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to connect to MongoDB at %s:%s: %v", mongoHost, mongoPort, err)
	}

	log.Printf("Successfully connected to MongoDB at %s:%s", mongoHost, mongoPort)
	Db = client.Database("librus")
}
