package youtube

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
)

// ChannelRoute defines the YouTube channel route
var ChannelRoute = registry.Route{
	Path:        "/channel/:id",
	Name:        "Channel Videos",
	Maintainers: []string{"grss"},
	Example:     "/youtube/channel/UCDwDMPOZfxVV0x_dz0eQ8KQ",
	Description: "Get videos from a YouTube channel by channel ID. Supports optional parameters: embed (default: true), filterShorts (default: true)",
	Parameters: map[string]interface{}{
		"id": "YouTube channel ID (must start with UC)",
	},
	Handler: channelHandler,
}

func channelHandler(c *gin.Context) (*feed.Data, error) {
	channelID := c.Param("id")

	// Validate channel ID
	if !isYouTubeChannelID(channelID) {
		return nil, fmt.Errorf("invalid YouTube channel ID: %s. Channel IDs must start with UC and be 24 characters long. You may want to use /youtube/user/:username instead", channelID)
	}

	// Parse parameters with defaults
	embed := parseBoolParam(c, "embed", true)
	filterShorts := parseBoolParam(c, "filterShorts", true)

	// Call API with fallback logic
	return callAPI(
		func(g *GoogleAPI) (*feed.Data, error) {
			return g.getDataByChannelID(channelID, embed, filterShorts)
		},
		func(i *InnertubeAPI) (*feed.Data, error) {
			return i.getDataByChannelID(channelID, embed, filterShorts)
		},
	)
}
