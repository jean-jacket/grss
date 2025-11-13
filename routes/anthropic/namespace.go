package anthropic

import "github.com/jean-jacket/grss/routes/registry"

// Namespace defines the Anthropic namespace
var Namespace = &registry.Namespace{
	Name:        "Anthropic",
	URL:         "https://www.anthropic.com",
	Description: "Anthropic news, research, and announcements",
	Lang:        "en",
	Categories:  []string{"ai", "research", "technology"},
}
