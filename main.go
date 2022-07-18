package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/juliotorresmoreno/zemona/config"
	"github.com/juliotorresmoreno/zemona/db"
	"github.com/juliotorresmoreno/zemona/router"
)

func main() {
	godotenv.Load()

	mongo, err := db.NewConnectionMongo()
	if err == nil {
		db.PrepareBD(mongo)
		db.PreloadData(mongo)
		mongo.Disconnect(context.Background())
	}

	config := config.GetConfig()
	handler := router.NewRouter()

	srv := &http.Server{
		Handler:      handler,
		Addr:         config.Addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Listening on", config.Addr)
	err = srv.ListenAndServe()
	log.Fatal(err)
}
