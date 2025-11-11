package utils

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// LogRequest logs HTTP requests and responses
func LogRequest(method, url string, resp *http.Response, duration time.Duration, err error) {
	if err != nil {
		log.Printf("[ERROR] %s %s - Error: %v (took %v)", method, url, err, duration)
		return
	}

	statusColor := getStatusColor(resp.StatusCode)
	log.Printf("[%sINFO%s] %s %s %d %v", statusColor, resetColor, method, url, resp.StatusCode, duration)
}

// LogInfo logs an informational message
func LogInfo(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// LogError logs an error message
func LogError(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// LogDebug logs a debug message
func LogDebug(format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

// LogWarn logs a warning message
func LogWarn(format string, args ...interface{}) {
	log.Printf("[WARN] "+format, args...)
}

const (
	resetColor  = "\033[0m"
	redColor    = "\033[31m"
	greenColor  = "\033[32m"
	yellowColor = "\033[33m"
	blueColor   = "\033[34m"
)

func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return greenColor
	case statusCode >= 300 && statusCode < 400:
		return blueColor
	case statusCode >= 400 && statusCode < 500:
		return yellowColor
	case statusCode >= 500:
		return redColor
	default:
		return resetColor
	}
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dus", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
