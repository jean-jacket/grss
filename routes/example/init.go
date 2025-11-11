package example

import "github.com/jean-jacket/grss/routes/registry"

// init registers all example routes
func init() {
	registry.RegisterNamespace("example", Namespace)
	registry.RegisterRoute("example", HelloRoute)
}
