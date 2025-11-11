package middleware

import (
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/utils"
)

// Parameter middleware processes query parameters (filter, limit, etc.)
func Parameter() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request first
		c.Next()

		// Get data from context
		dataInterface, exists := c.Get(ContextKeyData)
		if !exists {
			return
		}

		data, ok := dataInterface.(*feed.Data)
		if !ok {
			return
		}

		// Apply filters and transformations
		items := data.Item

		// Filter by regex
		if filter := c.Query("filter"); filter != "" {
			items = filterItems(items, filter, false)
		}

		// Filter out by regex
		if filterOut := c.Query("filterout"); filterOut != "" {
			items = filterItems(items, filterOut, true)
		}

		// Filter by title
		if filterTitle := c.Query("filter_title"); filterTitle != "" {
			items = filterByTitle(items, filterTitle, false)
		}

		// Filter by description
		if filterDesc := c.Query("filter_description"); filterDesc != "" {
			items = filterByDescription(items, filterDesc, false)
		}

		// Filter by time
		if filterTime := c.Query("filter_time"); filterTime != "" {
			items = filterByTime(items, filterTime)
		}

		// Sort items
		if sorted := c.Query("sorted"); sorted != "" {
			items = sortItems(items, sorted)
		}

		// Limit items
		if limitStr := c.Query("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				if limit < len(items) {
					items = items[:limit]
				}
			}
		}

		// Update data with processed items
		data.Item = items
		c.Set(ContextKeyData, data)
	}
}

// filterItems filters items by regex pattern
func filterItems(items []feed.Item, pattern string, inverse bool) []feed.Item {
	re, err := regexp.Compile(pattern)
	if err != nil {
		utils.LogError("Invalid regex pattern: %v", err)
		return items
	}

	filtered := []feed.Item{}
	for _, item := range items {
		// Check if pattern matches title or description
		matches := re.MatchString(item.Title) || re.MatchString(item.Description)

		if inverse {
			matches = !matches
		}

		if matches {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// filterByTitle filters items by title regex
func filterByTitle(items []feed.Item, pattern string, inverse bool) []feed.Item {
	re, err := regexp.Compile(pattern)
	if err != nil {
		utils.LogError("Invalid regex pattern: %v", err)
		return items
	}

	filtered := []feed.Item{}
	for _, item := range items {
		matches := re.MatchString(item.Title)

		if inverse {
			matches = !matches
		}

		if matches {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// filterByDescription filters items by description regex
func filterByDescription(items []feed.Item, pattern string, inverse bool) []feed.Item {
	re, err := regexp.Compile(pattern)
	if err != nil {
		utils.LogError("Invalid regex pattern: %v", err)
		return items
	}

	filtered := []feed.Item{}
	for _, item := range items {
		matches := re.MatchString(item.Description)

		if inverse {
			matches = !matches
		}

		if matches {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// filterByTime filters items by time (only items newer than N seconds)
func filterByTime(items []feed.Item, timeStr string) []feed.Item {
	seconds, err := strconv.Atoi(timeStr)
	if err != nil {
		utils.LogError("Invalid filter_time value: %v", err)
		return items
	}

	cutoff := time.Now().Add(-time.Duration(seconds) * time.Second)

	filtered := []feed.Item{}
	for _, item := range items {
		if item.PubDate.After(cutoff) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// sortItems sorts items by publication date
func sortItems(items []feed.Item, order string) []feed.Item {
	sorted := make([]feed.Item, len(items))
	copy(sorted, items)

	sort.Slice(sorted, func(i, j int) bool {
		if order == "desc" {
			return sorted[i].PubDate.After(sorted[j].PubDate)
		}
		return sorted[i].PubDate.Before(sorted[j].PubDate)
	})

	return sorted
}
