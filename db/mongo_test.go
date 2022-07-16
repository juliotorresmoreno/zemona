package db_test

import (
	"testing"

	"github.com/juliotorresmoreno/zemona/db"
)

func TestNewConnectionMongo(t *testing.T) {
	_, err := db.NewConnectionMongo()
	if err != nil {
		t.Error("Mongodb connection failed. " + err.Error())
	}
}
