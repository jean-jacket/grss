package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	// Application Settings
	DisallowRobot bool
	NodeName      string
	Debug         bool

	// Network Configuration
	Connect struct {
		Port            int
		ListenInaddrAny bool
	}

	RequestRetry   int
	RequestTimeout time.Duration
	UserAgent      string
	AllowOrigin    string

	// Cache Configuration
	Cache struct {
		Type          string // "memory", "redis", or "" (disabled)
		RouteExpire   time.Duration
		ContentExpire time.Duration
		MemoryMax     int
	}

	// Redis Configuration
	Redis struct {
		URL string
	}

	// Proxy Configuration
	Proxy struct {
		URI      string
		URIs     []string
		Strategy string // "all" or "on_retry"
		URLRegex string
	}

	// Access Control
	AccessKey string

	// Logging
	Logger struct {
		Level string
	}

	// Feed Features
	Hotlink struct {
		Template string
	}
	TitleLengthLimit int
	FilterRegexEngine string // "re2" or "regexp"

	// OpenAI Configuration
	OpenAI struct {
		APIKey string
		Model  string
	}

	// YouTube Configuration
	YouTube struct {
		Key          string // Comma-separated API keys for rotation
		ClientID     string
		ClientSecret string
		RefreshToken string
	}

	// Monitoring
	Sentry struct {
		DSN string
	}
}

var C *Config

// Load initializes configuration from environment variables
func Load() *Config {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	C = &Config{}

	// Application Settings
	C.DisallowRobot = viper.GetBool("DISALLOW_ROBOT")
	C.NodeName = viper.GetString("NODE_NAME")
	C.Debug = viper.GetBool("DEBUG_INFO")

	// Network Configuration
	C.Connect.Port = viper.GetInt("PORT")
	C.Connect.ListenInaddrAny = viper.GetBool("LISTEN_INADDR_ANY")
	C.RequestRetry = viper.GetInt("REQUEST_RETRY")
	C.RequestTimeout = time.Duration(viper.GetInt("REQUEST_TIMEOUT")) * time.Millisecond
	C.UserAgent = viper.GetString("UA")
	C.AllowOrigin = viper.GetString("ALLOW_ORIGIN")

	// Cache Configuration
	C.Cache.Type = viper.GetString("CACHE_TYPE")
	C.Cache.RouteExpire = time.Duration(viper.GetInt("CACHE_ROUTE_EXPIRE")) * time.Second
	C.Cache.ContentExpire = time.Duration(viper.GetInt("CACHE_CONTENT_EXPIRE")) * time.Second
	C.Cache.MemoryMax = viper.GetInt("MEMORY_MAX")

	// Redis Configuration
	C.Redis.URL = viper.GetString("REDIS_URL")

	// Proxy Configuration
	C.Proxy.URI = viper.GetString("PROXY_URI")
	proxyURIs := viper.GetString("PROXY_URIS")
	if proxyURIs != "" {
		C.Proxy.URIs = strings.Split(proxyURIs, ",")
	}
	C.Proxy.Strategy = viper.GetString("PROXY_STRATEGY")
	C.Proxy.URLRegex = viper.GetString("PROXY_URL_REGEX")

	// Access Control
	C.AccessKey = viper.GetString("ACCESS_KEY")

	// Logging
	C.Logger.Level = viper.GetString("LOGGER_LEVEL")

	// Feed Features
	C.Hotlink.Template = viper.GetString("HOTLINK_TEMPLATE")
	C.TitleLengthLimit = viper.GetInt("TITLE_LENGTH_LIMIT")
	C.FilterRegexEngine = viper.GetString("FILTER_REGEX_ENGINE")

	// OpenAI Configuration
	C.OpenAI.APIKey = viper.GetString("OPENAI_API_KEY")
	C.OpenAI.Model = viper.GetString("OPENAI_MODEL")

	// YouTube Configuration
	C.YouTube.Key = viper.GetString("YOUTUBE_KEY")
	C.YouTube.ClientID = viper.GetString("YOUTUBE_CLIENT_ID")
	C.YouTube.ClientSecret = viper.GetString("YOUTUBE_CLIENT_SECRET")
	C.YouTube.RefreshToken = viper.GetString("YOUTUBE_REFRESH_TOKEN")

	// Monitoring
	C.Sentry.DSN = viper.GetString("SENTRY_DSN")

	log.Printf("Configuration loaded: Port=%d, Cache=%s, Debug=%v", C.Connect.Port, C.Cache.Type, C.Debug)

	return C
}

func setDefaults() {
	// Application defaults
	viper.SetDefault("DISALLOW_ROBOT", false)
	viper.SetDefault("NODE_NAME", "")
	viper.SetDefault("DEBUG_INFO", false)

	// Network defaults
	viper.SetDefault("PORT", 1200)
	viper.SetDefault("LISTEN_INADDR_ANY", true)
	viper.SetDefault("REQUEST_RETRY", 2)
	viper.SetDefault("REQUEST_TIMEOUT", 30000)
	viper.SetDefault("UA", "")
	viper.SetDefault("ALLOW_ORIGIN", "*")

	// Cache defaults
	viper.SetDefault("CACHE_TYPE", "memory")
	viper.SetDefault("CACHE_ROUTE_EXPIRE", 300)
	viper.SetDefault("CACHE_CONTENT_EXPIRE", 3600)
	viper.SetDefault("MEMORY_MAX", 256)

	// Redis defaults
	viper.SetDefault("REDIS_URL", "")

	// Proxy defaults
	viper.SetDefault("PROXY_URI", "")
	viper.SetDefault("PROXY_URIS", "")
	viper.SetDefault("PROXY_STRATEGY", "all")
	viper.SetDefault("PROXY_URL_REGEX", ".*")

	// Access control defaults
	viper.SetDefault("ACCESS_KEY", "")

	// Logging defaults
	viper.SetDefault("LOGGER_LEVEL", "info")

	// Feed feature defaults
	viper.SetDefault("HOTLINK_TEMPLATE", "")
	viper.SetDefault("TITLE_LENGTH_LIMIT", 150)
	viper.SetDefault("FILTER_REGEX_ENGINE", "re2")

	// OpenAI defaults
	viper.SetDefault("OPENAI_API_KEY", "")
	viper.SetDefault("OPENAI_MODEL", "")

	// YouTube defaults
	viper.SetDefault("YOUTUBE_KEY", "")
	viper.SetDefault("YOUTUBE_CLIENT_ID", "")
	viper.SetDefault("YOUTUBE_CLIENT_SECRET", "")
	viper.SetDefault("YOUTUBE_REFRESH_TOKEN", "")

	// Monitoring defaults
	viper.SetDefault("SENTRY_DSN", "")
}
