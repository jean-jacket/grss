package feed

import (
	"encoding/xml"
	"time"
)

// AtomFeed represents an Atom 1.0 feed
type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Xmlns   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    []AtomFeedLink `xml:"link"`
	Updated string      `xml:"updated"`
	ID      string      `xml:"id"`
	Subtitle string     `xml:"subtitle,omitempty"`
	Icon    string      `xml:"icon,omitempty"`
	Logo    string      `xml:"logo,omitempty"`
	Author  *AtomAuthor `xml:"author,omitempty"`
	Entries []AtomEntry `xml:"entry"`
}

// AtomFeedLink represents a link in the Atom feed
type AtomFeedLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

// AtomAuthor represents the author of the feed
type AtomAuthor struct {
	Name string `xml:"name"`
}

// AtomEntry represents a single Atom entry
type AtomEntry struct {
	Title   string           `xml:"title"`
	Link    []AtomFeedLink   `xml:"link"`
	ID      string           `xml:"id"`
	Updated string           `xml:"updated"`
	Summary *AtomText        `xml:"summary,omitempty"`
	Content *AtomContent     `xml:"content,omitempty"`
	Author  *AtomAuthor      `xml:"author,omitempty"`
	Category []AtomCategory  `xml:"category,omitempty"`
}

// AtomText represents text content
type AtomText struct {
	Type  string `xml:"type,attr,omitempty"`
	Value string `xml:",chardata"`
}

// AtomContent represents content with HTML
type AtomContent struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",cdata"`
}

// AtomCategory represents a category
type AtomCategory struct {
	Term string `xml:"term,attr"`
}

// GenerateAtom converts Data to Atom 1.0 XML format
func GenerateAtom(data *Data, currentURL string) (string, error) {
	atom := AtomFeed{
		Xmlns: "http://www.w3.org/2005/Atom",
		Title: data.Title,
		Link: []AtomFeedLink{
			{Href: data.Link, Rel: "alternate"},
			{Href: currentURL, Rel: "self", Type: "application/atom+xml"},
		},
		ID:       data.Link,
		Subtitle: data.Subtitle,
		Icon:     data.Icon,
		Logo:     data.Logo,
	}

	// Set updated time
	if !data.PubDate.IsZero() {
		atom.Updated = formatRFC3339(data.PubDate)
	} else {
		atom.Updated = formatRFC3339(time.Now())
	}

	// Set author
	if data.Author != "" {
		atom.Author = &AtomAuthor{Name: data.Author}
	}

	// Convert items
	atom.Entries = make([]AtomEntry, len(data.Item))
	for i, item := range data.Item {
		entry := AtomEntry{
			Title: item.Title,
			Link: []AtomFeedLink{
				{Href: item.Link, Rel: "alternate"},
			},
			ID: item.GUID,
		}

		// Set ID from link if not provided
		if entry.ID == "" {
			entry.ID = item.Link
		}

		// Set updated time
		if !item.Updated.IsZero() {
			entry.Updated = formatRFC3339(item.Updated)
		} else if !item.PubDate.IsZero() {
			entry.Updated = formatRFC3339(item.PubDate)
		} else {
			entry.Updated = formatRFC3339(time.Now())
		}

		// Set content
		if item.Description != "" {
			entry.Content = &AtomContent{
				Type:  "html",
				Value: item.Description,
			}
		}

		// Set author
		if item.Author != "" {
			entry.Author = &AtomAuthor{Name: item.Author}
		}

		// Set categories
		if len(item.Category) > 0 {
			entry.Category = make([]AtomCategory, len(item.Category))
			for j, cat := range item.Category {
				entry.Category[j] = AtomCategory{Term: cat}
			}
		}

		atom.Entries[i] = entry
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(atom, "", "  ")
	if err != nil {
		return "", err
	}

	return xml.Header + string(output), nil
}

// formatRFC3339 formats a time.Time to RFC3339 format (Atom date format)
func formatRFC3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
