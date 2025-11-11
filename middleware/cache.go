package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/cache"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/utils"
	"golang.org/x/sync/singleflight"
)

var sf singleflight.Group

// Cache middleware provides caching with request deduplication
func Cache(c cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Skip caching if disabled
		if config.C.Cache.Type == "" {
			ctx.Next()
			return
		}

		// Generate cache key
		path := ctx.Request.URL.Path
		format := ctx.DefaultQuery("format", "rss")
		limit := ctx.Query("limit")
		keyData := fmt.Sprintf("%s:%s:%s", path, format, limit)
		cacheKey := fmt.Sprintf("grss:cache:%x", sha256.Sum256([]byte(keyData)))

		// Try to get from cache
		cached, err := c.Get(ctx.Request.Context(), cacheKey)
		if err == nil && cached != "" {
			// Cache hit
			ctx.Header("GRSS-Cache-Status", "HIT")
			ctx.Header("Content-Type", getContentType(format))
			ctx.String(http.StatusOK, cached)
			ctx.Abort()
			return
		}

		// Cache miss - use singleflight to prevent thundering herd
		result, err, _ := sf.Do(cacheKey, func() (interface{}, error) {
			// Create a custom response writer to capture the response
			writer := &responseWriter{
				ResponseWriter: ctx.Writer,
				body:           []byte{},
			}
			ctx.Writer = writer

			// Execute handler chain
			ctx.Next()

			// Get the response body
			response := string(writer.body)

			// Cache the response if status is 200
			if ctx.Writer.Status() == http.StatusOK && response != "" {
				go func() {
					bgCtx := context.Background()
					err := c.Set(bgCtx, cacheKey, response, config.C.Cache.RouteExpire)
					if err != nil {
						utils.LogError("Failed to cache response: %v", err)
					}
				}()
			}

			return response, nil
		})

		// If we got the result from singleflight, write it
		if err == nil && result != nil {
			response := result.(string)
			if response != "" && ctx.Writer.Status() == http.StatusOK {
				ctx.Header("GRSS-Cache-Status", "MISS")
			}
		}
	}
}

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body = append(w.body, []byte(s)...)
	return w.ResponseWriter.WriteString(s)
}

// getContentType returns the content type for a given format
func getContentType(format string) string {
	switch format {
	case "atom":
		return "application/atom+xml; charset=utf-8"
	case "json":
		return "application/json; charset=utf-8"
	default:
		return "application/rss+xml; charset=utf-8"
	}
}
