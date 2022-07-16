package config

import "os"

type Config struct {
	Addr                string
	TwitterClientID     string
	TwitterClientSecret string

	TwitterApiKey       string
	TwitterApiKeySecret string
	RedirectURL         string

	MongoDBUri      string
	MongoDBDatabase string
	RedisUri        string
}

func getValue(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetConfig() *Config {
	c := &Config{
		Addr:                getValue("ADDR", ":8080"),
		TwitterClientID:     getValue("TWITTER_CLIENT_ID", ""),
		TwitterClientSecret: getValue("TWITTER_CLIENT_SECRET", ""),
		TwitterApiKey:       getValue("TWITTER_API_KEY", ""),
		TwitterApiKeySecret: getValue("TWITTER_API_KEY_SECRET", ""),
		RedirectURL:         getValue("TWITTER_REDIRECT_URL", ""),
		MongoDBUri:          getValue("MONGODB_URI", ""),
		MongoDBDatabase:     getValue("MONGODB_DATABASE", ""),
		RedisUri:            getValue("REDIS_URI", ""),
	}

	return c
}
