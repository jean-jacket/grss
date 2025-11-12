package apple

import "github.com/jean-jacket/grss/routes/registry"

// Namespace defines the Apple namespace
var Namespace = &registry.Namespace{
	Name:        "Apple Developer Design",
	URL:         "https://developer.apple.com/design",
	Description: "Apple Developer Design resources, guidelines, and updates",
	Lang:        "en",
	Categories:  []string{"design", "development", "apple"},
}
