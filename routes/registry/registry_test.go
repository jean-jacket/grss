package registry

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/feed"
)

func TestRegistry_RegisterNamespace(t *testing.T) {
	reg := NewRegistry()

	ns := &Namespace{
		Name:        "Test Namespace",
		URL:         "https://test.com",
		Description: "Test description",
	}

	reg.RegisterNamespace("test", ns)

	namespaces := reg.GetNamespaces()
	if len(namespaces) != 1 {
		t.Errorf("Expected 1 namespace, got %d", len(namespaces))
	}

	retrieved := namespaces["test"]
	if retrieved == nil {
		t.Fatal("Namespace not found")
	}
	if retrieved.Name != "Test Namespace" {
		t.Errorf("Expected 'Test Namespace', got '%s'", retrieved.Name)
	}
}

func TestRegistry_RegisterRoute(t *testing.T) {
	reg := NewRegistry()

	route := Route{
		Path:        "/test/:id",
		Name:        "Test Route",
		Maintainers: []string{"test"},
		Example:     "/test/123",
		Handler: func(c *gin.Context) (*feed.Data, error) {
			return &feed.Data{
				Title: "Test",
				Link:  "https://test.com",
			}, nil
		},
	}

	reg.RegisterRoute("test", route)

	namespaces := reg.GetNamespaces()
	ns := namespaces["test"]
	if ns == nil {
		t.Fatal("Namespace should be auto-created")
	}
	if len(ns.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(ns.Routes))
	}
	if ns.Routes[0].Name != "Test Route" {
		t.Errorf("Expected 'Test Route', got '%s'", ns.Routes[0].Name)
	}
}

func TestRegistry_MountRoutes(t *testing.T) {
	reg := NewRegistry()

	// Register test route
	route := Route{
		Path:    "/test",
		Name:    "Test Route",
		Handler: testRouteHandler,
	}
	reg.RegisterRoute("test", route)

	// Create Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	reg.MountRoutes(router)

	// Test the route
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestWrapHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context) (*feed.Data, error) {
		return &feed.Data{
			Title: "Success",
			Link:  "https://example.com",
		}, nil
	}

	wrapped := wrapHandler(handler)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	wrapped(c)

	// Check that data was set in context
	dataInterface, exists := c.Get("feed_data")
	if !exists {
		t.Fatal("Expected feed_data in context")
	}

	data, ok := dataInterface.(*feed.Data)
	if !ok {
		t.Fatal("Expected *feed.Data type")
	}

	if data.Title != "Success" {
		t.Errorf("Expected title 'Success', got '%s'", data.Title)
	}
}

func TestWrapHandler_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := func(c *gin.Context) (*feed.Data, error) {
		return nil, errors.New("test error")
	}

	wrapped := wrapHandler(handler)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	wrapped(c)

	// Should return error response
	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Test that default registry exists and works
	route := Route{
		Path:    "/global-test",
		Name:    "Global Test",
		Handler: testRouteHandler,
	}

	RegisterRoute("global", route)

	namespaces := DefaultRegistry.GetNamespaces()
	if namespaces["global"] == nil {
		t.Error("Expected global namespace in default registry")
	}
}

// Helper function for tests
func testRouteHandler(c *gin.Context) (*feed.Data, error) {
	return &feed.Data{
		Title: "Test Feed",
		Link:  "https://test.com",
		Item:  []feed.Item{},
	}, nil
}
