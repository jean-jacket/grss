package client

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jean-jacket/grss/config"
)

func TestClient_GetSuccess(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create client
	cfg := &config.Config{
		RequestRetry:   2,
		RequestTimeout: 5 * time.Second,
	}
	client := New(cfg)

	// Test GET
	data, err := client.Get(server.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if string(data) != "test response" {
		t.Errorf("Expected 'test response', got '%s'", string(data))
	}
}

func TestClient_GetWithHeaders(t *testing.T) {
	// Create test server that checks headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Error("Missing or wrong Authorization header")
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("Missing User-Agent header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		RequestRetry:   0,
		RequestTimeout: 5 * time.Second,
	}
	client := New(cfg)

	headers := map[string]string{
		"Authorization": "Bearer token",
	}

	_, err := client.Get(server.URL, headers)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
}

func TestClient_RetryOn500(t *testing.T) {
	var callCount int32

	// Server that fails first time, succeeds second time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	cfg := &config.Config{
		RequestRetry:   2,
		RequestTimeout: 5 * time.Second,
	}
	client := New(cfg)

	data, err := client.Get(server.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if string(data) != "success" {
		t.Errorf("Expected 'success', got '%s'", string(data))
	}

	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Expected 2 calls (1 fail + 1 retry), got %d", callCount)
	}
}

func TestClient_RetryOn429(t *testing.T) {
	var callCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	cfg := &config.Config{
		RequestRetry:   2,
		RequestTimeout: 5 * time.Second,
	}
	client := New(cfg)

	_, err := client.Get(server.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}

	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("Expected 2 calls (429 + retry), got %d", callCount)
	}
}

func TestClient_NoRetryOn404(t *testing.T) {
	var callCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cfg := &config.Config{
		RequestRetry:   2,
		RequestTimeout: 5 * time.Second,
	}
	client := New(cfg)

	_, err := client.Get(server.URL, nil)
	if err == nil {
		t.Fatal("Expected error for 404")
	}

	// Should not retry on 404
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected 1 call (no retry on 404), got %d", callCount)
	}
}

func TestClient_MaxRetries(t *testing.T) {
	var callCount int32

	// Server always returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{
		RequestRetry:   2,
		RequestTimeout: 5 * time.Second,
	}
	client := New(cfg)

	_, err := client.Get(server.URL, nil)
	if err == nil {
		t.Fatal("Expected error after max retries")
	}

	// Should be initial request + 2 retries = 3 total
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		code  int
		retry bool
	}{
		{200, false},
		{201, false},
		{301, false},
		{400, true}, // Retryable
		{403, false},
		{404, false},
		{408, true}, // Retryable
		{409, true}, // Retryable
		{425, true}, // Retryable
		{429, true}, // Retryable
		{500, true}, // Retryable
		{502, true}, // Retryable
		{503, true}, // Retryable
	}

	for _, test := range tests {
		result := shouldRetry(test.code)
		if result != test.retry {
			t.Errorf("shouldRetry(%d) = %v, want %v", test.code, result, test.retry)
		}
	}
}

func TestClient_CustomUserAgent(t *testing.T) {
	customUA := "CustomBot/1.0"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != customUA {
			t.Errorf("Expected User-Agent '%s', got '%s'", customUA, r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &config.Config{
		UserAgent:      customUA,
		RequestRetry:   0,
		RequestTimeout: 5 * time.Second,
	}
	client := New(cfg)

	_, err := client.Get(server.URL, nil)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
}
