package feed

import (
	"encoding/json"
)

// JSONFeed represents a JSON Feed 1.1
type JSONFeed struct {
	Version     string         `json:"version"`
	Title       string         `json:"title"`
	HomePageURL string         `json:"home_page_url,omitempty"`
	FeedURL     string         `json:"feed_url,omitempty"`
	Description string         `json:"description,omitempty"`
	Icon        string         `json:"icon,omitempty"`
	Favicon     string         `json:"favicon,omitempty"`
	Language    string         `json:"language,omitempty"`
	Authors     []JSONAuthor   `json:"authors,omitempty"`
	Items       []JSONItem     `json:"items"`
}

// JSONAuthor represents an author in JSON Feed
type JSONAuthor struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// JSONItem represents a single item in JSON Feed
type JSONItem struct {
	ID            string         `json:"id"`
	URL           string         `json:"url,omitempty"`
	Title         string         `json:"title,omitempty"`
	ContentHTML   string         `json:"content_html,omitempty"`
	ContentText   string         `json:"content_text,omitempty"`
	Summary       string         `json:"summary,omitempty"`
	DatePublished string         `json:"date_published,omitempty"`
	DateModified  string         `json:"date_modified,omitempty"`
	Authors       []JSONAuthor   `json:"authors,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	Attachments   []JSONAttachment `json:"attachments,omitempty"`
}

// JSONAttachment represents an attachment in JSON Feed
type JSONAttachment struct {
	URL      string `json:"url"`
	MIMEType string `json:"mime_type"`
	SizeInBytes int64 `json:"size_in_bytes,omitempty"`
}

// GenerateJSON converts Data to JSON Feed 1.1 format
func GenerateJSON(data *Data, currentURL string) (string, error) {
	feed := JSONFeed{
		Version:     "https://jsonfeed.org/version/1.1",
		Title:       data.Title,
		HomePageURL: data.Link,
		FeedURL:     currentURL,
		Description: data.Description,
		Icon:        data.Icon,
		Language:    data.Language,
	}

	// Set authors
	if data.Author != "" {
		feed.Authors = []JSONAuthor{
			{Name: data.Author},
		}
	}

	// Convert items
	feed.Items = make([]JSONItem, len(data.Item))
	for i, item := range data.Item {
		jsonItem := JSONItem{
			ID:          item.GUID,
			URL:         item.Link,
			Title:       item.Title,
			ContentHTML: item.Description,
		}

		// Set ID from link if not provided
		if jsonItem.ID == "" {
			jsonItem.ID = item.Link
		}

		// Set dates
		if !item.PubDate.IsZero() {
			jsonItem.DatePublished = formatRFC3339(item.PubDate)
		}
		if !item.Updated.IsZero() {
			jsonItem.DateModified = formatRFC3339(item.Updated)
		}

		// Set authors
		if item.Author != "" {
			jsonItem.Authors = []JSONAuthor{
				{Name: item.Author},
			}
		}

		// Set tags
		if len(item.Category) > 0 {
			jsonItem.Tags = item.Category
		}

		// Set attachments
		if item.EnclosureURL != "" {
			jsonItem.Attachments = []JSONAttachment{
				{
					URL:         item.EnclosureURL,
					MIMEType:    item.EnclosureType,
					SizeInBytes: item.EnclosureLength,
				},
			}
		}

		feed.Items[i] = jsonItem
	}

	// Marshal to JSON
	output, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}

	return string(output), nil
}
