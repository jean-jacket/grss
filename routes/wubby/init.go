package wubby

import "github.com/jean-jacket/grss/routes/registry"

// init registers all wubby routes
func init() {
	registry.RegisterNamespace("wubby", Namespace)
	registry.RegisterRoute("wubby", VodsRoute)
}
