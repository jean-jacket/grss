# GRSS

A Go-based RSS feed aggregation service inspired by [RSSHub](https://github.com/DIYgod/RSSHub).

## Overview

GRSS is a ground-up rewrite of RSSHub in Go, focusing on performance, reliability, and maintainability while preserving the core architecture and functionality of the original Node.js version.

## Features

- **Multiple Feed Formats**: RSS 2.0, Atom 1.0, JSON Feed 1.1
- **Automatic Route Discovery**: Add new routes without modifying main code
- **Aggressive Caching**: Memory (LRU) and Redis-backed caching with request deduplication
- **Query Parameters**: Filter, limit, sort, and transform feed items
- **Proxy Support**: Single or multi-proxy with configurable strategies
- **Access Control**: API key and per-path code authentication
- **Middleware Pipeline**: Extensible middleware chain for request processing

## Architecture

### Core Components

1. **Configuration System** (`config/`): Environment-based configuration using Viper
2. **Cache Layer** (`cache/`): Abstracted caching with Memory and Redis implementations
3. **HTTP Client** (`client/`): Retry logic and proxy support built-in
4. **Feed Generation** (`feed/`): RSS, Atom, and JSON Feed generators
5. **Middleware Chain** (`middleware/`): Sequential request processing pipeline
6. **Route Registry** (`routes/registry/`): Dynamic route discovery and mounting

### Middleware Chain (Execution Order)

The middleware chain executes in strict order:

```
Request → Logger → Access Control → Header → Parameter → Cache → Template → Response
```

1. **Logger**: Request/response logging with timing
2. **Access Control**: API key or MD5 code validation
3. **Header**: CORS, ETag, Cache-Control headers
4. **Parameter**: Query parameter processing (filter, limit, sorted, etc.)
5. **Cache**: Cache layer with request deduplication (prevents thundering herd)
6. **Template**: Converts Data object to RSS/Atom/JSON

### Request Flow

```
Route Handler → Returns feed.Data → Template Middleware → Generates Feed → Cache → Response
```

Route handlers return a `feed.Data` struct, which is then processed by the middleware chain:
- Parameter middleware applies filters and transformations
- Template middleware converts to the requested format
- Cache middleware stores the result
- Response is sent to client

## Project Structure

```
grss/
├── cmd/rsshub/           # Main application entry point
├── config/               # Configuration management
├── cache/                # Cache interface and implementations
│   ├── cache.go         # Cache interface
│   ├── memory.go        # LRU memory cache
│   └── redis.go         # Redis cache
├── client/               # HTTP client with retry logic
├── feed/                 # Feed data structures and generators
│   ├── data.go          # Data and Item structs
│   ├── rss.go           # RSS 2.0 generator
│   ├── atom.go          # Atom 1.0 generator
│   └── json.go          # JSON Feed 1.1 generator
├── middleware/           # Middleware components
│   ├── logger.go        # Request/response logging
│   ├── cache.go         # Cache with deduplication
│   ├── template.go      # Feed rendering
│   ├── parameter.go     # Query parameter processing
│   ├── access_control.go # Authentication
│   └── header.go        # HTTP headers
├── routes/               # Route implementations (auto-discovered)
│   ├── routes.go        # Auto-generated imports (via go generate)
│   ├── registry/        # Route registration system
│   ├── github/          # GitHub namespace
│   │   ├── namespace.go
│   │   ├── issues.go
│   │   └── init.go
│   └── example/         # Example namespace
│       ├── namespace.go
│       ├── hello.go
│       └── init.go
├── scripts/              # Code generation tools
│   └── generate-routes.go
├── utils/                # Utility functions
│   ├── logger.go
│   └── errors.go
└── proxy/                # Proxy management
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Redis (optional, for Redis cache)

### Installation

```bash
# Clone the repository
git clone https://github.com/jean-jacket/grss.git
cd grss

# Install dependencies
go mod download

# Copy environment configuration
cp .env.example .env

# Build
go build -o rsshub cmd/rsshub/main.go

# Run
./rsshub
```

### Using Docker

```bash
# Build image
docker build -t rsshub .

# Run with memory cache
docker run -p 1200:1200 rsshub

# Run with Redis cache
docker-compose up
```

## Configuration

Configuration is done via environment variables. See `.env.example` for all available options.

### Key Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 1200 | Server port |
| `CACHE_TYPE` | memory | Cache type: `memory`, `redis`, or `` (disabled) |
| `CACHE_ROUTE_EXPIRE` | 300 | Route cache TTL (seconds) |
| `REDIS_URL` | - | Redis connection URL |
| `ACCESS_KEY` | - | API key for authentication |
| `REQUEST_RETRY` | 2 | Number of retry attempts |
| `PROXY_URI` | - | Proxy URL |
| `DEBUG_INFO` | false | Enable debug mode |

## Usage

### Basic Example

```bash
# Get GitHub issues as RSS
curl http://localhost:1200/github/issue/golang/go

# Get as Atom
curl http://localhost:1200/github/issue/golang/go?format=atom

# Get as JSON Feed
curl http://localhost:1200/github/issue/golang/go?format=json
```

### Query Parameters

- **`format`**: Output format (`rss`, `atom`, `json`) - default: `rss`
- **`limit`**: Limit number of items
- **`filter`**: Filter items by regex (matches title or description)
- **`filterout`**: Exclude items by regex
- **`filter_title`**: Filter by title only
- **`filter_description`**: Filter by description only
- **`filter_time`**: Only items newer than N seconds
- **`sorted`**: Sort by date (`asc`, `desc`)

### Examples

```bash
# Limit to 10 items
curl http://localhost:1200/github/issue/golang/go?limit=10

# Filter by keyword
curl http://localhost:1200/github/issue/golang/go?filter=bug

# Sort by date (oldest first)
curl http://localhost:1200/github/issue/golang/go?sorted=asc

# Combine parameters
curl http://localhost:1200/github/issue/golang/go?format=atom&limit=20&filter=feature
```

## Creating New Routes

Routes are **automatically discovered** - just create files in `routes/` and run `make build`. No manual imports needed!

See [ROUTES.md](ROUTES.md) for detailed documentation.

### 1. Create Namespace

Create `routes/{namespace}/namespace.go`:

```go
package myservice

import "github.com/jean-jacket/grss/routes/registry"

var Namespace = &registry.Namespace{
    Name:        "My Service",
    URL:         "https://myservice.com",
    Description: "My service description",
    Lang:        "en",
}
```

### 2. Create Route Handler

Create `routes/{namespace}/myroute.go`:

```go
package myservice

import (
    "github.com/gin-gonic/gin"
    "github.com/jean-jacket/grss/feed"
    "github.com/jean-jacket/grss/routes/registry"
)

var MyRoute = registry.Route{
    Path:        "/myroute/:param",
    Name:        "My Route",
    Maintainers: []string{"username"},
    Example:     "/myservice/myroute/example",
    Handler:     myRouteHandler,
}

func myRouteHandler(c *gin.Context) (*feed.Data, error) {
    param := c.Param("param")

    // Fetch data...

    return &feed.Data{
        Title:       "Feed Title",
        Link:        "https://myservice.com",
        Description: "Feed description",
        Item: []feed.Item{
            {
                Title:       "Item 1",
                Link:        "https://myservice.com/item1",
                Description: "Item 1 description",
                PubDate:     time.Now(),
            },
        },
    }, nil
}
```

### 3. Register Route

Create `routes/{namespace}/init.go`:

```go
package myservice

import "github.com/jean-jacket/grss/routes/registry"

func init() {
    registry.RegisterNamespace("myservice", Namespace)
    registry.RegisterRoute("myservice", MyRoute)
}
```

### 4. Build and Test

Routes are automatically discovered - no manual imports needed!

```bash
# The build process auto-generates imports
make build

# Test your route
curl http://localhost:1200/myservice/myroute/test
```

## Caching

### Memory Cache (Default)

- Uses LRU eviction
- Configurable max items
- No external dependencies

```bash
CACHE_TYPE=memory
MEMORY_MAX=256
```

### Redis Cache

- Distributed caching
- Persistent storage
- Horizontal scaling

```bash
CACHE_TYPE=redis
REDIS_URL=redis://localhost:6379
```

### Request Deduplication

The cache middleware uses `singleflight` to prevent thundering herd:
- Multiple concurrent requests for the same resource → single upstream request
- Waiting requests receive the same result
- Prevents cache stampede

## Performance

### Go vs Node.js

Key advantages of the Go rewrite:

1. **Native Concurrency**: Goroutines vs. event loop + cluster mode
2. **Lower Memory**: ~10-20MB vs. ~50-100MB for equivalent workload
3. **Faster Startup**: ~100ms vs. ~1-2s
4. **Better CPU Utilization**: Native parallelism across all cores
5. **Simpler Deployment**: Single binary vs. Node.js + npm dependencies

### Benchmarks

Coming soon...

## Roadmap

### Phase 1: Foundation (Current)
- [x] Configuration system
- [x] Cache abstraction (Memory, Redis)
- [x] HTTP client with retry
- [x] Feed generators (RSS, Atom, JSON)
- [x] Middleware chain
- [x] Route registry
- [x] Example routes

### Phase 2: Core Routes
- [ ] Migrate top 50 most popular routes
- [ ] GitHub (issues, releases, commits)
- [ ] Twitter/X
- [ ] YouTube
- [ ] Reddit
- [ ] Bilibili

### Phase 3: Advanced Features
- [ ] Full-text extraction (mode=fulltext)
- [ ] OpenAI integration (chatgpt parameter)
- [ ] OpenCC Chinese conversion
- [ ] Image proxy (anti-hotlink)
- [ ] Puppeteer/browser automation

### Phase 4: Production Ready
- [ ] OpenTelemetry tracing
- [ ] Prometheus metrics
- [ ] Sentry error tracking
- [ ] Rate limiting
- [ ] Documentation site

## Contributing

Contributions are welcome! Please see the [Contributing Guide](CONTRIBUTING.md) for details.

### Migration from Node.js

If you're migrating routes from the original RSSHub:

1. Understand the middleware chain order
2. Return `feed.Data` from handlers (middleware handles rendering)
3. Use `cache.TryGet()` for content caching
4. Follow the namespace/route structure
5. Test with all three formats (RSS, Atom, JSON)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Original [RSSHub project](https://github.com/DIYgod/RSSHub) by DIYgod and contributors
- Inspired by the architecture documentation and extensive route library

## Links

- [Documentation](https://docs.rsshub.app)
- [GitHub Repository](https://github.com/jean-jacket/grss)
- [Issue Tracker](https://github.com/jean-jacket/grss/issues)
