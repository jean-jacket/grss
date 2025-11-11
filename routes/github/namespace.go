package github

import "github.com/jean-jacket/grss/routes/registry"

// Namespace defines the GitHub namespace
var Namespace = &registry.Namespace{
	Name:        "GitHub",
	URL:         "https://github.com",
	Description: "GitHub repositories, issues, releases, and more",
	Lang:        "en",
	Categories:  []string{"programming"},
}
