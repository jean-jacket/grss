package youtube

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// parseBoolParam parses a boolean query parameter with a default value
// Supports: true, false, 1, 0, yes, no (case-insensitive)
// Returns defaultValue if parameter is not provided
func parseBoolParam(c *gin.Context, param string, defaultValue bool) bool {
	value := c.Query(param)

	// If not provided, return default
	if value == "" {
		return defaultValue
	}

	// Normalize to lowercase
	value = strings.ToLower(strings.TrimSpace(value))

	// Check for true values
	if value == "true" || value == "1" || value == "yes" {
		return true
	}

	// Check for false values
	if value == "false" || value == "0" || value == "no" {
		return false
	}

	// Default to false for any other value (safety)
	return false
}

// getThumbnail returns the best quality thumbnail from a ThumbnailSet
// Priority: maxres > standard > high > medium > default
func getThumbnail(thumbnails ThumbnailSet) *Thumbnail {
	if thumbnails.Maxres != nil {
		return thumbnails.Maxres
	}
	if thumbnails.Standard != nil {
		return thumbnails.Standard
	}
	if thumbnails.High != nil {
		return thumbnails.High
	}
	if thumbnails.Medium != nil {
		return thumbnails.Medium
	}
	return thumbnails.Default
}

// formatDescription formats a description by replacing newlines with <br> tags
func formatDescription(description string) string {
	// Replace all types of line breaks with <br>
	description = strings.ReplaceAll(description, "\r\n", "<br>")
	description = strings.ReplaceAll(description, "\r", "<br>")
	description = strings.ReplaceAll(description, "\n", "<br>")
	return description
}

// renderDescription renders a video description with optional embed
func renderDescription(embed bool, videoID string, img *Thumbnail, description string) string {
	var builder strings.Builder

	if embed {
		// Embed iframe player
		builder.WriteString(fmt.Sprintf(
			`<iframe id="ytplayer" type="text/html" width="640" height="360" src="https://www.youtube-nocookie.com/embed/%s" frameborder="0" allowfullscreen referrerpolicy="strict-origin-when-cross-origin"></iframe>`,
			videoID,
		))
	} else if img != nil {
		// Show thumbnail image
		builder.WriteString(fmt.Sprintf(`<img src="%s">`, img.URL))
	}

	builder.WriteString("<br>")
	builder.WriteString(description)

	return builder.String()
}

// getVideoURL returns the embed URL for a video
func getVideoURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube-nocookie.com/embed/%s?controls=1&autoplay=1&mute=0", videoID)
}

// isYouTubeChannelID validates if a string is a valid YouTube channel ID
// Format: UC followed by 21 alphanumeric/hyphen characters, ending with A, Q, g, or w
func isYouTubeChannelID(id string) bool {
	matched, _ := regexp.MatchString(`^UC[\w-]{21}[AQgw]$`, id)
	return matched
}

// getPlaylistWithShortsFilter converts a channel/uploads playlist ID to filter shorts
// UC... -> UULF... (long-form videos only)
// UU... -> UULF... (long-form videos only)
func getPlaylistWithShortsFilter(id string, filterShorts bool) string {
	if !filterShorts {
		return id
	}

	if strings.HasPrefix(id, "UC") {
		return "UULF" + id[2:]
	}
	if strings.HasPrefix(id, "UU") {
		return "UULF" + id[2:]
	}

	return id
}

// parseDuration converts ISO 8601 duration (PT1H2M3S) to seconds
func parseDuration(duration string) int {
	if duration == "" {
		return 0
	}

	// Remove PT prefix
	duration = strings.TrimPrefix(duration, "PT")

	var hours, minutes, seconds int

	// Parse hours
	if idx := strings.Index(duration, "H"); idx != -1 {
		fmt.Sscanf(duration[:idx], "%d", &hours)
		duration = duration[idx+1:]
	}

	// Parse minutes
	if idx := strings.Index(duration, "M"); idx != -1 {
		fmt.Sscanf(duration[:idx], "%d", &minutes)
		duration = duration[idx+1:]
	}

	// Parse seconds
	if idx := strings.Index(duration, "S"); idx != -1 {
		fmt.Sscanf(duration[:idx], "%d", &seconds)
	}

	return hours*3600 + minutes*60 + seconds
}

// parseRelativeDate parses relative dates like "2 days ago", "1 week ago"
func parseRelativeDate(relativeDate string) time.Time {
	now := time.Now()

	// Handle various relative date formats
	relativeDate = strings.ToLower(relativeDate)
	relativeDate = strings.Split(relativeDate, "(")[0] // Remove timezone info like "(edited)"
	relativeDate = strings.TrimSpace(relativeDate)

	var amount int
	var unit string

	// Try to parse format like "2 days ago"
	n, _ := fmt.Sscanf(relativeDate, "%d %s ago", &amount, &unit)
	if n != 2 {
		// Try format like "1 day ago"
		n, _ = fmt.Sscanf(relativeDate, "%d %s ago", &amount, &unit)
		if n != 2 {
			// Default to now if parsing fails
			return now
		}
	}

	// Normalize unit (remove 's' if present)
	unit = strings.TrimSuffix(unit, "s")

	switch unit {
	case "second":
		return now.Add(-time.Duration(amount) * time.Second)
	case "minute":
		return now.Add(-time.Duration(amount) * time.Minute)
	case "hour":
		return now.Add(-time.Duration(amount) * time.Hour)
	case "day":
		return now.AddDate(0, 0, -amount)
	case "week":
		return now.AddDate(0, 0, -amount*7)
	case "month":
		return now.AddDate(0, -amount, 0)
	case "year":
		return now.AddDate(-amount, 0, 0)
	default:
		return now
	}
}
