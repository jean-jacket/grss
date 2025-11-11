package middleware

import (
	"crypto/sha256"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/config"
)

// Header middleware sets HTTP headers (CORS, ETag, Cache-Control)
func Header() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", config.C.AllowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		// Handle OPTIONS preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		// Process request
		c.Next()

		// Generate ETag from response body
		if c.Writer.Status() == 200 {
			// Get response body
			body := c.Writer.Header().Get("X-Response-Body")
			if body != "" {
				etag := fmt.Sprintf(`"%x"`, sha256.Sum256([]byte(body)))
				c.Header("ETag", etag)

				// Check If-None-Match header
				if c.Request.Header.Get("If-None-Match") == etag {
					c.AbortWithStatus(304)
					return
				}
			}

			// Set Cache-Control
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(config.C.Cache.RouteExpire.Seconds())))
		}
	}
}
