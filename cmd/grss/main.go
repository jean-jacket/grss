package main

import (
	"flag"
	"fmt"
	"log"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/cache"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/middleware"
	"github.com/jean-jacket/grss/routes/registry"

	// Import routes package to auto-register all route namespaces
	_ "github.com/jean-jacket/grss/routes"
)

func main() {
	// Command-line flags
	testRoute := flag.String("test-route", "", "Test a route and print its output (e.g., '/github/issue/golang/go')")
	testLimit := flag.Int("test-limit", 5, "Number of items to display when testing a route")
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// If test-route flag is set, run route test and exit
	if *testRoute != "" {
		testRouteHandler(*testRoute, *testLimit)
		return
	}

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

// testRouteHandler tests a route and prints debug information
func testRouteHandler(routePath string, limit int) {
	// Ensure route starts with /
	if !strings.HasPrefix(routePath, "/") {
		routePath = "/" + routePath
	}

	fmt.Printf("🧪 Testing route: %s\n", routePath)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Get all registered routes
	allRoutes := registry.GetAllRoutes()

	// Find matching route by trying to match the path pattern
	var matchedRoute *registry.RouteInfo
	var params map[string]string

	for pattern, routeInfo := range allRoutes {
		if matched, extractedParams := matchRoutePath(pattern, routePath); matched {
			matchedRoute = &routeInfo
			params = extractedParams
			break
		}
	}

	if matchedRoute == nil {
		fmt.Printf("❌ Route not found: %s\n\n", routePath)
		fmt.Println("Available routes:")
		for path := range allRoutes {
			fmt.Printf("  - %s\n", path)
		}
		return
	}

	fmt.Printf("✓ Route found: %s\n", matchedRoute.Route.Name)
	if matchedRoute.Route.Description != "" {
		fmt.Printf("  Description: %s\n", matchedRoute.Route.Description)
	}
	fmt.Println()

	// Create a mock Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up the request with params
	c.Request = httptest.NewRequest("GET", routePath, nil)
	c.Params = make(gin.Params, 0, len(params))
	for key, value := range params {
		c.Params = append(c.Params, gin.Param{Key: key, Value: value})
	}

	// Parse query parameters if present
	if strings.Contains(routePath, "?") {
		parts := strings.SplitN(routePath, "?", 2)
		if len(parts) == 2 {
			queryParams, err := url.ParseQuery(parts[1])
			if err == nil {
				c.Request.URL.RawQuery = parts[1]
				for key, values := range queryParams {
					if len(values) > 0 {
						c.Request.Form = url.Values{}
						c.Request.Form[key] = values
					}
				}
			}
		}
	}

	// Execute the handler and measure time
	fmt.Println("⏱️  Executing handler...")
	startTime := time.Now()

	feedData, err := matchedRoute.Route.Handler(c)

	duration := time.Since(startTime)

	fmt.Println()
	fmt.Printf("✓ Execution time: %v\n", duration)
	fmt.Println()

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if feedData == nil {
		fmt.Println("❌ Handler returned nil feed data")
		return
	}

	// Print feed metadata
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("📊 %s\n", feedData.Title)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Link: %s\n", feedData.Link)
	if feedData.Description != "" {
		fmt.Printf("Description: %s\n", feedData.Description)
	}
	fmt.Printf("Items: %d\n", len(feedData.Item))
	fmt.Println()

	// Print items in table format
	displayCount := limit
	if displayCount > len(feedData.Item) {
		displayCount = len(feedData.Item)
	}

	if displayCount > 0 {
		// Print table header
		fmt.Printf("%-3s | %-45s | %-12s | %-12s\n", "#", "Title", "Author", "Date")
		fmt.Println(strings.Repeat("-", 80))

		for i := 0; i < displayCount; i++ {
			item := feedData.Item[i]

			// Truncate title if too long
			title := item.Title
			if len(title) > 45 {
				title = title[:42] + "..."
			}

			// Truncate author if too long
			author := item.Author
			if len(author) > 12 {
				author = author[:9] + "..."
			}
			if author == "" {
				author = "-"
			}

			// Format date
			dateStr := "-"
			if !item.PubDate.IsZero() {
				dateStr = item.PubDate.Format("2006-01-02")
			}

			fmt.Printf("%-3d | %-45s | %-12s | %-12s\n", i+1, title, author, dateStr)
		}

		if len(feedData.Item) > displayCount {
			fmt.Printf("\n... and %d more items\n", len(feedData.Item)-displayCount)
		}
	} else {
		fmt.Println("ℹ️  No items in feed")
	}

	fmt.Println()
	fmt.Println("✅ Test completed successfully")
	fmt.Println()
}

// matchRoutePath matches a route pattern against an actual path and extracts params
// For example: "/github/issue/:user/:repo" matches "/github/issue/golang/go"
func matchRoutePath(pattern, path string) (bool, map[string]string) {
	params := make(map[string]string)

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Remove query string from last part if present
	if len(pathParts) > 0 {
		lastPart := pathParts[len(pathParts)-1]
		if idx := strings.Index(lastPart, "?"); idx != -1 {
			pathParts[len(pathParts)-1] = lastPart[:idx]
		}
	}

	// Must have same number of parts
	if len(patternParts) != len(pathParts) {
		return false, nil
	}

	// Match each part
	for i := 0; i < len(patternParts); i++ {
		patternPart := patternParts[i]
		pathPart := pathParts[i]

		// Check if this is a parameter (starts with :)
		if strings.HasPrefix(patternPart, ":") {
			paramName := strings.TrimPrefix(patternPart, ":")
			params[paramName] = pathPart
		} else if strings.HasPrefix(patternPart, "*") {
			// Wildcard match - match anything
			paramName := strings.TrimPrefix(patternPart, "*")
			if paramName != "" {
				params[paramName] = pathPart
			}
		} else {
			// Exact match required
			if patternPart != pathPart {
				return false, nil
			}
		}
	}

	return true, params
}

// matchRouteWithRegex provides more sophisticated pattern matching
func matchRouteWithRegex(pattern, path string) (bool, map[string]string) {
	params := make(map[string]string)

	// Convert Gin route pattern to regex
	regexPattern := "^" + pattern + "$"

	// Replace :param with named capture group
	paramRegex := regexp.MustCompile(`:([a-zA-Z0-9_]+)`)
	matches := paramRegex.FindAllStringSubmatch(pattern, -1)

	for _, match := range matches {
		paramName := match[1]
		regexPattern = strings.Replace(regexPattern, ":"+paramName, "(?P<"+paramName+">[^/]+)", 1)
	}

	// Try to match
	re := regexp.MustCompile(regexPattern)
	if !re.MatchString(path) {
		return false, nil
	}

	// Extract parameters
	matchResult := re.FindStringSubmatch(path)
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(matchResult) {
			params[name] = matchResult[i]
		}
	}

	return true, params
}
