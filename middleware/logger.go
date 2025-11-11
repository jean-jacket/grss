package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/utils"
)

// Logger middleware logs HTTP requests and responses
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		if query != "" {
			path = path + "?" + query
		}

		// Log request
		utils.LogInfo("%s %s", c.Request.Method, path)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Log response
		utils.LogInfo("%s %s %d %s", c.Request.Method, path, statusCode, utils.FormatDuration(latency))
	}
}
