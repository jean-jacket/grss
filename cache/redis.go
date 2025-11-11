package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-backed cache implementation
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
	}, nil
}

// Get retrieves a value from Redis cache
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	// Check TTL tracking key
	ttlKey := "rsshub:cacheTtl:" + key
	ttl, err := r.client.TTL(ctx, ttlKey).Result()
	if err != nil || ttl <= 0 {
		return "", errors.New("cache miss or expired")
	}

	// Refresh TTL on hit
	r.client.Expire(ctx, ttlKey, ttl)

	// Get actual value
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

// Set stores a value in Redis cache
func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	// Store the actual value (no expiration on the data key)
	if err := r.client.Set(ctx, key, value, 0).Err(); err != nil {
		return err
	}

	// Store TTL tracking key
	ttlKey := "rsshub:cacheTtl:" + key
	if err := r.client.Set(ctx, ttlKey, "1", ttl).Err(); err != nil {
		return err
	}

	return nil
}

// Delete removes a value from Redis cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	// Delete both the data key and TTL tracking key
	ttlKey := "rsshub:cacheTtl:" + key
	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, ttlKey)
	_, err := pipe.Exec(ctx)
	return err
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}
