package mongo

import (
	"context"
	"fmt"
	"librus/model"
	"librus/mongo/client"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateOrUpdateLibrusAccount creates or updates a Librus account
func CreateOrUpdateLibrusAccount(login, password string) error {
	collection := client.Db.Collection("librus_account")

	account := bson.M{
		"_id":      login,
		"password": password,
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": login},
		bson.M{"$set": account},
		opts,
	)
	return err
}

// GetLibrusAccountsFromDatabase returns all Librus accounts
func GetLibrusAccountsFromDatabase() []model.LibrusAccount {
	collection := client.Db.Collection("librus_account")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	var accounts []model.LibrusAccount
	for cursor.Next(context.Background()) {
		var account model.LibrusAccount
		err = cursor.Decode(&account)
		if err != nil {
			log.Println(err)
			continue
		}
		accounts = append(accounts, account)
	}
	if err = cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return accounts
}

// GetTelegramUsersByLibrusLogin returns all telegram users for a Librus account
func GetTelegramUsersByLibrusLogin(librusLogin string) ([]model.TelegramUser, error) {
	collection := client.Db.Collection("telegram_user")
	filter := bson.M{
		"librus_login": librusLogin,
		"state":        model.StateAuthenticated,
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []model.TelegramUser
	for cursor.Next(context.Background()) {
		var user model.TelegramUser
		err = cursor.Decode(&user)
		if err != nil {
			log.Println(err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// FindTelegramUserByTelegramID finds a telegram user by telegram ID
func FindTelegramUserByTelegramID(telegramID int64) (*model.TelegramUser, error) {
	collection := client.Db.Collection("telegram_user")
	filter := bson.M{"telegram_id": telegramID}
	var user model.TelegramUser
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindLibrusAccount finds a Librus account by login
func FindLibrusAccount(login string) (*model.LibrusAccount, error) {
	collection := client.Db.Collection("librus_account")
	filter := bson.M{"_id": login}
	var account model.LibrusAccount
	err := collection.FindOne(context.Background(), filter).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// UpdateTelegramUserLanguage updates telegram user's language
func UpdateTelegramUserLanguage(telegramID int64, language string) error {
	collection := client.Db.Collection("telegram_user")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"telegram_id": telegramID},
		bson.M{"$set": bson.M{
			"language":       language,
			"last_active_at": time.Now(),
		}},
	)
	return err
}

// UpdateTelegramUserState updates telegram user's state
func UpdateTelegramUserState(telegramID int64, state model.UserState) error {
	collection := client.Db.Collection("telegram_user")
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"telegram_id": telegramID},
		bson.M{"$set": bson.M{
			"state":          state,
			"last_active_at": time.Now(),
		}},
	)
	return err
}

// CreateTelegramUserWithState creates a new telegram user with initial state
func CreateTelegramUserWithState(telegramID int64, state model.UserState) error {
	collection := client.Db.Collection("telegram_user")

	user := bson.M{
		"_id":            primitive.NewObjectID().Hex(),
		"telegram_id":    telegramID,
		"librus_login":   "", // Will be set during login
		"language":       "", // Will be set later
		"state":          state,
		"created_at":     time.Now(),
		"last_active_at": time.Now(),
	}

	_, err := collection.InsertOne(context.Background(), user)
	return err
}

// UpdateTelegramUserField updates a specific field for telegram user
func UpdateTelegramUserField(telegramID int64, field string, value interface{}) error {
	collection := client.Db.Collection("telegram_user")
	updateDoc := bson.M{
		field:            value,
		"last_active_at": time.Now(),
	}
	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"telegram_id": telegramID},
		bson.M{"$set": updateDoc},
	)
	return err
}

// GetTelegramUserCollection returns the telegram user collection
func GetTelegramUserCollection() *mongo.Collection {
	return client.Db.Collection("telegram_user")
}

// GetLibrusAccountCollection returns the librus account collection
func GetLibrusAccountCollection() *mongo.Collection {
	return client.Db.Collection("librus_account")
}

func AddMessagesToDatabase(messages []model.Message, librusLogin string) ([]model.Message, error) {
	collection := client.Db.Collection("message")

	// Find existing messages by ID (regardless of librus_login since _id must be unique)
	existingMessages := make(map[string]bool)
	cursor, err := collection.Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": getIds(messages)}},
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
				"_id":             m.Id,
				"type":            m.Type,
				"link":            m.Link,
				"author":          m.Author,
				"title":           m.Title,
				"content":         m.Content,
				"date":            primitive.NewDateTimeFromTime(m.Date),
				"librus_login":    m.LibrusLogin,
				"attachments_dir": m.AttachmentsDir,
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

// IsMessageSentToUser checks if a message was already sent to a telegram user
func IsMessageSentToUser(telegramUserID, messageID string) bool {
	collection := client.Db.Collection("user_message_status")
	filter := bson.M{
		"telegram_user_id": telegramUserID,
		"message_id":       messageID,
	}

	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Printf("Error checking message status: %v", err)
		return false
	}

	return count > 0
}

// MarkMessageAsSent marks a message as sent to a telegram user
func MarkMessageAsSent(telegramUserID, messageID string) error {
	collection := client.Db.Collection("user_message_status")

	status := bson.M{
		"_id":              primitive.NewObjectID().Hex(),
		"telegram_user_id": telegramUserID,
		"message_id":       messageID,
		"sent_at":          time.Now(),
	}

	_, err := collection.InsertOne(context.Background(), status)
	return err
}

// DeleteTelegramUserByTelegramID deletes a telegram user and related data
func DeleteTelegramUserByTelegramID(telegramID int64) error {
	// Find the telegram user first
	telegramUser, err := FindTelegramUserByTelegramID(telegramID)
	if err != nil || telegramUser == nil {
		return err
	}

	// Delete telegram user
	collection := client.Db.Collection("telegram_user")
	_, err = collection.DeleteOne(context.Background(), bson.M{"telegram_id": telegramID})
	if err != nil {
		return err
	}

	// Delete user message statuses
	collection = client.Db.Collection("user_message_status")
	_, err = collection.DeleteMany(context.Background(), bson.M{"telegram_user_id": telegramUser.Id})
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
