package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache(10)
	ctx := context.Background()

	// Test set and get
	err := cache.Set(ctx, "key1", "value1", 1*time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%s'", val)
	}
}

func TestMemoryCache_CacheMiss(t *testing.T) {
	cache := NewMemoryCache(10)
	ctx := context.Background()

	// Test cache miss
	_, err := cache.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for cache miss, got nil")
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache(10)
	ctx := context.Background()

	// Set with short TTL
	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should exist immediately
	_, err = cache.Get(ctx, "key1")
	if err != nil {
		t.Error("Expected key to exist, but got error")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(ctx, "key1")
	if err == nil {
		t.Error("Expected error for expired key, got nil")
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(10)
	ctx := context.Background()

	// Set a value
	cache.Set(ctx, "key1", "value1", 1*time.Minute)

	// Verify it exists
	_, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Error("Expected key to exist")
	}

	// Delete it
	err = cache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = cache.Get(ctx, "key1")
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}

func TestMemoryCache_Eviction(t *testing.T) {
	cache := NewMemoryCache(3) // Small cache
	ctx := context.Background()

	// Fill cache beyond capacity
	cache.Set(ctx, "key1", "value1", 1*time.Minute)
	cache.Set(ctx, "key2", "value2", 1*time.Minute)
	cache.Set(ctx, "key3", "value3", 1*time.Minute)
	cache.Set(ctx, "key4", "value4", 1*time.Minute)

	// Should have evicted oldest item
	// Note: This is a simple test - actual LRU would need more sophisticated testing
	cache.mu.RLock()
	count := len(cache.items)
	cache.mu.RUnlock()

	if count > 3 {
		t.Errorf("Expected max 3 items, got %d", count)
	}
}

func TestMemoryCache_Concurrent(t *testing.T) {
	cache := NewMemoryCache(100)
	ctx := context.Background()

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := string(rune('a' + n))
			cache.Set(ctx, key, "value", 1*time.Minute)
			cache.Get(ctx, key)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestTryGet(t *testing.T) {
	cache := NewMemoryCache(10)
	ctx := context.Background()

	callCount := 0
	getter := func() (string, error) {
		callCount++
		return "computed", nil
	}

	// First call should execute getter
	result, err := TryGet(ctx, cache, "key1", getter, 1*time.Minute)
	if err != nil {
		t.Fatalf("TryGet failed: %v", err)
	}
	if result != "computed" {
		t.Errorf("Expected 'computed', got '%s'", result)
	}
	if callCount != 1 {
		t.Errorf("Expected getter to be called once, called %d times", callCount)
	}

	// Second call should use cache
	result, err = TryGet(ctx, cache, "key1", getter, 1*time.Minute)
	if err != nil {
		t.Fatalf("TryGet failed: %v", err)
	}
	if result != "computed" {
		t.Errorf("Expected 'computed', got '%s'", result)
	}
	if callCount != 1 {
		t.Errorf("Expected getter to still be called once, called %d times", callCount)
	}
}
