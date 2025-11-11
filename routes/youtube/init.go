package youtube

import "github.com/jean-jacket/grss/routes/registry"

// init registers all YouTube routes
func init() {
	registry.RegisterNamespace("youtube", Namespace)
	registry.RegisterRoute("youtube", ChannelRoute)
	registry.RegisterRoute("youtube", UserRoute)
	registry.RegisterRoute("youtube", PlaylistRoute)
}
