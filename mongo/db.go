package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"librus/model"
	"librus/mongo/client"
	"log"
	"time"
)

func AddUserToDatabase(login string, password string, telegramId int64) {
	collection := client.Db.Collection("user")

	update := bson.M{
		"$addToSet": bson.M{"telegram_ids": telegramId},
	}
	filter := bson.M{"login": login, "password": password, "telegram_ids": bson.M{"$ne": telegramId}}

	opt := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(context.Background(), filter, update, opt)

	if err != nil {
		log.Fatal(err)
	}
}

func GetUsersFromDatabase() []model.User {
	collection := client.Db.Collection("user")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())
	var users []model.User
	for cursor.Next(context.Background()) {
		var user model.User
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

func FindUserByTelegramID(telegramID int64) (*model.User, error) {
	collection := client.Db.Collection("user")
	filter := bson.M{"telegram_ids": telegramID}
	var user model.User
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
	collection := client.Db.Collection("user")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"telegram_ids": telegramID},
		bson.M{"$set": bson.M{"language": language}},
	)
	if err != nil {
		return err
	}
	return nil
}

func AddMessagesToDatabase(messages []model.Message, userId string) ([]model.Message, error) {
	collection := client.Db.Collection("message")

	// Find existing messages
	existingMessages := make(map[string]bool)
	cursor, err := collection.Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": getIds(messages)}, "user_id": userId},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var m model.Message
		err = cursor.Decode(&m)
		if err != nil {
			return nil, err
		}
		existingMessages[m.Id] = true
	}

	// Insert new messages using bulk.Write
	var newMessages []model.Message
	var bulkOps []mongo.WriteModel
	for _, m := range messages {
		if _, ok := existingMessages[m.Id]; !ok {
			newMessages = append(newMessages, m)
			doc := bson.M{
				"_id":     m.Id,
				"link":    m.Link,
				"author":  m.Author,
				"title":   m.Title,
				"content": m.Content,
				"date":    primitive.NewDateTimeFromTime(m.Date),
				"user_id": m.UserID,
				"type":    m.Type,
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

func getIds(messages []model.Message) []string {
	var ids []string
	for _, m := range messages {
		ids = append(ids, m.Id)
	}
	return ids
}

func DeleteUserByTelegramID(telegramID int64) error {
	collection := client.Db.Collection("user")
	filter := bson.M{"telegram_ids": telegramID}
	_, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	collection = client.Db.Collection("message")
	_, err = collection.DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}

func DeleteAllMessages() error {
	collection := client.Db.Collection("message")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to delete messages: %v", err)
	}
	fmt.Printf("Deleted %d documents from collection", result.DeletedCount)
	return nil
}
