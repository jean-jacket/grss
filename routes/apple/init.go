package apple

import "github.com/jean-jacket/grss/routes/registry"

// init registers all Apple routes
func init() {
	registry.RegisterNamespace("apple", Namespace)
	registry.RegisterRoute("apple", DesignUpdatesRoute)
}
