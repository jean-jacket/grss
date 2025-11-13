package anthropic

import (
	"fmt"
	"regexp"
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

// EngineeringRoute defines the Anthropic engineering blog route
var EngineeringRoute = registry.Route{
	Path:        "/engineering",
	Name:        "Anthropic Engineering Blog",
	Maintainers: []string{"example"},
	Example:     "/anthropic/engineering",
	Parameters:  map[string]interface{}{},
	Description: "Get latest posts from Anthropic's engineering blog",
	Handler:     engineeringHandler,
}

func engineeringHandler(c *gin.Context) (*feed.Data, error) {
	engineeringURL := "https://www.anthropic.com/engineering"

	// Create HTTP client
	httpClient := client.New(config.C)

	// Fetch engineering page
	data, err := httpClient.Get(engineeringURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch engineering page: %w", err)
	}

	htmlContent := string(data)

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Build feed data
	feedData := &feed.Data{
		Title:       "Anthropic Engineering Blog",
		Link:        engineeringURL,
		Description: "Latest posts from Anthropic's engineering team",
		Language:    "en",
		Item:        make([]feed.Item, 0),
	}

	// Parse featured item (if present)
	// Look for any anchor that contains a "Featured" label
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		// Check if this link contains a "Featured" span/label
		if s.Find("span").FilterFunction(func(j int, span *goquery.Selection) bool {
			text := strings.TrimSpace(span.Text())
			return strings.Contains(strings.ToLower(text), "featured")
		}).Length() > 0 {
			item := parseFeaturedItem(s, htmlContent)
			if item.Title != "" {
				feedData.Item = append(feedData.Item, item)
			}
		}
	})

	// Parse regular engineering items
	// These have h3 with class "display-sans-s bold" and a div with class "detail-m ArticleList_date__2VTRg"
	doc.Find("h3.display-sans-s.bold").Each(func(i int, s *goquery.Selection) {
		item := parseEngineeringItem(s)
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

func parseFeaturedItem(s *goquery.Selection, htmlContent string) feed.Item {
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

	// Extract title from h2 (try specific classes first, then any h2)
	title := strings.TrimSpace(s.Find("h2.display-sans-l").Text())
	if title == "" {
		title = strings.TrimSpace(s.Find("h2").Text())
	}
	item.Title = title

	// Extract summary if present (try specific class first, then any paragraph)
	summary := strings.TrimSpace(s.Find("p[class*='summary']").Text())
	if summary == "" {
		summary = strings.TrimSpace(s.Find("p").Text())
	}
	if summary != "" {
		item.Description = summary
	}

	// Extract slug from href for date lookup
	slug := strings.TrimPrefix(href, "https://www.anthropic.com/engineering/")
	slug = strings.TrimPrefix(slug, "/engineering/")
	slug = strings.TrimLeft(slug, "/")

	// Try to extract date from embedded JS
	// Look for the slug in the HTML and walk back to find publishedOn
	if slug != "" {
		pubDate := extractDateFromJS(htmlContent, slug)
		item.PubDate = pubDate
	}

	return item
}

func parseEngineeringItem(h3 *goquery.Selection) feed.Item {
	var item feed.Item

	// Get title
	title := strings.TrimSpace(h3.Text())
	item.Title = title

	// Find the parent anchor tag to get the link
	parent := h3.ParentsFiltered("a").First()
	if parent.Length() > 0 {
		href, exists := parent.Attr("href")
		if exists {
			if !strings.HasPrefix(href, "http") {
				href = "https://www.anthropic.com" + href
			}
			item.Link = href
			item.GUID = href
		}
	}

	// Find the date div - it should be a sibling or nearby in the structure
	// Look for div with classes "detail-m ArticleList_date__2VTRg"
	dateDiv := h3.Parent().Find("div.detail-m.ArticleList_date__2VTRg").First()
	if dateDiv.Length() == 0 {
		// Try looking in the parent's siblings
		dateDiv = h3.Parent().NextAll().Find("div.detail-m.ArticleList_date__2VTRg").First()
	}
	if dateDiv.Length() == 0 {
		// Try looking in parent's parent
		dateDiv = h3.Parent().Parent().Find("div.detail-m.ArticleList_date__2VTRg").First()
	}

	if dateDiv.Length() > 0 {
		dateStr := strings.TrimSpace(dateDiv.Text())
		item.PubDate = parseEngineeringDate(dateStr)
	}

	return item
}

func extractDateFromJS(htmlContent, slug string) time.Time {
	// Escape special regex characters in slug
	escapedSlug := regexp.QuoteMeta(slug)

	// Look for pattern: "publishedOn":"YYYY-MM-DD","slug":{"_type":"slug","current":"<slug>"}
	// or: "slug":{"_type":"slug","current":"<slug>"}...walk back to find publishedOn

	// Find the slug in the JSON
	slugPattern := `"slug":\{"_type":"slug","current":"` + escapedSlug + `"`
	slugRegex := regexp.MustCompile(slugPattern)

	matches := slugRegex.FindStringIndex(htmlContent)
	if matches == nil {
		return time.Time{}
	}

	// Get context around the match (500 chars before should be enough)
	start := matches[0] - 500
	if start < 0 {
		start = 0
	}
	contextBefore := htmlContent[start:matches[0]]

	// Look for publishedOn in the context before
	datePattern := `"publishedOn":"(\d{4}-\d{2}-\d{2})"`
	dateRegex := regexp.MustCompile(datePattern)
	dateMatches := dateRegex.FindStringSubmatch(contextBefore)

	if len(dateMatches) > 1 {
		// Parse the date
		t, err := time.Parse("2006-01-02", dateMatches[1])
		if err == nil {
			return t
		}
	}

	return time.Time{}
}

func parseEngineeringDate(dateStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}
	}

	// Try various date formats
	// Expected format: "Oct 20, 2025"
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
