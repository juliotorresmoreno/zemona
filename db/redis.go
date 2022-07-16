package db

import (
	"github.com/go-redis/redis/v8"
	"github.com/juliotorresmoreno/zemona/config"
)

func NewConnectionRedis() (*redis.Client, error) {
	config := config.GetConfig()
	opts, err := redis.ParseURL(config.RedisUri)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opts), nil
}
