package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/juliotorresmoreno/zemona/config"
	"github.com/juliotorresmoreno/zemona/helpers"
	"github.com/juliotorresmoreno/zemona/integrations/twitter"
	"github.com/juliotorresmoreno/zemona/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type twitterController struct {
}

func NewTwitterController() http.Handler {
	c := &twitterController{}
	router := mux.NewRouter()

	router.HandleFunc("/{profileId}/tweets", c.handleGetLatestNTweetsForProfile).
		Methods("GET")

	return router
}

// GetLatestNTweetsForProfile: Obtenga los N Tweets más recientes para el perfil
func (c *twitterController) handleGetLatestNTweetsForProfile(w http.ResponseWriter, r *http.Request) {
	t := twitter.NewTwitterClient(&twitter.TwitterClientArgs{})
	defer t.Close()

	elements, _ := strconv.Atoi(r.URL.Query().Get("n"))
	if elements < 10 {
		elements = 10
	}
	if elements > 100 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(errors.New("el numero maximo de elementos es 100")))
		return
	}
	profileId := mux.Vars(r)["profileId"]
	result, err := t.GetLatestNTweetsForProfile(profileId, int64(elements))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(result)
}

func updateProfileInformation(client *mongo.Client, profile *models.Profile) error {
	config := config.GetConfig()
	tableName := profile.TableName()
	filter := bson.D{bson.E{Key: "username", Value: profile.Username}}
	collection := client.Database(config.MongoDBDatabase).
		Collection(tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	update := bson.D{bson.E{
		Key:   "$set",
		Value: profile,
	}}
	_, err := collection.UpdateOne(ctx, filter, update)

	return err
}
