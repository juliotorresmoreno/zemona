package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
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

type profileController struct {
}

func NewProfileController() http.Handler {
	c := &profileController{}
	router := mux.NewRouter()

	router.HandleFunc("/{profileId}", c.handleGetProfileInformation).
		Methods("GET")

	router.HandleFunc("/{profileId}", c.handleModifyProfileInformation).
		Methods("PATCH")

	router.HandleFunc("/{profileId}/requests", c.handleGetProfileRequests).
		Methods("GET")

	return router
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
func (c *profileController) handleGetProfileInformation(w http.ResponseWriter, r *http.Request) {
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

func handleModifyProfileInformationValidate(profile *models.Profile) error {
	if profile.Name == "" {
		return errors.New("name is not valid")
	}
	if profile.Description == "" {
		return errors.New("description is not valid")
	}
	description := strings.ToLower(profile.Description)
	if strings.Contains(description, "script") {
		return errors.New("description is insecure. The word script is not permitted")
	}
	return nil
}

// ModifyProfileInformation: Modificar la información del perfil
func (c *profileController) handleModifyProfileInformation(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	session, err := helpers.Auth(w, r)
	if err != nil || session.Token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		err := errors.New("unauthorized")
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
	}

	t := twitter.NewTwitterClient(&twitter.TwitterClientArgs{})
	defer t.Close()

	values := helpers.GetPostParams(r)

	profileId := mux.Vars(r)["profileId"]
	name := values.Get("name")
	description := values.Get("description")
	imageSrc := values.Get("image_src")

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
	profile.ImageSrc = imageSrc
	profile.Description = description
	err = handleModifyProfileInformationValidate(profile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
	}

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
func (c *profileController) handleGetProfileRequests(w http.ResponseWriter, r *http.Request) {
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
