package wubby

import "github.com/jean-jacket/grss/routes/registry"

// Namespace defines the Wubby namespace
var Namespace = &registry.Namespace{
	Name:        "Wubby",
	URL:         "https://parasoci.al/",
	Description: "Wubby VOD routes",
	Lang:        "en",
	Categories:  []string{"video", "livestream"},
}
