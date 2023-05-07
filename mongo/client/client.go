package client

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"librus/helper"
	"log"
)

var Db *mongo.Database

func init() {
	clientOptions := options.Client().ApplyURI(
		fmt.Sprintf(
			"mongodb://%s:27017",
			helper.GetEnv("MONGO_HOST", "localhost")),
	)
	var err error
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	Db = client.Database("librus")
}
