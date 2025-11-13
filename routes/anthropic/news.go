package anthropic

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/client"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
)

// NewsRoute defines the Anthropic news route
var NewsRoute = registry.Route{
	Path:        "/news",
	Name:        "Anthropic News",
	Maintainers: []string{"example"},
	Example:     "/anthropic/news",
	Parameters:  map[string]interface{}{},
	Description: "Get latest news and announcements from Anthropic",
	Handler:     newsHandler,
}

func newsHandler(c *gin.Context) (*feed.Data, error) {
	newsURL := "https://www.anthropic.com/news"

	// Create HTTP client
	httpClient := client.New(config.C)

	// Fetch news page
	data, err := httpClient.Get(newsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news page: %w", err)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Build feed data
	feedData := &feed.Data{
		Title:       "Anthropic News",
		Link:        newsURL,
		Description: "Latest news and announcements from Anthropic",
		Language:    "en",
		Item:        make([]feed.Item, 0),
	}

	// Parse spotlight items
	doc.Find("a[class*='CardSpotlight_spotlightCard']").Each(func(i int, s *goquery.Selection) {
		item := parseNewsItem(s, "spotlight")
		if item.Title != "" {
			feedData.Item = append(feedData.Item, item)
		}
	})

	// Parse regular news items
	doc.Find("a[class*='Card_linkRoot']").Each(func(i int, s *goquery.Selection) {
		item := parseNewsItem(s, "regular")
		if item.Title != "" {
			feedData.Item = append(feedData.Item, item)
		}
	})

	// Sort items by date (most recent first)
	sort.Slice(feedData.Item, func(i, j int) bool {
		return feedData.Item[i].PubDate.After(feedData.Item[j].PubDate)
	})

	// Set latest item date as feed pubDate
	if len(feedData.Item) > 0 {
		feedData.PubDate = feedData.Item[0].PubDate
	}

	return feedData, nil
}

func parseNewsItem(s *goquery.Selection, itemType string) feed.Item {
	var item feed.Item

	// Get link href
	href, exists := s.Attr("href")
	if !exists {
		return item
	}

	// Build full URL
	if !strings.HasPrefix(href, "http") {
		href = "https://www.anthropic.com" + href
	}
	item.Link = href
	item.GUID = href

	// Extract category, title, and date
	// Category is the first <p class="detail-m">
	// Title is the <h3>
	// Date is the last <p class="detail-m">
	var category, title, dateStr string

	s.Find("p.detail-m").Each(func(i int, p *goquery.Selection) {
		text := strings.TrimSpace(p.Text())
		if text == "" {
			return
		}

		// First p.detail-m is the category
		if i == 0 {
			category = text
		} else {
			// Later p.detail-m elements are likely dates
			// We'll take the last one
			dateStr = text
		}
	})

	// Extract title from h3
	s.Find("h3").Each(func(i int, h *goquery.Selection) {
		if i == 0 {
			title = strings.TrimSpace(h.Text())
		}
	})

	item.Title = title
	item.Category = []string{category}

	// Parse date
	pubDate := parseDate(dateStr)
	item.PubDate = pubDate

	return item
}

func parseDate(dateStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}
	}

	// Try various date formats
	formats := []string{
		"Jan 2, 2006",
		"January 2, 2006",
		"Jan 02, 2006",
		"January 02, 2006",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
}
