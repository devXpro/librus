package client

import (
	"context"
	"fmt"
	"time"

	"librus/helper"

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
		// Use fmt.Printf since logger might not be initialized yet
		fmt.Printf("FATAL: Failed to create MongoDB client for %s: %v\n", mongoURI, err)
		panic(err)
	}

	// Ping the MongoDB server to verify connection
	if err = client.Ping(ctx, nil); err != nil {
		// Use fmt.Printf since logger might not be initialized yet
		fmt.Printf("FATAL: Failed to connect to MongoDB at %s:%s: %v\n", mongoHost, mongoPort, err)
		panic(err)
	}

	// Use fmt.Printf since logger might not be initialized yet
	fmt.Printf("Successfully connected to MongoDB at %s:%s\n", mongoHost, mongoPort)
	Db = client.Database("librus")
}
