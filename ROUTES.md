# Route System Documentation

This document describes how routes are organized and registered in GRSS.

## Overview

GRSS uses an **automatic route discovery system** that scans the `routes/` directory for namespaces and automatically imports them. This eliminates the need to manually import each route namespace in `main.go`.

### Key Benefits

- **Scalable**: Add new namespaces without touching `main.go`
- **Automatic**: Routes are discovered and registered via `init()` functions
- **Type-Safe**: Full compile-time type checking
- **Simple**: Just create a new directory and Go files

## Architecture

### 1. Namespace Organization

Routes are organized by namespace (subdirectories in `routes/`):

```
routes/
├── github/          # GitHub namespace
│   ├── namespace.go
│   ├── issues.go
│   └── init.go
├── example/         # Example namespace
│   ├── namespace.go
│   ├── hello.go
│   └── init.go
└── routes.go        # Auto-generated import file
```

Each namespace becomes a URL prefix: `/github/*`, `/example/*`, etc.

### 2. Route Discovery Process

```
┌─────────────────┐
│  make build      │
└────────┬─────────┘
         │
         ▼
┌─────────────────────────────┐
│  scripts/generate-routes.go │  ← Scans routes/ directory
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  routes/routes.go            │  ← Generated with imports
│  import (                    │
│    _ "routes/github"         │
│    _ "routes/example"        │
│  )                           │
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  init() functions execute   │  ← Routes register themselves
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│  Routes available at:       │
│  /github/issue/:user/:repo  │
│  /example/hello             │
└─────────────────────────────┘
```

### 3. Automatic Generation

The `scripts/generate-routes.go` tool:

1. Scans `routes/` for subdirectories
2. Skips `registry/` and hidden directories
3. Checks for `.go` files (indicating a valid namespace)
4. Generates `routes/routes.go` with all imports
5. Runs automatically as part of `make build`

## Creating a New Route Namespace

### Step 1: Create Directory Structure

Create a new directory in `routes/`:

```bash
mkdir routes/myservice
```

### Step 2: Define the Namespace

Create `routes/myservice/namespace.go`:

```go
package myservice

import "github.com/jean-jacket/grss/routes/registry"

var Namespace = &registry.Namespace{
    Name:        "My Service",
    URL:         "https://myservice.com",
    Description: "RSS feeds for My Service",
    Lang:        "en",
    Categories:  []string{"social-media"},
}
```

### Step 3: Create a Route Handler

Create `routes/myservice/latest.go`:

```go
package myservice

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jean-jacket/grss/feed"
    "github.com/jean-jacket/grss/routes/registry"
)

var LatestRoute = registry.Route{
    Path:        "/latest",
    Name:        "Latest Posts",
    Maintainers: []string{"yourusername"},
    Example:     "/myservice/latest",
    Description: "Get latest posts from My Service",
    Handler:     latestHandler,
}

func latestHandler(c *gin.Context) (*feed.Data, error) {
    // Fetch data from API
    // ...

    return &feed.Data{
        Title:       "My Service - Latest",
        Link:        "https://myservice.com",
        Description: "Latest posts from My Service",
        Item: []feed.Item{
            {
                Title:       "Example Post",
                Link:        "https://myservice.com/post/1",
                Description: "Post content",
                PubDate:     time.Now(),
            },
        },
    }, nil
}
```

### Step 4: Register Routes

Create `routes/myservice/init.go`:

```go
package myservice

import "github.com/jean-jacket/grss/routes/registry"

func init() {
    registry.RegisterNamespace("myservice", Namespace)
    registry.RegisterRoute("myservice", LatestRoute)
}
```

### Step 5: Build and Test

```bash
# Routes are automatically discovered
make build

# Test your route
curl http://localhost:1200/myservice/latest
```

**That's it!** No changes to `main.go` or any other files needed.

## Route Parameters

Routes support Gin's path parameters:

```go
var UserRoute = registry.Route{
    Path: "/user/:username",
    Handler: func(c *gin.Context) (*feed.Data, error) {
        username := c.Param("username")
        // ...
    },
}
```

**Optional parameters:**

```go
Path: "/posts/:category?"  // category is optional
```

**Query parameters:**

```go
func handler(c *gin.Context) (*feed.Data, error) {
    limit := c.DefaultQuery("limit", "10")
    filter := c.Query("filter")
    // ...
}
```

## Route Metadata

### Features

Declare route capabilities and requirements:

