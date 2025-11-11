package youtube

import (
	"sync"

	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/feed"
)

var (
	googleAPI    *GoogleAPI
	innertubeAPI *InnertubeAPI
	initOnce     sync.Once
)

// initAPIs initializes the API clients (singleton pattern)
func initAPIs() {
	initOnce.Do(func() {
		googleAPI = NewGoogleAPI()
		innertubeAPI = NewInnertubeAPI()
	})
}

// callAPI executes a function with Google API first, falling back to Innertube if needed
func callAPI(
	googleFn func(*GoogleAPI) (*feed.Data, error),
	innertubeFn func(*InnertubeAPI) (*feed.Data, error),
) (*feed.Data, error) {
	initAPIs()

	// Try Google API first if key is available
	if config.C.YouTube.Key != "" && googleAPI != nil {
		data, err := googleFn(googleAPI)
		if err == nil {
			return data, nil
		}
		// If Google API fails, try Innertube fallback
	}

	// Fallback to Innertube API
	return innertubeFn(innertubeAPI)
}
