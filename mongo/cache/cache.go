package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	mongo2 "librus/mongo/client"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Cache struct {
	Key       string    `bson:"_id"`
	Value     string    `bson:"value"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func GetCachedTranslation(targetLanguage, text string) (string, error) {
	key := generateCacheKey(targetLanguage, text)

	col := mongo2.Db.Collection("translation_cache")
	var cache Cache
	err := col.FindOne(context.Background(), bson.M{"_id": key}).Decode(&cache)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	return cache.Value, nil
}

func SetCachedTranslation(targetLanguage, text, translation string) error {
	key := generateCacheKey(targetLanguage, text)

	col := mongo2.Db.Collection("translation_cache")
	cache := Cache{
		Key:       key,
		Value:     translation,
		UpdatedAt: time.Now(),
	}

	_, err := col.UpdateOne(
		context.Background(),
		bson.M{"_id": key},
		bson.M{"$set": cache},
		options.Update().SetUpsert(true),
	)
	return err
}

func generateCacheKey(targetLanguage, text string) string {
	hash := md5.Sum([]byte(targetLanguage + text))
	return hex.EncodeToString(hash[:])
}

// New functions to work with arbitrary cache keys

func GetCachedTranslationByKey(key string) (string, error) {
	col := mongo2.Db.Collection("translation_cache")
	var cache Cache
	err := col.FindOne(context.Background(), bson.M{"_id": key}).Decode(&cache)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}

	return cache.Value, nil
}

func SetCachedTranslationByKey(key, translation string) error {
	col := mongo2.Db.Collection("translation_cache")
	cache := Cache{
		Key:       key,
		Value:     translation,
		UpdatedAt: time.Now(),
	}

	_, err := col.UpdateOne(
		context.Background(),
		bson.M{"_id": key},
		bson.M{"$set": cache},
		options.Update().SetUpsert(true),
	)
	return err
}
