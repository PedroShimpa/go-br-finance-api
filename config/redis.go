package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisHost + ":6379",
	})

	// Test connection
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Println("⚠️  Redis não conectado, usando modo sem persistência:", err)
		RedisClient = nil // Set to nil to indicate no Redis
		return
	}

	log.Println("✅ Conectado ao Redis")
}
