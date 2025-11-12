package apple

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/client"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
)

// DesignUpdatesRoute defines the Apple Developer Design updates route
var DesignUpdatesRoute = registry.Route{
	Path:        "/design",
	Name:        "Design updates",
	Maintainers: []string{"jean-jacket"},
	Example:     "/apple/design",
	Parameters:  map[string]interface{}{},
	Description: "Get latest design updates from Apple Developer Design",
	Handler:     designUpdatesHandler,
}

func designUpdatesHandler(c *gin.Context) (*feed.Data, error) {
	// Create HTTP client
	httpClient := client.New(config.C)

	// Fetch the page
	url := "https://developer.apple.com/design/whats-new/"
	data, err := httpClient.Get(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Build feed data
	feedData := &feed.Data{
		Title:       "Apple design updates",
		Link:        url,
		Description: "Latest design updates from Apple Developer",
		Language:    "en",
		Item:        make([]feed.Item, 0),
	}

	// Parse updates from tables
	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		// Get the date for this table
		dateStr := strings.TrimSpace(table.Find(".date").First().Text())

		// Find all topic items in this table
		table.Find(".topic-item").Each(func(j int, topicItem *goquery.Selection) {
			// Get title and link
			titleLink := topicItem.Find("span.topic-title a")
			title := strings.TrimSpace(titleLink.Text())

			if title == "" {
				return
			}

			// Get href and construct full URL
			href, exists := titleLink.Attr("href")
			link := url
			if exists && href != "" {
				if strings.HasPrefix(href, "/") {
					link = "https://developer.apple.com" + href
				} else if strings.HasPrefix(href, "http") {
					link = href
				}
			}

			// Get description
			description := strings.TrimSpace(topicItem.Find("span.description").Text())

			// Parse date
			pubDate := parseDate(dateStr)

			// Generate GUID from title + description + date
			guid := generateGUID(title, description, dateStr)

			item := feed.Item{
				Title:       title,
				Link:        link,
				Description: description,
				PubDate:     pubDate,
				GUID:        guid,
			}

			feedData.Item = append(feedData.Item, item)
		})
	})

	// Set latest update date as feed pubDate
	if len(feedData.Item) > 0 {
		feedData.PubDate = feedData.Item[0].PubDate
	}

	return feedData, nil
}

// generateGUID creates an MD5 hash from the title, description, and date
func generateGUID(title, description, date string) string {
	combined := title + description + date
	hash := md5.Sum([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// parseDate attempts to parse various date formats from the page
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Now()
	}

	// Try common formats
	formats := []string{
		"January 2006",
		"Jan 2006",
		"2006-01-02",
		"January 2, 2006",
		time.RFC3339,
	}

	// Try to extract month and year from strings like "September 2025"
	monthYearRe := regexp.MustCompile(`(?i)(january|february|march|april|may|june|july|august|september|october|november|december)\s+(\d{4})`)
	if matches := monthYearRe.FindStringSubmatch(dateStr); len(matches) >= 3 {
		dateStr = matches[1] + " " + matches[2]
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Now()
}
