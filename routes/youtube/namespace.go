package youtube

import "github.com/jean-jacket/grss/routes/registry"

// Namespace defines the YouTube namespace
var Namespace = &registry.Namespace{
	Name:        "YouTube",
	URL:         "https://www.youtube.com",
	Description: "YouTube videos, channels, playlists, and more",
	Lang:        "en",
	Categories:  []string{"social-media", "multimedia"},
}
