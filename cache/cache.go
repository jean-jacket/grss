package cache

import (
	"context"
	"encoding/json"
	"time"
)

// Cache is the interface for cache implementations
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// TryGet is a helper function that gets from cache or executes getter function
func TryGet[T any](ctx context.Context, c Cache, key string, getter func() (T, error), ttl time.Duration) (T, error) {
	var result T

	// Try to get from cache
	cached, err := c.Get(ctx, key)
	if err == nil && cached != "" {
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			return result, nil
		}
	}

	// Cache miss or error - execute getter
	result, err = getter()
	if err != nil {
		return result, err
	}

	// Store in cache
	data, err := json.Marshal(result)
	if err == nil {
		_ = c.Set(ctx, key, string(data), ttl)
	}

	return result, nil
}
