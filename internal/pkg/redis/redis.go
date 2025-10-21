package redis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func ConnectRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}

	log.Println("Подключено к Redis")
	return client
}

// BlacklistToken добавляет токен в blacklist
func BlacklistToken(client *redis.Client, token string, expiration time.Duration) error {
	return client.Set(ctx, "blacklist:"+token, "1", expiration).Err()
}

// IsTokenBlacklisted проверяет, находится ли токен в blacklist
func IsTokenBlacklisted(client *redis.Client, token string) (bool, error) {
	result, err := client.Get(ctx, "blacklist:"+token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == "1", nil
}
