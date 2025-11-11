package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/utils"
)

const (
	// ContextKeyData is the key for storing feed data in context
	ContextKeyData = "feed_data"
)

// Template middleware converts Data object to feed format (RSS/Atom/JSON)
func Template() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute handler first
		c.Next()

		// Check if data was set by handler
		dataInterface, exists := c.Get(ContextKeyData)
		if !exists {
			// No data to render, continue
			return
		}

		// Type assert to feed.Data
		data, ok := dataInterface.(*feed.Data)
		if !ok {
			utils.LogError("Invalid data type in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"message": "Internal server error",
				},
			})
			c.Abort()
			return
		}

		// Get format from query parameter
		format := c.DefaultQuery("format", "rss")

		// Get current URL
		currentURL := c.Request.URL.String()
		if c.Request.Host != "" {
			scheme := "http"
			if c.Request.TLS != nil {
				scheme = "https"
			}
			currentURL = scheme + "://" + c.Request.Host + c.Request.URL.String()
		}

		// Generate feed based on format
		var output string
		var err error
		var contentType string

		switch format {
		case "atom":
			output, err = feed.GenerateAtom(data, currentURL)
			contentType = "application/atom+xml; charset=utf-8"
		case "json":
			output, err = feed.GenerateJSON(data, currentURL)
			contentType = "application/json; charset=utf-8"
		case "rss":
			fallthrough
		default:
			output, err = feed.GenerateRSS(data, currentURL)
			contentType = "application/rss+xml; charset=utf-8"
		}

		if err != nil {
			utils.LogError("Failed to generate feed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"message": "Failed to generate feed",
				},
			})
			c.Abort()
			return
		}

		// Set content type and return feed
		c.Header("Content-Type", contentType)
		c.String(http.StatusOK, output)
	}
}
