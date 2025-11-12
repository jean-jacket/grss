# GRSS 🌾

A high-performance RSS feed aggregator written in Go, inspired by [RSSHub](https://github.com/DIYgod/RSSHub).

## Available Routes

### Example
- `/example/hello` - Hello World feed (demo)

### GitHub
- `/github/issue/:user/:repo` - Repository issues
  - Example: `/github/issue/golang/go`
  - Query params: `state=open|closed|all`

### Apple
- `/apple/design` - Apple Developer Design updates
  - Example: `/apple/design`

### YouTube
- `/youtube/channel/:id` - Channel videos by channel ID
  - Example: `/youtube/channel/UCDwDMPOZfxVV0x_dz0eQ8KQ`
  - Params: `embed=true|false`, `filterShorts=true|false`
- `/youtube/user/:username` - Channel videos by username/handle
  - Example: `/youtube/user/@JFlaMusic`
  - Params: `embed=true|false`, `filterShorts=true|false`
- `/youtube/playlist/:id` - Playlist videos
  - Example: `/youtube/playlist/PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf`
  - Params: `embed=true|false`

## Build Instructions

```bash
# Clone repository
git clone https://github.com/jean-jacket/grss.git
cd grss

# Build
go build -o rsshub cmd/grss/main.go

# Run
./rsshub
```

## Adding New Routes

Use AI to help you create new routes! Here's a basic prompt:

```
Create a new RSS route for GRSS that fetches [data source].

The route should:
1. Create a namespace package in routes/[namespace]/
2. Define a namespace.go with Name, URL, Description, Lang
3. Create [route].go with Route definition including Path, Name, Example, Handler
4. Implement the handler function that returns *feed.Data
5. Register both in init.go

Follow the pattern in routes/github/issues.go for reference.
```

### Manual Steps

1. **Create namespace** at `routes/[namespace]/namespace.go`:
```go
package myservice

import "github.com/jean-jacket/grss/routes/registry"

var Namespace = &registry.Namespace{
    Name:        "My Service",
    URL:         "https://myservice.com",
    Description: "Description here",
    Lang:        "en",
}
```

2. **Create route handler** at `routes/[namespace]/myroute.go`:
```go
package myservice

import (
    "github.com/gin-gonic/gin"
    "github.com/jean-jacket/grss/feed"
    "github.com/jean-jacket/grss/routes/registry"
)

var MyRoute = registry.Route{
    Path:    "/myroute/:param",
    Name:    "My Route",
    Example: "/myservice/myroute/example",
    Handler: myRouteHandler,
}

func myRouteHandler(c *gin.Context) (*feed.Data, error) {
    // Fetch and return feed data
    return &feed.Data{
        Title: "Feed Title",
        Link:  "https://example.com",
        Item:  []feed.Item{...},
    }, nil
}
```

3. **Register route** at `routes/[namespace]/init.go`:
```go
package myservice

import "github.com/jean-jacket/grss/routes/registry"

func init() {
    registry.RegisterNamespace("myservice", Namespace)
    registry.RegisterRoute("myservice", MyRoute)
}
```

4. **Build**: Routes are auto-discovered during build - no manual imports needed!

## Query Parameters

All routes support:
- `format=rss|atom|json` - Output format (default: rss)
- `limit=N` - Limit items
- `filter=regex` - Filter by title/description
- `filterout=regex` - Exclude items
- `sorted=asc|desc` - Sort by date

## Comparison to RSSHub

GRSS is a ground-up rewrite of RSSHub in Go, focusing on performance and simplicity.

### Architecture
- **RSSHub**: Node.js with Express, cluster mode for multi-core
- **GRSS**: Go with Gin, native goroutines for concurrency

### Performance Benchmarks

Real-world performance comparison (tested on same machine):

| Metric | RSSHub | GRSS | Winner |
|--------|--------|------|--------|
| **Memory Usage** | 757 MB | 38 MB | GRSS (20x less) |
| **Response Time** | 5ms avg | 1ms avg | GRSS (5x faster) |
| **Load Test (50 req)** | 0.27s | 0.26s | Similar |
| **Startup Time** | 30-40s | 100ms | GRSS (300x faster) |
| **Binary Size** | 1.5 GB total | 23 MB | GRSS (65x smaller) |
| **Dependencies** | 1258 packages | 0 (single binary) | GRSS |

**Test Setup:**
- GRSS: Single Go binary with memory cache
- RSSHub: Node.js 22 with 1258 npm packages
- Both running simple routes under identical conditions

### Key Differences

| Feature | RSSHub | GRSS |
|---------|--------|------|
| Language | JavaScript (Node.js) | Go |
| Memory | ~750MB typical | ~38MB typical |
| Startup | ~30-40s | ~100ms |
| Binary | Requires Node.js + deps | Single static binary |
| Concurrency | Event loop + cluster | Native goroutines |
| Routes | 1000+ | Growing (6 routes) |

**GRSS Advantages:**
- Lower memory footprint
- Faster startup time
- Simple deployment (single binary)
- Native concurrency with goroutines
- Strong typing and compile-time safety

**RSSHub Advantages:**
- Mature ecosystem with 1000+ routes
- Large community and contributors
- More features (Puppeteer, OpenAI, etc.)
- Extensive documentation

### When to Use Each

**Use GRSS if:**
- You need specific routes and want minimal resource usage
- You prefer static binaries and simple deployment
- You're running on resource-constrained environments
- You want to contribute Go code

**Use RSSHub if:**
- You need access to hundreds of pre-built routes
- You need advanced features (browser automation, AI summaries)
- You prefer JavaScript/TypeScript development
- You want a battle-tested, production-ready solution

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by the amazing [RSSHub project](https://github.com/DIYgod/RSSHub) by DIYgod and contributors.
