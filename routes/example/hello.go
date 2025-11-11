package example

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
)

// HelloRoute defines a simple hello world route
var HelloRoute = registry.Route{
	Path:        "/hello",
	Name:        "Hello World",
	Maintainers: []string{"example"},
	Example:     "/example/hello",
	Description: "A simple hello world feed",
	Handler:     helloHandler,
}

func helloHandler(c *gin.Context) (*feed.Data, error) {
	now := time.Now()

	return &feed.Data{
		Title:       "Hello World Feed",
		Link:        "https://example.com",
		Description: "This is an example feed",
		Language:    "en",
		PubDate:     now,
		Item: []feed.Item{
			{
				Title:       "Hello, GRSS!",
				Link:        "https://example.com/hello",
				Description: "Welcome to the GRSS feed aggregation system",
				PubDate:     now,
				Author:      "GRSS Team",
				Category:    []string{"demo", "example"},
			},
			{
				Title:       "Second Item",
				Link:        "https://example.com/second",
				Description: "This demonstrates multiple items in a feed",
				PubDate:     now.Add(-1 * time.Hour),
				Author:      "GRSS Team",
			},
		},
	}, nil
}
