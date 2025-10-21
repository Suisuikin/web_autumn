package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"

	"rip/internal/pkg/config"
)

type Client struct {
	client *redis.Client
}

func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{client: rdb}, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// JWT blacklist methods
func (c *Client) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	key := fmt.Sprintf("jwt:blacklist:%s", token)
	return c.Set(ctx, key, "1", expiration)
}

func (c *Client) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("jwt:blacklist:%s", token)
	return c.Exists(ctx, key)
}
