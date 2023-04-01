package telegram

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"librus/helper"
	"librus/librus"
	"log"
	"time"
)

type User struct {
	Login      string `bson:"login"`
	Password   string `bson:"password"`
	TelegramID int64  `bson:"telegram_id"`
}

func init() {
	clientOptions := options.Client().ApplyURI(
		fmt.Sprintf(
			"mongodb://%s:27017",
			helper.GetEnv("MONGO_HOST", "localhost")),
	)
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
}

func addUserToDatabase(user User) {
	collection := client.Database("librus").Collection("user")
	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Fatal(err)
	}
}

func getUsersFromDatabase() []User {
	collection := client.Database("librus").Collection("user")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())
	var users []User
	for cursor.Next(context.Background()) {
		var user User
		err = cursor.Decode(&user)
		if err != nil {
			log.Println(err)
			continue
		}
		users = append(users, user)
	}
	if err = cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return users
}

func findUserByTelegramID(telegramID int64) (*User, error) {
	collection := client.Database("librus").Collection("user")
	filter := bson.M{"telegram_id": telegramID}
	var user User
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func addMessagesToDatabase(messages []librus.Message) ([]librus.Message, error) {
	collection := client.Database("librus").Collection("message")

	// Create index on "link" field
	indexModel := mongo.IndexModel{
		Keys: bson.M{"link": 1},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := collection.Indexes().CreateOne(context.Background(), indexModel, opts)
	if err != nil {
		return nil, err
	}

	// Find existing messages
	existingMessages := make(map[string]bool)
	cursor, err := collection.Find(context.Background(), bson.M{"link": bson.M{"$in": getLinks(messages)}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var m librus.Message
		err := cursor.Decode(&m)
		if err != nil {
			return nil, err
		}
		existingMessages[m.Link] = true
	}

	// Insert new messages using bulk.Write
	var newMessages []librus.Message
	var bulkOps []mongo.WriteModel
	for _, m := range messages {
		if _, ok := existingMessages[m.Link]; !ok {
			newMessages = append(newMessages, m)
			doc := bson.M{
				"link":        m.Link,
				"author":      m.Author,
				"title":       m.Title,
				"content":     m.Content,
				"date":        primitive.NewDateTimeFromTime(m.Date),
				"telegram_id": m.TelegramID,
			}
			bulkOps = append(bulkOps, mongo.NewInsertOneModel().SetDocument(doc))
		}
	}
	if len(bulkOps) > 0 {
		_, err = collection.BulkWrite(context.Background(), bulkOps)
		if err != nil {
			return nil, err
		}
	}

	// If no new messages were added, return empty list of messages
	if len(newMessages) == 0 {
		return newMessages, nil
	}

	return newMessages, nil
}

func getLinks(messages []librus.Message) []string {
	var links []string
	for _, m := range messages {
		links = append(links, m.Link)
	}
	return links
}

func deleteUserByTelegramID(telegramID int64) error {
	collection := client.Database("librus").Collection("user")
	filter := bson.M{"telegram_id": telegramID}
	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	collection = client.Database("librus").Collection("message")
	_, err = collection.DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}
