package feed

import (
	"encoding/xml"
	"time"
)

// RSS represents an RSS 2.0 feed
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	AtomNS  string   `xml:"xmlns:atom,attr"`
	Content string   `xml:"xmlns:content,attr"`
	MediaNS string   `xml:"xmlns:media,attr"`
	Channel Channel  `xml:"channel"`
}

// Channel represents the RSS channel
type Channel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language,omitempty"`
	PubDate       string    `xml:"pubDate,omitempty"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	TTL           int       `xml:"ttl,omitempty"`
	AtomLink      *AtomLink `xml:"atom:link,omitempty"`
	Image         *Image    `xml:"image,omitempty"`
	Items         []RSSItem `xml:"item"`
}

// AtomLink represents an Atom link element in RSS
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

// Image represents the channel image
type Image struct {
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

// RSSItem represents a single RSS item
type RSSItem struct {
	Title       string       `xml:"title"`
	Link        string       `xml:"link"`
	Description CDATA        `xml:"description"`
	PubDate     string       `xml:"pubDate,omitempty"`
	GUID        *GUID        `xml:"guid,omitempty"`
	Author      string       `xml:"author,omitempty"`
	Category    []string     `xml:"category,omitempty"`
	Comments    string       `xml:"comments,omitempty"`
	Enclosure   *Enclosure   `xml:"enclosure,omitempty"`
	Content     *ContentHTML `xml:"content:encoded,omitempty"`
}

// GUID represents the GUID element
type GUID struct {
	Value       string `xml:",chardata"`
	IsPermaLink bool   `xml:"isPermaLink,attr,omitempty"`
}

// Enclosure represents the enclosure element
type Enclosure struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length int64  `xml:"length,attr"`
}

// ContentHTML represents content:encoded element
type ContentHTML struct {
	Value string `xml:",cdata"`
}

// CDATA wraps a string in CDATA tags
type CDATA struct {
	Value string `xml:",cdata"`
}

// GenerateRSS converts Data to RSS 2.0 XML format
func GenerateRSS(data *Data, currentURL string) (string, error) {
	rss := RSS{
		Version: "2.0",
		AtomNS:  "http://www.w3.org/2005/Atom",
		Content: "http://purl.org/rss/1.0/modules/content/",
		MediaNS: "http://search.yahoo.com/mrss/",
		Channel: Channel{
			Title:       data.Title,
			Link:        data.Link,
			Description: data.Description,
			Language:    data.Language,
			TTL:         data.TTL,
			AtomLink: &AtomLink{
				Href: currentURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
		},
	}

	// Set dates
	if !data.PubDate.IsZero() {
		rss.Channel.PubDate = formatRFC822(data.PubDate)
	}
	if !data.LastBuildDate.IsZero() {
		rss.Channel.LastBuildDate = formatRFC822(data.LastBuildDate)
	} else if !data.PubDate.IsZero() {
		rss.Channel.LastBuildDate = formatRFC822(data.PubDate)
	}

	// Set image
	if data.Image != "" {
		rss.Channel.Image = &Image{
			URL:   data.Image,
			Title: data.Title,
			Link:  data.Link,
		}
	}

	// Convert items
	rss.Channel.Items = make([]RSSItem, len(data.Item))
	for i, item := range data.Item {
		rssItem := RSSItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: CDATA{Value: item.Description},
			Author:      item.Author,
			Category:    item.Category,
			Comments:    item.Comments,
		}

		// Set GUID
		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}
		rssItem.GUID = &GUID{
			Value:       guid,
			IsPermaLink: false,
		}

		// Set date
		if !item.PubDate.IsZero() {
			rssItem.PubDate = formatRFC822(item.PubDate)
		}

		// Set enclosure
		if item.EnclosureURL != "" {
			rssItem.Enclosure = &Enclosure{
				URL:    item.EnclosureURL,
				Type:   item.EnclosureType,
				Length: item.EnclosureLength,
			}
		}

		// Set content:encoded if description is HTML
		if item.Description != "" {
			rssItem.Content = &ContentHTML{Value: item.Description}
		}

		rss.Channel.Items[i] = rssItem
	}

	// Marshal to XML
	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return "", err
	}

	return xml.Header + string(output), nil
}

// formatRFC822 formats a time.Time to RFC822 format (RSS date format)
func formatRFC822(t time.Time) string {
	return t.UTC().Format(time.RFC1123Z)
}
