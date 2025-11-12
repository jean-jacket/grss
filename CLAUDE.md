# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GRSS is a high-performance RSS feed aggregator written in Go, inspired by RSSHub. It uses Gin for HTTP routing and provides an extensible architecture for creating custom RSS routes.

## Build and Development Commands

```bash
# Build the application (runs code generation first)
make build

# Run the application
make run

# Run tests
go test -v ./...

# Run a specific test
go test -v ./path/to/package -run TestName

# Run tests with coverage
make test-coverage

# Format code
go fmt ./...

# Lint code (requires golangci-lint)
make lint

# Development with hot reload (requires air)
make dev

# Generate routes (auto-discovers route namespaces)
make generate
# or directly:
go run scripts/generate-routes.go
```

## Architecture

### Request Flow

1. **HTTP Request** → Gin Router
2. **Middleware Chain** (order matters):
   - `Logger()` - Request logging
   - `AccessControl()` - API key validation
   - `Header()` - Set response headers
   - `Parameter()` - Process query parameters (filter, limit, sorted)
   - `Cache()` - Handle caching with singleflight
   - `Template()` - Convert feed.Data to RSS/Atom/JSON
3. **Route Handler** → Returns `*feed.Data`
4. **Response** → Feed output (RSS/Atom/JSON)

### Core Components

**Routes (`routes/`)**
- Auto-registered via `init()` functions
- Pattern: `routes/{namespace}/` contains namespace definition, routes, and init
- Code generation: `scripts/generate-routes.go` scans for namespaces and generates `routes/routes.go`
- Import: `_ "github.com/jean-jacket/grss/routes"` in main triggers all route registrations

**Registry (`routes/registry/`)**
- Central route registration system
- `Route` struct defines path, handler, and metadata
- `RouteHandler` signature: `func(c *gin.Context) (*feed.Data, error)`
- `MountRoutes()` mounts all registered routes to Gin router

**Feed (`feed/`)**
- `feed.Data` is the universal return type for all route handlers
- Contains title, link, description, and slice of `feed.Item`
- Converted to RSS/Atom/JSON by middleware/template.go
- Support for media, enclosures, torrents, iTunes podcasts

**Middleware (`middleware/`)**
- `Parameter()` - Processes query params: filter, filterout, limit, sorted, filter_time
- `Cache()` - Uses singleflight to prevent thundering herd
- `Template()` - Reads `feed_data` from context and generates output format

**Client (`client/`)**
- HTTP client wrapper with retry logic
- Configured via config.Config (timeout, user agent, proxy)

**Configuration (`config/`)**
- Uses viper for environment variable loading
- All config via env vars (e.g., PORT, CACHE_TYPE, REDIS_URL)
- Defaults set in `setDefaults()`

**Cache (`cache/`)**
- Interface with memory and Redis implementations
- Memory cache: LRU with configurable max size
- Cache key format: `grss:cache:{sha256(path:format:limit)}`

### Route Registration Pattern

Each namespace follows this structure:

```
routes/{namespace}/
├── namespace.go   # Namespace metadata
├── init.go        # Registers namespace and routes
└── {route}.go     # Route definition and handler
```

Example from `routes/github/`:
1. `namespace.go` defines namespace metadata
2. `issues.go` defines IssuesRoute with handler
3. `init.go` calls `registry.RegisterNamespace()` and `registry.RegisterRoute()`

### Query Parameters

All routes automatically support:
- `format=rss|atom|json` - Output format (default: rss)
- `limit=N` - Limit number of items
- `filter=regex` - Filter items by title/description
- `filterout=regex` - Exclude items matching regex
- `filter_title=regex` - Filter by title only
- `filter_description=regex` - Filter by description only
- `filter_time=N` - Only items newer than N seconds
- `sorted=asc|desc` - Sort by publication date

## Adding New Routes

1. Create namespace directory: `routes/{namespace}/`
2. Create `namespace.go` with `Namespace` definition
3. Create route file (e.g., `myroute.go`) with:
   - `Route` struct with Path, Name, Example, Handler
   - Handler function returning `(*feed.Data, error)`
4. Create `init.go` to register namespace and routes
5. Run `make generate` to update `routes/routes.go`
6. Build and test

See `routes/github/issues.go` for reference implementation.

## Configuration via Environment Variables

Key environment variables:
- `PORT` - Server port (default: 1200)
- `CACHE_TYPE` - Cache backend: "memory", "redis", or "" (default: "memory")
- `REDIS_URL` - Redis connection URL
- `DEBUG_INFO` - Enable debug logging (default: false)
- `ACCESS_KEY` - API access key for authentication
- `YOUTUBE_KEY` - YouTube API key (comma-separated for rotation)
- `REQUEST_TIMEOUT` - HTTP request timeout in ms (default: 30000)
- `MEMORY_MAX` - Max memory cache entries (default: 256)

See `config/config.go` for full list.

## Testing

- Tests use standard Go testing framework
- Test files: `*_test.go`
- Mock data should be self-contained in test files
- Use `t.Parallel()` for parallel test execution where possible

## Important Patterns

**Error Handling**
- Route handlers return `(*feed.Data, error)`
- Errors are caught by registry.wrapHandler and returned as JSON

**HTTP Requests**
- Use `client.New(config.C)` for all external HTTP requests
- Supports retry logic and proxy configuration

**Code Generation**
- `routes/routes.go` is auto-generated - do not edit manually
- Run `make generate` or `go run scripts/generate-routes.go` after adding namespaces

**Middleware Context**
- Feed data stored in context with key `"feed_data"`
- Parameter middleware operates on this after handler execution
- Template middleware converts it to final format
