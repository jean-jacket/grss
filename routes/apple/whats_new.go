package apple

import (
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

// WhatsNewRoute defines the Apple Developer Design What's New route
var WhatsNewRoute = registry.Route{
	Path:        "/whats-new",
	Name:        "What's New - Design",
	Maintainers: []string{"example"},
	Example:     "/apple/whats-new",
	Parameters:  map[string]interface{}{},
	Description: "Get latest updates from Apple Developer Design What's New page",
	Handler:     whatsNewHandler,
}

func whatsNewHandler(c *gin.Context) (*feed.Data, error) {
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
		Title:       "Apple Developer Design - What's New",
		Link:        url,
		Description: "Latest updates from Apple Developer Design",
		Language:    "en",
		Item:        make([]feed.Item, 0),
	}

	// Parse updates - they appear to be in a list or section format
	// This selector may need adjustment based on the actual HTML structure
	doc.Find("article, .update-item, section[class*='update'], li[class*='update']").Each(func(i int, s *goquery.Selection) {
		// Try to find title
		title := ""
		titleElem := s.Find("h2, h3, h4, .title, [class*='title']").First()
		if titleElem.Length() > 0 {
			title = strings.TrimSpace(titleElem.Text())
		}

		// If no title found, skip
		if title == "" {
			return
		}

		// Try to find link
		link := ""
		linkElem := s.Find("a").First()
		if linkElem.Length() > 0 {
			href, exists := linkElem.Attr("href")
			if exists {
				if strings.HasPrefix(href, "/") {
					link = "https://developer.apple.com" + href
				} else if strings.HasPrefix(href, "http") {
					link = href
				}
			}
		}
		if link == "" {
			link = url
		}

		// Try to find description
		description := ""
		descElem := s.Find("p, .description, [class*='description']").First()
		if descElem.Length() > 0 {
			description = strings.TrimSpace(descElem.Text())
		}

		// Try to find date
		dateStr := ""
		dateElem := s.Find("time, .date, [class*='date']").First()
		if dateElem.Length() > 0 {
			dateStr = strings.TrimSpace(dateElem.Text())
		}

		// Parse date
		pubDate := parseDate(dateStr)

		// Try to find category
		category := ""
		categoryElem := s.Find(".category, [class*='category'], .label").First()
		if categoryElem.Length() > 0 {
			category = strings.TrimSpace(categoryElem.Text())
		}

		item := feed.Item{
			Title:       title,
			Link:        link,
			Description: description,
			PubDate:     pubDate,
			GUID:        link,
		}

		if category != "" {
			item.Category = []string{category}
		}

		feedData.Item = append(feedData.Item, item)
	})

	// If we didn't find updates with the above selector, try a more generic approach
	if len(feedData.Item) == 0 {
		doc.Find("article, section, .content li").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text == "" || len(text) < 10 {
				return
			}

			// Try to find any link
			link := ""
			linkElem := s.Find("a").First()
			if linkElem.Length() > 0 {
				href, exists := linkElem.Attr("href")
				if exists {
					if strings.HasPrefix(href, "/") {
						link = "https://developer.apple.com" + href
					} else if strings.HasPrefix(href, "http") {
						link = href
					}
				}
			}
			if link == "" {
				link = url
			}

			// Extract title (first sentence or up to 100 chars)
			title := text
			if len(title) > 100 {
				title = title[:100] + "..."
			}
			if idx := strings.Index(title, ". "); idx > 0 && idx < 100 {
				title = title[:idx]
			}

			item := feed.Item{
				Title:       title,
				Link:        link,
				Description: text,
				PubDate:     time.Now(),
				GUID:        link,
			}

			feedData.Item = append(feedData.Item, item)
		})
	}

	// Set latest update date as feed pubDate
	if len(feedData.Item) > 0 {
		feedData.PubDate = feedData.Item[0].PubDate
	}

	return feedData, nil
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
