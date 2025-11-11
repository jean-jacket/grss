package feed

import "time"

// Data represents the structured data returned by route handlers
type Data struct {
	// Feed metadata
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description,omitempty"`
	Language    string    `json:"language,omitempty"`
	PubDate     time.Time `json:"pubDate,omitempty"`
	LastBuildDate time.Time `json:"lastBuildDate,omitempty"`
	TTL         int       `json:"ttl,omitempty"`
	AllowEmpty  bool      `json:"allowEmpty,omitempty"`

	// Items
	Item []Item `json:"item"`

	// Optional metadata
	Author   string `json:"author,omitempty"`
	Image    string `json:"image,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Logo     string `json:"logo,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`

	// iTunes podcast support
	ItunesAuthor   string `json:"itunes_author,omitempty"`
	ItunesCategory string `json:"itunes_category,omitempty"`
	ItunesExplicit bool   `json:"itunes_explicit,omitempty"`
}

// Item represents a single feed item
type Item struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description,omitempty"`
	PubDate     time.Time `json:"pubDate,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`

	// Optional fields
	Author   string   `json:"author,omitempty"`
	Category []string `json:"category,omitempty"`
	GUID     string   `json:"guid,omitempty"`
	Comments string   `json:"comments,omitempty"`

	// Enclosure (media)
	EnclosureURL    string `json:"enclosure_url,omitempty"`
	EnclosureType   string `json:"enclosure_type,omitempty"`
	EnclosureLength int64  `json:"enclosure_length,omitempty"`

	// Media RSS
	Media *Media `json:"media,omitempty"`

	// Torrent support
	Torrent *Torrent `json:"torrent,omitempty"`
}

// Media represents media RSS content
type Media struct {
	Content   *MediaContent   `json:"content,omitempty"`
	Thumbnail *MediaThumbnail `json:"thumbnail,omitempty"`
}

// MediaContent represents media content
type MediaContent struct {
	URL  string `json:"url"`
	Type string `json:"type,omitempty"`
}

// MediaThumbnail represents media thumbnail
type MediaThumbnail struct {
	URL string `json:"url"`
}

// Torrent represents torrent metadata
type Torrent struct {
	Link          string    `json:"link"`
	ContentLength int64     `json:"contentLength"`
	PubDate       time.Time `json:"pubDate"`
}
