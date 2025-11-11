package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// MemoryCache is an in-memory LRU cache implementation
type MemoryCache struct {
	items    map[string]*cacheItem
	maxItems int
	mu       sync.RWMutex
}

type cacheItem struct {
	value      string
	expiration time.Time
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache(maxItems int) *MemoryCache {
	mc := &MemoryCache{
		items:    make(map[string]*cacheItem),
		maxItems: maxItems,
	}

	// Start cleanup goroutine
	go mc.cleanup()

	return mc
}

// Get retrieves a value from cache
func (m *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.items[key]
	if !exists {
		return "", errors.New("cache miss")
	}

	// Check expiration
	if time.Now().After(item.expiration) {
		return "", errors.New("cache expired")
	}

	return item.value, nil
}

// Set stores a value in cache
func (m *MemoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Evict oldest item if cache is full
	if len(m.items) >= m.maxItems {
		m.evictOldest()
	}

	m.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a value from cache
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.items, key)
	return nil
}

// evictOldest removes the oldest item from cache (must be called with lock held)
func (m *MemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range m.items {
		if oldestKey == "" || item.expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expiration
		}
	}

	if oldestKey != "" {
		delete(m.items, oldestKey)
	}
}

// cleanup runs periodically to remove expired items
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, item := range m.items {
			if now.After(item.expiration) {
				delete(m.items, key)
			}
		}
		m.mu.Unlock()
	}
}
