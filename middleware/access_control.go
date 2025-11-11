package middleware

import (
	"crypto/md5"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/config"
)

var bypassPaths = map[string]bool{
	"/":            true,
	"/robots.txt":  true,
	"/favicon.ico": true,
	"/healthz":     true,
}

// AccessControl middleware validates access key or code
func AccessControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Check if path is in bypass list
		if bypassPaths[path] {
			c.Next()
			return
		}

		// Check if access key is configured
		if config.C.AccessKey == "" {
			c.Next()
			return
		}

		// Get key and code from query parameters
		key := c.Query("key")
		code := c.Query("code")

		// Validate key
		if key == config.C.AccessKey {
			c.Next()
			return
		}

		// Validate code (MD5 of path + access key)
		expectedCode := fmt.Sprintf("%x", md5.Sum([]byte(path+config.C.AccessKey)))
		if code == expectedCode {
			c.Next()
			return
		}

		// Authentication failed
		c.JSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"message": "Access denied",
			},
		})
		c.Abort()
	}
}
