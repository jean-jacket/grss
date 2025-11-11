package youtube

import "time"

// Thumbnail represents a YouTube thumbnail
type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// ThumbnailSet represents a set of thumbnails at different resolutions
type ThumbnailSet struct {
	Default  *Thumbnail `json:"default,omitempty"`
	Medium   *Thumbnail `json:"medium,omitempty"`
	High     *Thumbnail `json:"high,omitempty"`
	Standard *Thumbnail `json:"standard,omitempty"`
	Maxres   *Thumbnail `json:"maxres,omitempty"`
}

// Video represents basic video information
type Video struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	PublishedAt  time.Time    `json:"publishedAt"`
	Thumbnails   ThumbnailSet `json:"thumbnails"`
	ChannelTitle string       `json:"channelTitle"`
	Duration     string       `json:"duration,omitempty"` // ISO 8601 format
}

// Channel represents YouTube channel information
type Channel struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Thumbnails  ThumbnailSet `json:"thumbnails"`
}

// Playlist represents YouTube playlist information
type Playlist struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Thumbnails  ThumbnailSet `json:"thumbnails"`
	ChannelID   string       `json:"channelId"`
	ChannelTitle string      `json:"channelTitle"`
}

// PlaylistItem represents an item in a playlist
type PlaylistItem struct {
	VideoID      string       `json:"videoId"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	PublishedAt  time.Time    `json:"publishedAt"`
	Thumbnails   ThumbnailSet `json:"thumbnails"`
	ChannelTitle string       `json:"channelTitle"`
}
