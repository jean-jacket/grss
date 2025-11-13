# GRSS 🌾

A high-performance RSS feed generator written in Go, inspired by [RSSHub](https://github.com/DIYgod/RSSHub).

## Philosophy

GRSS is designed as a **minimal, extensible foundation** for generating your own RSS feeds. Unlike RSSHub's approach of providing hundreds of pre-built routes, GRSS encourages you to:

1. **Start lean** - Include only the routes you actually use
2. **Add on-demand** - Implement new routes as you need them
3. **Deploy efficiently** - Run it for pennies per month on platforms like Fly.io

### AI-First Development

GRSS is architected for AI-assisted development. The codebase is structured so that AI coding assistants (like Claude, GPT, etc.) can easily:
- Understand the route registration pattern
- Implement new routes following existing conventions
- Generate code that integrates seamlessly with the middleware chain

### How to Add a New Route

1. **Fork the project** and clone locally
2. **Research the data source** - Find the API endpoint, HTML structure, or data format you want to parse
3. **Prompt an AI assistant** - Share the research and ask it to implement a new route following the GRSS pattern (see `CLAUDE.md` for architecture details)
4. **Test locally** - Verify the route works and returns proper feed data
5. **Deploy** - Push to Fly.io or your preferred hosting platform 

## Included Routes

| Route | Description | Example | Parameters |
|-------|-------------|---------|------------|
| `/example/hello` | Hello World feed (demo) | `/example/hello` | - |
| `/github/issue/:user/:repo` | Repository issues | `/github/issue/golang/go` | `state=open\|closed\|all` |
| `/apple/design` | Apple Developer Design updates | `/apple/design` | - |
| `/youtube/channel/:id` | Channel videos by channel ID | `/youtube/channel/UCDwDMPOZfxVV0x_dz0eQ8KQ` | `embed=true\|false`, `filterShorts=true\|false` |
| `/youtube/user/:username` | Channel videos by username/handle | `/youtube/user/@JFlaMusic` | `embed=true\|false`, `filterShorts=true\|false` |
| `/youtube/playlist/:id` | Playlist videos | `/youtube/playlist/PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf` | `embed=true\|false` |


## Query Parameters

All routes support:
- `format=rss|atom|json` - Output format (default: rss)
- `limit=N` - Limit items
- `filter=regex` - Filter by title/description
- `filterout=regex` - Exclude items
- `sorted=asc|desc` - Sort by date

## Build Instructions

```bash
# Clone repository
git clone https://github.com/jean-jacket/grss.git
cd grss

# Build
go build -o grss cmd/grss/main.go

# Run
./grss

# Test a specific route (debug mode)
./grss -test-route /github/issue/golang/go
./grss -test-route /github/issue/golang/go -test-limit 3
```

## Testing Routes

GRSS includes a built-in route testing feature for debugging and development:

```bash
# Test a route and see the output
./grss -test-route /github/issue/golang/go

# Limit the number of items displayed
./grss -test-route /github/issue/golang/go -test-limit 3
```

This will:
- Execute the route handler without starting the server
- Display timing information (how long the route took to execute)
- Print the feed metadata (title, description, etc.)
- Show the first few feed items with their details
- Display a sample JSON output
- Exit immediately

Perfect for quickly verifying that a route is working correctly during development!

## Comparison to RSSHub

GRSS is a ground-up rewrite of RSSHub in Go, focusing on performance and simplicity.

### Architecture
- **RSSHub**: Node.js with Hono, cluster mode for multi-core
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
