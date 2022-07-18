package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/juliotorresmoreno/zemona/config"
	"github.com/juliotorresmoreno/zemona/db"
	"github.com/juliotorresmoreno/zemona/helpers"
	"github.com/juliotorresmoreno/zemona/models"
	"go.mongodb.org/mongo-driver/bson"
)

type sessionController struct {
}

func NewSessionController() http.Handler {
	c := &sessionController{}
	router := mux.NewRouter().StrictSlash(false)

	router.HandleFunc("/", c.handleGetSession).
		Methods("GET")

	router.HandleFunc("/", c.handlePostLogin).
		Methods("POST")

	return router
}

func (c *sessionController) handleGetSession(w http.ResponseWriter, r *http.Request) {
	session, err := helpers.Auth(w, r)
	if err != nil || session.Token == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": session.Token,
		"profile": map[string]interface{}{
			"id":       session.Profile.ID,
			"username": session.Profile.Username,
			"name":     session.Profile.Name,
		},
	})
}

type Credentials struct {
	Username string
	Password string
}

func handlePostLoginValidate(credentials *Credentials) error {
	config := config.GetConfig()

	if credentials.Username != config.Username {
		return errors.New("usuario o contrase;a invalida")
	}

	h := sha256.New()
	h.Write([]byte(credentials.Password))
	pwdSha256 := fmt.Sprintf("%x", h.Sum(nil))

	pwd := credentials.Password

	if pwd != config.Password && pwd != pwdSha256 {
		return errors.New("usuario o contrase;a invalida")
	}

	return nil
}

func (c *sessionController) handlePostLogin(w http.ResponseWriter, r *http.Request) {
	session := &models.Session{}

	values := helpers.GetPostParams(r)
	username := values.Get("username")
	password := values.Get("password")

	err := handlePostLoginValidate(&Credentials{
		username, password,
	})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
		return
	}

	mongo, err := db.NewConnectionMongo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("service not available")
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
		return
	}
	defer mongo.Disconnect(context.Background())

	profile := &models.Profile{}
	config := config.GetConfig()
	tableName := profile.TableName()
	filter := bson.D{bson.E{Key: "username", Value: username}}
	collection := mongo.Database(config.MongoDBDatabase).
		Collection(tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := collection.FindOne(ctx, filter)

	err = result.Decode(profile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("service not available")
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
		return
	}

	session.Profile = profile
	session.Token, err = makeTokenForUser(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		err := errors.New("service not available")
		json.NewEncoder(w).Encode(helpers.MakeHTTPError(err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": session.Token,
		"profile": map[string]interface{}{
			"id":       session.Profile.ID,
			"username": session.Profile.Username,
			"name":     session.Profile.Name,
		},
	})
}

func makeTokenForUser(username string) (string, error) {
	redisCli, err := db.NewConnectionRedis()
	if err != nil {
		return "", err
	}
	defer redisCli.Close()

	token := RandStringRunes(64)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	redisCli.Set(ctx, token, username, 2*time.Hour)

	return token, nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
