package youtube

import (
	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
)

// PlaylistRoute defines the YouTube playlist route
var PlaylistRoute = registry.Route{
	Path:        "/playlist/:id",
	Name:        "Playlist Videos",
	Maintainers: []string{"grss"},
	Example:     "/youtube/playlist/PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
	Description: "Get videos from a YouTube playlist. Supports optional parameter: embed (default: true)",
	Parameters: map[string]interface{}{
		"id": "YouTube playlist ID",
	},
	Handler: playlistHandler,
}

func playlistHandler(c *gin.Context) (*feed.Data, error) {
	playlistID := c.Param("id")

	// Parse embed parameter (default: true)
	embedParam := c.Query("embed")
	embed := embedParam == "" || embedParam == "true"

	// Call API with fallback logic
	return callAPI(
		func(g *GoogleAPI) (*feed.Data, error) {
			return g.getDataByPlaylistID(playlistID, embed)
		},
		func(i *InnertubeAPI) (*feed.Data, error) {
			return i.getDataByPlaylistID(playlistID, embed)
		},
	)
}
