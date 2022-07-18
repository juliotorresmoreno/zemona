package helpers

import (
	"context"
	"net/http"
	"time"

	"github.com/juliotorresmoreno/zemona/config"
	"github.com/juliotorresmoreno/zemona/db"
	"github.com/juliotorresmoreno/zemona/models"
	"go.mongodb.org/mongo-driver/bson"
)

func Auth(w http.ResponseWriter, r *http.Request) (*models.Session, error) {
	session := &models.Session{}
	redisCli, err := db.NewConnectionRedis()
	if err != nil {
		return session, err
	}
	defer redisCli.Close()

	token := GetToken(r)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	username := redisCli.Get(ctx, token).Val()

	mongo, err := db.NewConnectionMongo()
	if err != nil {
		return session, err
	}
	defer mongo.Disconnect(context.Background())

	profile := &models.Profile{}
	config := config.GetConfig()
	tableName := profile.TableName()
	filter := bson.D{bson.E{Key: "username", Value: username}}
	collection := mongo.Database(config.MongoDBDatabase).
		Collection(tableName)

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := collection.FindOne(ctx, filter)

	err = result.Decode(profile)
	session.Profile = profile
	session.Token = token

	return session, err
}
