package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/cache"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/middleware"
	"github.com/jean-jacket/grss/routes/registry"

	// Import routes package to auto-register all route namespaces
	_ "github.com/jean-jacket/grss/routes"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Initialize cache
	var cacheInstance cache.Cache
	switch cfg.Cache.Type {
	case "memory":
		cacheInstance = cache.NewMemoryCache(cfg.Cache.MemoryMax)
		log.Printf("Using memory cache with max %d items", cfg.Cache.MemoryMax)
	case "redis":
		redisCache, err := cache.NewRedisCache(cfg.Redis.URL)
		if err != nil {
			log.Fatalf("Failed to connect to Redis: %v", err)
		}
		cacheInstance = redisCache
		log.Printf("Using Redis cache at %s", cfg.Redis.URL)
	default:
		log.Printf("Cache disabled")
	}

	// Middleware chain (order matters!)
	router.Use(middleware.Logger())
	router.Use(middleware.AccessControl())
	router.Use(middleware.Header())
	router.Use(middleware.Parameter())
	if cacheInstance != nil {
		router.Use(middleware.Cache(cacheInstance))
	}
	router.Use(middleware.Template())

	// Built-in routes
	router.GET("/", homeHandler)
	router.GET("/healthz", healthzHandler)
	router.GET("/robots.txt", robotsHandler)

	// Mount all registered routes
	registry.MountRoutes(router)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Connect.Port)
	if !cfg.Connect.ListenInaddrAny {
		addr = fmt.Sprintf("127.0.0.1:%d", cfg.Connect.Port)
	}

	log.Printf("Starting GRSS on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// homeHandler serves the homepage
func homeHandler(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>GRSS</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        p { color: #666; line-height: 1.6; }
        a { color: #0366d6; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .example { background: #f6f8fa; padding: 10px; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <h1>🎉 GRSS</h1>
    <p>Welcome to GRSS! A Go-based RSS feed aggregation service inspired by RSSHub.</p>

    <h2>Example Routes</h2>
    <div class="example">
        <strong>GitHub Issues:</strong><br>
        <a href="/github/issue/golang/go">/github/issue/golang/go</a>
    </div>

    <h2>Query Parameters</h2>
    <ul>
        <li><code>format</code>: Output format (rss, atom, json) - default: rss</li>
        <li><code>limit</code>: Limit number of items</li>
        <li><code>filter</code>: Filter items by regex</li>
        <li><code>sorted</code>: Sort by date (asc, desc)</li>
    </ul>

    <h2>Links</h2>
    <ul>
        <li><a href="/healthz">Health Check</a></li>
        <li><a href="https://github.com/jean-jacket/grss">GitHub Repository</a></li>
    </ul>
</body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, html)
}

// healthzHandler serves the health check endpoint
func healthzHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

// robotsHandler serves the robots.txt file
func robotsHandler(c *gin.Context) {
	robots := "User-agent: *\n"
	if config.C.DisallowRobot {
		robots += "Disallow: /\n"
	} else {
		robots += "Allow: /\n"
	}
	c.Header("Content-Type", "text/plain")
	c.String(200, robots)
}
