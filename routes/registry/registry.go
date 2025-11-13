package registry

import (
	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/feed"
)

// Route defines a route handler and metadata
type Route struct {
	// Route configuration
	Path        string
	Name        string
	Maintainers []string
	Example     string

	// Route handler
	Handler RouteHandler

	// Optional metadata
	Parameters  map[string]interface{}
	Description string
	Categories  []string
	Features    *Features
}

// RouteHandler is the function signature for route handlers
type RouteHandler func(c *gin.Context) (*feed.Data, error)

// Features defines route features and requirements
type Features struct {
	RequireConfig     []ConfigRequirement
	RequirePuppeteer  bool
	AntiCrawler       bool
	SupportBT         bool
	SupportScihub     bool
}

// ConfigRequirement defines a required configuration key
type ConfigRequirement struct {
	Name     string
	Optional bool
}

// Namespace defines a route namespace
type Namespace struct {
	Name        string
	URL         string
	Description string
	Lang        string
	Categories  []string
	Routes      []Route
}

// Registry holds all registered namespaces and routes
type Registry struct {
	namespaces map[string]*Namespace
}

// NewRegistry creates a new route registry
func NewRegistry() *Registry {
	return &Registry{
		namespaces: make(map[string]*Namespace),
	}
}

// RegisterNamespace registers a new namespace
func (r *Registry) RegisterNamespace(name string, ns *Namespace) {
	r.namespaces[name] = ns
}

// RegisterRoute registers a route under a namespace
func (r *Registry) RegisterRoute(namespaceName string, route Route) {
	ns, exists := r.namespaces[namespaceName]
	if !exists {
		ns = &Namespace{
			Name:   namespaceName,
			Routes: []Route{},
		}
		r.namespaces[namespaceName] = ns
	}

	ns.Routes = append(ns.Routes, route)
}

// MountRoutes mounts all registered routes to the Gin router
func (r *Registry) MountRoutes(router *gin.Engine) {
	for namespaceName, namespace := range r.namespaces {
		for _, route := range namespace.Routes {
			path := "/" + namespaceName + route.Path
			router.GET(path, wrapHandler(route.Handler))
		}
	}
}

// wrapHandler wraps a RouteHandler to work with Gin
func wrapHandler(handler RouteHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute handler
		data, err := handler(c)
		if err != nil {
			// Error handling - return error response
			c.JSON(500, gin.H{
				"error": gin.H{
					"message": err.Error(),
				},
			})
			c.Abort()
			return
		}

		// Set data in context for template middleware
		c.Set("feed_data", data)
	}
}

// GetNamespaces returns all registered namespaces
func (r *Registry) GetNamespaces() map[string]*Namespace {
	return r.namespaces
}

// Global registry instance
var DefaultRegistry = NewRegistry()

// RegisterNamespace registers a namespace in the default registry
func RegisterNamespace(name string, ns *Namespace) {
	DefaultRegistry.RegisterNamespace(name, ns)
}

// RegisterRoute registers a route in the default registry
func RegisterRoute(namespaceName string, route Route) {
	DefaultRegistry.RegisterRoute(namespaceName, route)
}

// MountRoutes mounts all routes from the default registry
func MountRoutes(router *gin.Engine) {
	DefaultRegistry.MountRoutes(router)
}

// GetAllRoutes returns a flat list of all routes with their full paths
func (r *Registry) GetAllRoutes() map[string]RouteInfo {
	routes := make(map[string]RouteInfo)
	for namespaceName, namespace := range r.namespaces {
		for _, route := range namespace.Routes {
			path := "/" + namespaceName + route.Path
			routes[path] = RouteInfo{
				Path:      path,
				Namespace: namespaceName,
				Route:     route,
			}
		}
	}
	return routes
}

// RouteInfo contains route path and handler information
type RouteInfo struct {
	Path      string
	Namespace string
	Route     Route
}

// GetAllRoutes returns all routes from the default registry
func GetAllRoutes() map[string]RouteInfo {
	return DefaultRegistry.GetAllRoutes()
}
