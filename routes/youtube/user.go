package youtube

import (
	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
)

// UserRoute defines the YouTube user/handle route
var UserRoute = registry.Route{
	Path:        "/user/:username",
	Name:        "User/Handle Videos",
	Maintainers: []string{"grss"},
	Example:     "/youtube/user/@JFlaMusic",
	Description: "Get videos from a YouTube channel by username or handle (e.g., @username). Supports optional parameters: embed (default: true), filterShorts (default: true)",
	Parameters: map[string]interface{}{
		"username": "YouTube username or handle (with @ prefix for handles)",
	},
	Handler: userHandler,
}

func userHandler(c *gin.Context) (*feed.Data, error) {
	username := c.Param("username")

	// Parse embed parameter (default: true)
	embedParam := c.Query("embed")
	embed := embedParam == "" || embedParam == "true"

	// Parse filterShorts parameter (default: true)
	filterShortsParam := c.Query("filterShorts")
	filterShorts := filterShortsParam == "" || filterShortsParam == "true"

	// Call API with fallback logic
	return callAPI(
		func(g *GoogleAPI) (*feed.Data, error) {
			return g.getDataByUsername(username, embed, filterShorts)
		},
		func(i *InnertubeAPI) (*feed.Data, error) {
			return i.getDataByUsername(username, embed, filterShorts)
		},
	)
}
