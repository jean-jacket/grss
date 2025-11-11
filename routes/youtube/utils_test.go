package youtube

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParseBoolParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		queryString  string
		paramName    string
		defaultValue bool
		expected     bool
	}{
		// Test default value when parameter not provided
		{"no param - default true", "", "embed", true, true},
		{"no param - default false", "", "embed", false, false},

		// Test explicit true values
		{"explicit true", "embed=true", "embed", false, true},
		{"explicit TRUE", "embed=TRUE", "embed", false, true},
		{"explicit 1", "embed=1", "embed", false, true},
		{"explicit yes", "embed=yes", "embed", false, true},
		{"explicit YES", "embed=YES", "embed", false, true},

		// Test explicit false values
		{"explicit false", "embed=false", "embed", true, false},
		{"explicit FALSE", "embed=FALSE", "embed", true, false},
		{"explicit 0", "embed=0", "embed", true, false},
		{"explicit no", "embed=no", "embed", true, false},
		{"explicit NO", "embed=NO", "embed", true, false},

		// Test invalid values (should default to false for safety)
		{"invalid value", "embed=invalid", "embed", true, false},
		// Empty value is treated same as not provided (uses default)
		{"empty value", "embed=", "embed", true, true},

		// Test filterShorts parameter
		{"filterShorts default", "", "filterShorts", true, true},
		{"filterShorts true", "filterShorts=true", "filterShorts", false, true},
		{"filterShorts false", "filterShorts=false", "filterShorts", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			req := httptest.NewRequest("GET", "/?"+tt.queryString, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Test the function
			result := parseBoolParam(c, tt.paramName, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("parseBoolParam() = %v, expected %v (query: %s, default: %v)",
					result, tt.expected, tt.queryString, tt.defaultValue)
			}
		})
	}
}

func TestIsYouTubeChannelID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid channel ID", "UCDwDMPOZfxVV0x_dz0eQ8KQ", true},
		{"valid with ending A", "UC123456789012345678901A", true},
		{"valid with ending Q", "UC123456789012345678901Q", true},
		{"valid with ending g", "UC123456789012345678901g", true},
		{"valid with ending w", "UC123456789012345678901w", true},
		{"invalid - too short", "UC123", false},
		{"invalid - too long", "UC1234567890123456789012345", false},
		{"invalid - wrong prefix", "UA123456789012345678901A", false},
		{"invalid - wrong ending", "UC123456789012345678901X", false},
		{"invalid - username", "@username", false},
		{"invalid - playlist", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isYouTubeChannelID(tt.input)
			if result != tt.expected {
				t.Errorf("isYouTubeChannelID(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetPlaylistWithShortsFilter(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		filterShorts bool
		expected     string
	}{
		{"UC with filter", "UCDwDMPOZfxVV0x_dz0eQ8KQ", true, "UULFDwDMPOZfxVV0x_dz0eQ8KQ"},
		{"UC without filter", "UCDwDMPOZfxVV0x_dz0eQ8KQ", false, "UCDwDMPOZfxVV0x_dz0eQ8KQ"},
		{"UU with filter", "UUDwDMPOZfxVV0x_dz0eQ8KQ", true, "UULFDwDMPOZfxVV0x_dz0eQ8KQ"},
		{"UU without filter", "UUDwDMPOZfxVV0x_dz0eQ8KQ", false, "UUDwDMPOZfxVV0x_dz0eQ8KQ"},
		{"PL with filter", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", true, "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
		{"PL without filter", "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", false, "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPlaylistWithShortsFilter(tt.id, tt.filterShorts)
			if result != tt.expected {
				t.Errorf("getPlaylistWithShortsFilter(%q, %v) = %q, expected %q",
					tt.id, tt.filterShorts, result, tt.expected)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		expected int
	}{
		{"empty", "", 0},
		{"15 seconds", "PT15S", 15},
		{"1 minute 30 seconds", "PT1M30S", 90},
		{"1 hour", "PT1H", 3600},
		{"1 hour 2 minutes 3 seconds", "PT1H2M3S", 3723},
		{"2 hours 15 seconds", "PT2H15S", 7215},
		{"30 minutes", "PT30M", 1800},
		{"10 hours", "PT10H", 36000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("parseDuration(%q) = %d, expected %d", tt.duration, result, tt.expected)
			}
		})
	}
}
