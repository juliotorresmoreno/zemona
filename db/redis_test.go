package db_test

import (
	"context"
	"testing"

	"github.com/juliotorresmoreno/zemona/db"
)

func TestNewConnectionRedis(t *testing.T) {
	client, err := db.NewConnectionRedis()
	if err != nil {
		t.Error("Redis connection failed. " + err.Error())
		return
	}
	ctx := context.Background()
	result, _ := client.Ping(ctx).Result()
	if result != "PONG" {
		t.Error("Redis connection failed")
	}
}
