package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/juliotorresmoreno/zemona/config"
	"github.com/juliotorresmoreno/zemona/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewConnectionMongo() (*mongo.Client, error) {
	config := config.GetConfig()

	uri := config.MongoDBUri
	conf := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), conf)
	return client, err
}

func existsKey(indexes []index, name string) bool {
	for _, idx := range indexes {
		value := fmt.Sprintf("%v", idx.Key[name])
		if value == "1" {
			return true
		}
	}
	return false
}

type index struct {
	Key map[string]interface{}
}

func PrepareBD(client *mongo.Client) {
	config := config.GetConfig()

	profile := &models.Profile{}
	collection := client.
		Database(config.MongoDBDatabase).
		Collection(profile.TableName())

	indexes, _ := collection.Indexes().List(context.Background())
	result := make([]index, 0, 10)
	indexes.All(context.Background(), &result)

	if !existsKey(result, "username") {
		mod := mongo.IndexModel{
			Keys: bson.M{
				"username": 1,
			},
			Options: options.Index().SetUnique(true).SetName("profile_username_uq"),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_, err := collection.Indexes().CreateOne(ctx, mod)
		if err != nil {
			log.Println(err)
		}
	}

	log.Println("BD has prepared!")
}