```go
var Route = registry.Route{
    Path: "/data",
    Features: &registry.Features{
        RequireConfig: []registry.ConfigRequirement{
            {Name: "API_KEY", Optional: false},
            {Name: "API_SECRET", Optional: true},
        },
        RequirePuppeteer: false,
        AntiCrawler:      false,
        SupportBT:        false,
        SupportScihub:    false,
    },
    Handler: handler,
}
```

### Parameters

Document route parameters:

```go
var Route = registry.Route{
    Path: "/search/:query/:page?",
    Parameters: map[string]interface{}{
        "query": "Search query string",
        "page": map[string]interface{}{
            "description": "Page number",
            "default":     "1",
        },
    },
    Handler: handler,
}
```

### Categories

Categorize routes for organization:

```go
var Route = registry.Route{
    Path:       "/feed",
    Categories: []string{"social-media", "news"},
    Handler:    handler,
}
```

## Development Workflow

### Manual Generation

If you need to regenerate routes without building:

```bash
make generate
```

Or directly:

```bash
go run scripts/generate-routes.go
```

### Using go generate

The `routes/routes.go` file includes a `//go:generate` directive:

```bash
cd routes
go generate
```

### Adding to Git

The generated `routes/routes.go` file **should be committed** to git, so that:
- CI/CD pipelines don't need to run the generator
- Other developers see what routes are registered
- Builds are reproducible

## Advanced Features

### Multiple Routes per File

You can define multiple routes in a single file:

```go
package myservice

var (
    LatestRoute = registry.Route{
        Path: "/latest",
        Handler: latestHandler,
    }

    TrendingRoute = registry.Route{
        Path: "/trending",
        Handler: trendingHandler,
    }
)

func init() {
    registry.RegisterNamespace("myservice", Namespace)
    registry.RegisterRoute("myservice", LatestRoute)
    registry.RegisterRoute("myservice", TrendingRoute)
}
```

### Shared Utilities

Create helper functions within your namespace:

```go
package myservice

import "github.com/jean-jacket/grss/client"

// Shared HTTP client for this namespace
var httpClient = client.New(config.C)

func fetchAPI(endpoint string) ([]byte, error) {
    return httpClient.Get("https://api.myservice.com"+endpoint, nil)
}
```

### Sub-routes

Use path nesting:

```go
var Routes = []registry.Route{
    {
        Path: "/user/:id",
        Handler: userHandler,
    },
    {
        Path: "/user/:id/posts",
        Handler: userPostsHandler,
    },
    {
        Path: "/user/:id/followers",
        Handler: userFollowersHandler,
    },
}
```

## Testing Routes

Create route-specific tests:

```go
// routes/myservice/latest_test.go
package myservice

import (
    "testing"

    "github.com/gin-gonic/gin"
)

func TestLatestHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    c, _ := gin.CreateTestContext(nil)

    data, err := latestHandler(c)
    if err != nil {
        t.Fatalf("Handler failed: %v", err)
    }

    if data.Title == "" {
        t.Error("Expected title to be set")
    }
}
```

## Troubleshooting

### Route not appearing

1. Check that your namespace directory has `.go` files
2. Verify `init()` function registers the route
3. Run `make generate` to regenerate `routes/routes.go`
4. Check the generated file includes your namespace

### Build fails after adding route

1. Ensure all imports are correct
2. Run `go mod tidy` to add missing dependencies
3. Check for syntax errors in your route files

### Route returns 404

1. Verify route is registered in `init()`
2. Check the path matches your request
3. Check namespace name matches directory name
4. Look for path parameter syntax errors

## Migration from Manual Imports

If you have manual imports in `main.go`:

```go
// OLD - Manual imports
import (
    _ "github.com/jean-jacket/grss/routes/github"
    _ "github.com/jean-jacket/grss/routes/twitter"
    _ "github.com/jean-jacket/grss/routes/reddit"
)
```

Replace with:

```go
// NEW - Automatic discovery
import (
    _ "github.com/jean-jacket/grss/routes"
)
```

Then run `make generate` to create the auto-import file.

## Summary

The route system provides:

✅ **Automatic Discovery** - No manual imports in `main.go`
✅ **Scalability** - Add hundreds of routes easily
✅ **Type Safety** - Full compile-time checking
✅ **Simplicity** - Just create files in `routes/`
✅ **Code Generation** - Runs automatically on build
✅ **Namespace Organization** - Clean URL structure

To add a new route:
1. Create directory in `routes/`
2. Add `namespace.go`, route files, and `init.go`
3. Run `make build`
4. Your route is live!
