package example

import "github.com/jean-jacket/grss/routes/registry"

// Namespace defines the Example namespace
var Namespace = &registry.Namespace{
	Name:        "Example",
	URL:         "https://example.com",
	Description: "Example routes demonstrating the route system",
	Lang:        "en",
	Categories:  []string{"demo"},
}
