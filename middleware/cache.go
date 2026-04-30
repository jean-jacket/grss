package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/cache"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
	"github.com/jean-jacket/grss/utils"
	"golang.org/x/sync/singleflight"
)

var sf singleflight.Group

// CachingHandlerWrapper returns a registry.HandlerWrapper that caches the
// route handler's *feed.Data (gob-encoded) and serves matching requests
// straight from cache. It also emits standard HTTP caching headers
// (ETag + Cache-Control) and answers conditional requests with 304.
//
// Caching the parsed feed.Data — rather than the rendered XML — is much more
// memory-efficient (one entry serves all output formats) and lets the
// downstream Parameter/Template middlewares run normally.
func CachingHandlerWrapper(c cache.Cache) registry.HandlerWrapper {
	return func(handler registry.RouteHandler) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			// If caching is disabled or no backend is configured, fall through.
			if c == nil || config.C.Cache.Type == "" {
				runUncached(ctx, handler)
				return
			}

			key := cacheKey(ctx.Request.URL.Path)

			// Cache hit?
			if cached, err := c.Get(ctx.Request.Context(), key); err == nil && cached != "" {
				if data, ok := decodeFeed(cached); ok {
					etag := computeETag(cached)
					applyCacheHeaders(ctx, etag)
					ctx.Header("GRSS-Cache-Status", "HIT")
					if ctx.GetHeader("If-None-Match") == etag {
						ctx.Status(http.StatusNotModified)
						ctx.Abort()
						return
					}
					ctx.Set("feed_data", data)
					return
				}
				// Decode failure: fall through to regenerate.
			}

			// Cache miss — coalesce concurrent fetches for the same key.
			result, err, _ := sf.Do(key, func() (interface{}, error) {
				data, err := handler(ctx)
				if err != nil {
					return nil, err
				}
				return data, nil
			})
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"message": err.Error()},
				})
				ctx.Abort()
				return
			}
			data, _ := result.(*feed.Data)
			if data == nil {
				return
			}

			// Encode and store.
			encoded, ok := encodeFeed(data)
			if ok {
				etag := computeETag(encoded)
				applyCacheHeaders(ctx, etag)
				ctx.Header("GRSS-Cache-Status", "MISS")
				go func() {
					if err := c.Set(context.Background(), key, encoded, config.C.Cache.RouteExpire); err != nil {
						utils.LogError("Failed to cache feed: %v", err)
					}
				}()
			}
			ctx.Set("feed_data", data)
		}
	}
}

// runUncached mirrors the default registry wrapHandler so behaviour is
// identical when caching is disabled.
func runUncached(ctx *gin.Context, handler registry.RouteHandler) {
	data, err := handler(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"message": err.Error()},
		})
		ctx.Abort()
		return
	}
	ctx.Set("feed_data", data)
}

func cacheKey(path string) string {
	return fmt.Sprintf("grss:cache:v2:%x", sha256.Sum256([]byte(path)))
}

func encodeFeed(data *feed.Data) (string, bool) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		utils.LogError("Failed to gob-encode feed: %v", err)
		return "", false
	}
	return buf.String(), true
}

func decodeFeed(encoded string) (*feed.Data, bool) {
	var data feed.Data
	if err := gob.NewDecoder(bytes.NewReader([]byte(encoded))).Decode(&data); err != nil {
		return nil, false
	}
	return &data, true
}

func computeETag(payload string) string {
	h := fnv.New64a()
	h.Write([]byte(payload))
	return fmt.Sprintf(`W/"%x"`, h.Sum64())
}

func applyCacheHeaders(ctx *gin.Context, etag string) {
	ctx.Header("ETag", etag)
	maxAge := int(config.C.Cache.RouteExpire.Seconds())
	if maxAge > 0 {
		ctx.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	}
}
