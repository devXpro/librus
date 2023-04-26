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
	Language   string `bson:"language"`
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

func UpdateUserLanguageByTelegramID(telegramID int64, language string) error {
	collection := client.Database("librus").Collection("user")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"telegram_id": telegramID},
		bson.M{"$set": bson.M{"language": language}},
	)
	if err != nil {
		return err
	}
	return nil
}

func addMessagesToDatabase(messages []librus.Message, telegramId int64) ([]librus.Message, error) {
	collection := client.Database("librus").Collection("message")

	// Find existing messages
	existingMessages := make(map[string]bool)
	cursor, err := collection.Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": getIds(messages)}, "telegram_id": telegramId},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var m librus.Message
		err = cursor.Decode(&m)
		if err != nil {
			return nil, err
		}
		existingMessages[m.Id] = true
	}

	// Insert new messages using bulk.Write
	var newMessages []librus.Message
	var bulkOps []mongo.WriteModel
	for _, m := range messages {
		if _, ok := existingMessages[m.Id]; !ok {
			newMessages = append(newMessages, m)
			doc := bson.M{
				"_id":         m.Id,
				"link":        m.Link,
				"author":      m.Author,
				"title":       m.Title,
				"content":     m.Content,
				"date":        primitive.NewDateTimeFromTime(m.Date),
				"telegram_id": m.TelegramID,
				"type":        m.Type,
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

func deleteAllMessages() error {
	collection := client.Database("librus").Collection("message")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to delete messages: %v", err)
	}
	fmt.Printf("Deleted %d documents from collection", result.DeletedCount)
	return nil
}

func getIds(messages []librus.Message) []string {
	var ids []string
	for _, m := range messages {
		ids = append(ids, m.Id)
	}
	return ids
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
