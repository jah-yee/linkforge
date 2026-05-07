package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(ctx context.Context, addr string) (*Cache, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Cache{client: client, ttl: 24 * time.Hour}, nil
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) GetURL(ctx context.Context, code string) (string, error) {
	v, err := c.client.Get(ctx, key(code)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("cache get: %w", err)
	}
	return v, nil
}

func (c *Cache) SetURL(ctx context.Context, code, rawURL string) error {
	if err := c.client.Set(ctx, key(code), rawURL, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	return nil
}

func key(code string) string {
	return "linkforge:link:" + code
}
