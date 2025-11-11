package github

import "github.com/jean-jacket/grss/routes/registry"

// init registers all GitHub routes
func init() {
	registry.RegisterNamespace("github", Namespace)
	registry.RegisterRoute("github", IssuesRoute)
}
