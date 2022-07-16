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
	"github.com/juliotorresmoreno/zemona/db"
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

	router.HandleFunc("/login", c.handleTwitterLogin).
		Methods("GET")

	router.HandleFunc("/{profileId}/profile", c.handleGetProfileInformation).
		Methods("GET")

	router.HandleFunc("/{profileId}/tweets", c.handleGetLatestNTweetsForProfile).
		Methods("GET")

	router.HandleFunc("/{profileId}/profile", c.handleModifyProfileInformation).
		Methods("PATCH")

	router.HandleFunc("/{profileId}/requests", c.handleGetProfileRequests).
		Methods("GET")

	return router
}

func (c *twitterController) handleTwitterLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	twitterClient := twitter.NewTwitterOauthClient()
	err := twitterClient.DoAuth()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
		return
	}
}

func queryGetProfileInformation(client *mongo.Client, profileId string) (int64, bool, error) {
	config := config.GetConfig()

	queryProfileInformation := &models.QueryProfileInformation{}
	tableName := queryProfileInformation.TableName()
	collection := client.Database(config.MongoDBDatabase).
		Collection(tableName)
	filter := bson.D{bson.E{Key: "_id", Value: profileId}}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := collection.FindOne(ctx, filter)
	result.Decode(queryProfileInformation)

	return queryProfileInformation.Count, queryProfileInformation.ID != "", nil
}

func registerGetProfileInformation(client *mongo.Client, profileId string) error {
	config := config.GetConfig()
	queryProfileInformation := &models.QueryProfileInformation{}
	tableName := queryProfileInformation.TableName()
	filter := bson.D{bson.E{Key: "_id", Value: profileId}}
	collection := client.Database(config.MongoDBDatabase).
		Collection(tableName)
	count, exists, err := queryGetProfileInformation(client, profileId)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if exists {
		queryProfileInformation.ID = profileId
		queryProfileInformation.Count = count + 1
		update := bson.D{bson.E{
			Key:   "$set",
			Value: queryProfileInformation,
		}}

		_, err = collection.UpdateOne(ctx, filter, update)
	} else {
		queryProfileInformation.ID = profileId
		queryProfileInformation.Count = 1
		_, err = collection.InsertOne(ctx, queryProfileInformation)
	}
	if err != nil {
		return err
	}

	return nil
}

func getProfileInformationFromBD(client *mongo.Client, profileId string) (*models.Profile, error) {
	profile := &models.Profile{}
	config := config.GetConfig()
	tableName := profile.TableName()
	filter := bson.D{bson.E{Key: "username", Value: profileId}}
	collection := client.Database(config.MongoDBDatabase).
		Collection(tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := collection.FindOne(ctx, filter)

	err := result.Decode(profile)
	return profile, err
}

func getProfileInformationToBD(client *mongo.Client, profile *models.Profile) error {
	config := config.GetConfig()
	tableName := profile.TableName()
	collection := client.Database(config.MongoDBDatabase).
		Collection(tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, profile)

	return err
}

// GetProfileInformation: Obtener información de perfil
func (c *twitterController) handleGetProfileInformation(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var result interface{}
	t := twitter.NewTwitterClient(&twitter.TwitterClientArgs{})
	defer t.Close()

	profileId := mux.Vars(r)["profileId"]
	mongo, err := db.NewConnectionMongo()
	if err != nil {
		return
	}
	defer mongo.Disconnect(context.Background())
	profile, err := getProfileInformationFromBD(mongo, profileId)
	if err == nil && profile.ID != "" {
		result = profile
	} else {
		profile, err := t.GetProfileInformation(profileId)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
			return
		}

		if profile.ID == "" {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Del("Content-Type")
			return
		}
		response := &models.Profile{
			ID:        profile.ID,
			Username:  profile.Username,
			Name:      profile.Name,
			CreatedAt: profile.CreatedAt,
		}
		getProfileInformationToBD(mongo, response)
		result = response
	}

	defer func() {
		mongo, err := db.NewConnectionMongo()
		if err != nil {
			return
		}
		defer mongo.Disconnect(context.Background())
		err = registerGetProfileInformation(mongo, profileId)
		if err != nil {
			log.Println(err)
		}
	}()

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(result)
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

// ModifyProfileInformation: Modificar la información del perfil
func (c *twitterController) handleModifyProfileInformation(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	t := twitter.NewTwitterClient(&twitter.TwitterClientArgs{})
	defer t.Close()

	values := helpers.GetPostParams(r)

	profileId := mux.Vars(r)["profileId"]
	name := values.Get("name")

	mongo, err := db.NewConnectionMongo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("service not available")
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
	}
	defer mongo.Disconnect(context.Background())

	profile, err := getProfileInformationFromBD(mongo, profileId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("service not available")
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
	}

	profile.Name = name
	err = updateProfileInformation(mongo, profile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("service not available")
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
	}

	w.Header().Del("Content-Type")
	w.WriteHeader(http.StatusNoContent)
}

// GetProfileRequests: Obtener solicitudes de perfil
func (c *twitterController) handleGetProfileRequests(w http.ResponseWriter, r *http.Request) {
	t := twitter.NewTwitterClient(&twitter.TwitterClientArgs{})
	defer t.Close()

	profileId := mux.Vars(r)["profileId"]
	mongo, err := db.NewConnectionMongo()
	if err != nil {
		return
	}
	defer mongo.Disconnect(context.Background())

	count, _, _ := queryGetProfileInformation(mongo, profileId)

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profileId": profileId,
		"count":     count,
	})
}
