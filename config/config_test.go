package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any existing env vars
	os.Clearenv()

	cfg := Load()

	// Test defaults
	if cfg.Connect.Port != 1200 {
		t.Errorf("Expected default port 1200, got %d", cfg.Connect.Port)
	}
	if cfg.Connect.ListenInaddrAny != true {
		t.Error("Expected ListenInaddrAny to be true by default")
	}
	if cfg.Cache.Type != "memory" {
		t.Errorf("Expected default cache type 'memory', got '%s'", cfg.Cache.Type)
	}
	if cfg.Cache.RouteExpire != 300*time.Second {
		t.Errorf("Expected default route expire 300s, got %v", cfg.Cache.RouteExpire)
	}
	if cfg.Cache.ContentExpire != 3600*time.Second {
		t.Errorf("Expected default content expire 3600s, got %v", cfg.Cache.ContentExpire)
	}
	if cfg.Cache.MemoryMax != 256 {
		t.Errorf("Expected default memory max 256, got %d", cfg.Cache.MemoryMax)
	}
	if cfg.RequestRetry != 2 {
		t.Errorf("Expected default request retry 2, got %d", cfg.RequestRetry)
	}
	if cfg.RequestTimeout != 30000*time.Millisecond {
		t.Errorf("Expected default request timeout 30s, got %v", cfg.RequestTimeout)
	}
	if cfg.AllowOrigin != "*" {
		t.Errorf("Expected default allow origin '*', got '%s'", cfg.AllowOrigin)
	}
	if cfg.Logger.Level != "info" {
		t.Errorf("Expected default logger level 'info', got '%s'", cfg.Logger.Level)
	}
	if cfg.TitleLengthLimit != 150 {
		t.Errorf("Expected default title length limit 150, got %d", cfg.TitleLengthLimit)
	}
	if cfg.FilterRegexEngine != "re2" {
		t.Errorf("Expected default filter regex engine 're2', got '%s'", cfg.FilterRegexEngine)
	}
}

func TestLoad_CustomPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("PORT", "3000")

	cfg := Load()

	if cfg.Connect.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", cfg.Connect.Port)
	}
}

func TestLoad_CustomCache(t *testing.T) {
	os.Clearenv()
	os.Setenv("CACHE_TYPE", "redis")
	os.Setenv("REDIS_URL", "redis://localhost:6379/1")
	os.Setenv("CACHE_ROUTE_EXPIRE", "600")
	os.Setenv("CACHE_CONTENT_EXPIRE", "7200")

	cfg := Load()

	if cfg.Cache.Type != "redis" {
		t.Errorf("Expected cache type 'redis', got '%s'", cfg.Cache.Type)
	}
	if cfg.Redis.URL != "redis://localhost:6379/1" {
		t.Errorf("Expected redis URL 'redis://localhost:6379/1', got '%s'", cfg.Redis.URL)
	}
	if cfg.Cache.RouteExpire != 600*time.Second {
		t.Errorf("Expected route expire 600s, got %v", cfg.Cache.RouteExpire)
	}
	if cfg.Cache.ContentExpire != 7200*time.Second {
		t.Errorf("Expected content expire 7200s, got %v", cfg.Cache.ContentExpire)
	}
}

func TestLoad_DebugMode(t *testing.T) {
	os.Clearenv()
	os.Setenv("DEBUG_INFO", "true")

	cfg := Load()

	if !cfg.Debug {
		t.Error("Expected debug to be true")
	}
}

func TestLoad_AccessKey(t *testing.T) {
	os.Clearenv()
	os.Setenv("ACCESS_KEY", "secret123")

	cfg := Load()

	if cfg.AccessKey != "secret123" {
		t.Errorf("Expected access key 'secret123', got '%s'", cfg.AccessKey)
	}
}

func TestLoad_Proxy(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_URI", "http://proxy:8080")
	os.Setenv("PROXY_STRATEGY", "on_retry")

	cfg := Load()

	if cfg.Proxy.URI != "http://proxy:8080" {
		t.Errorf("Expected proxy URI 'http://proxy:8080', got '%s'", cfg.Proxy.URI)
	}
	if cfg.Proxy.Strategy != "on_retry" {
		t.Errorf("Expected proxy strategy 'on_retry', got '%s'", cfg.Proxy.Strategy)
	}
}

func TestLoad_MultipleProxies(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_URIS", "http://proxy1:8080,http://proxy2:8080,http://proxy3:8080")

	cfg := Load()

	if len(cfg.Proxy.URIs) != 3 {
		t.Errorf("Expected 3 proxy URIs, got %d", len(cfg.Proxy.URIs))
	}
	if cfg.Proxy.URIs[0] != "http://proxy1:8080" {
		t.Errorf("Expected first proxy 'http://proxy1:8080', got '%s'", cfg.Proxy.URIs[0])
	}
}

func TestLoad_RequestSettings(t *testing.T) {
	os.Clearenv()
	os.Setenv("REQUEST_RETRY", "5")
	os.Setenv("REQUEST_TIMEOUT", "60000")
	os.Setenv("UA", "CustomBot/1.0")

	cfg := Load()

	if cfg.RequestRetry != 5 {
		t.Errorf("Expected request retry 5, got %d", cfg.RequestRetry)
	}
	if cfg.RequestTimeout != 60000*time.Millisecond {
		t.Errorf("Expected request timeout 60s, got %v", cfg.RequestTimeout)
	}
	if cfg.UserAgent != "CustomBot/1.0" {
		t.Errorf("Expected UA 'CustomBot/1.0', got '%s'", cfg.UserAgent)
	}
}

func TestLoad_FeedSettings(t *testing.T) {
	os.Clearenv()
	os.Setenv("TITLE_LENGTH_LIMIT", "200")
	os.Setenv("FILTER_REGEX_ENGINE", "regexp")
	os.Setenv("HOTLINK_TEMPLATE", "https://proxy.example.com/{url}")

	cfg := Load()

	if cfg.TitleLengthLimit != 200 {
		t.Errorf("Expected title length limit 200, got %d", cfg.TitleLengthLimit)
	}
	if cfg.FilterRegexEngine != "regexp" {
		t.Errorf("Expected filter regex engine 'regexp', got '%s'", cfg.FilterRegexEngine)
	}
	if cfg.Hotlink.Template != "https://proxy.example.com/{url}" {
		t.Errorf("Expected hotlink template 'https://proxy.example.com/{url}', got '%s'", cfg.Hotlink.Template)
	}
}

func TestLoad_DisallowRobot(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISALLOW_ROBOT", "true")

	cfg := Load()

	if !cfg.DisallowRobot {
		t.Error("Expected disallow robot to be true")
	}
}

func TestLoad_ListenAddress(t *testing.T) {
	os.Clearenv()
	os.Setenv("LISTEN_INADDR_ANY", "false")

	cfg := Load()

	if cfg.Connect.ListenInaddrAny {
		t.Error("Expected listen inaddr any to be false")
	}
}
