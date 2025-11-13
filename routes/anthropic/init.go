package anthropic

import "github.com/jean-jacket/grss/routes/registry"

// init registers all Anthropic routes
func init() {
	registry.RegisterNamespace("anthropic", Namespace)
	registry.RegisterRoute("anthropic", NewsRoute)
}
