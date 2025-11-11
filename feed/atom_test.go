package feed

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestGenerateAtom(t *testing.T) {
	now := time.Now()

	data := &Data{
		Title:       "Test Feed",
		Link:        "https://example.com",
		Description: "Test Description",
		Subtitle:    "Test Subtitle",
		Icon:        "https://example.com/icon.png",
		Author:      "Jane Doe",
		PubDate:     now,
		Item: []Item{
			{
				Title:       "Item 1",
				Link:        "https://example.com/item1",
				Description: "Item 1 description",
				PubDate:     now,
				Author:      "John Doe",
				Category:    []string{"tech", "golang"},
				GUID:        "item1",
			},
		},
	}

	output, err := GenerateAtom(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateAtom failed: %v", err)
	}

	// Check XML header
	if !strings.HasPrefix(output, xml.Header) {
		t.Error("Expected XML header")
	}

	// Check it's valid XML
	var atom AtomFeed
	err = xml.Unmarshal([]byte(output), &atom)
	if err != nil {
		t.Fatalf("Invalid XML: %v", err)
	}

	// Verify structure
	if atom.Xmlns != "http://www.w3.org/2005/Atom" {
		t.Errorf("Wrong xmlns: %s", atom.Xmlns)
	}
	if atom.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", atom.Title)
	}
	if atom.Subtitle != "Test Subtitle" {
		t.Errorf("Expected subtitle 'Test Subtitle', got '%s'", atom.Subtitle)
	}
	if atom.Icon != "https://example.com/icon.png" {
		t.Errorf("Wrong icon: %s", atom.Icon)
	}
	if atom.Author == nil || atom.Author.Name != "Jane Doe" {
		t.Error("Expected author 'Jane Doe'")
	}
	if len(atom.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(atom.Entries))
	}

	// Check first entry
	entry := atom.Entries[0]
	if entry.Title != "Item 1" {
		t.Errorf("Expected 'Item 1', got '%s'", entry.Title)
	}
	if entry.ID != "item1" {
		t.Errorf("Expected ID 'item1', got '%s'", entry.ID)
	}
	if entry.Author == nil || entry.Author.Name != "John Doe" {
		t.Error("Expected author 'John Doe'")
	}
	if len(entry.Category) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(entry.Category))
	}
	if entry.Content == nil || entry.Content.Type != "html" {
		t.Error("Expected HTML content")
	}
}

func TestGenerateAtom_Links(t *testing.T) {
	data := &Data{
		Title: "Test Feed",
		Link:  "https://example.com",
		Item:  []Item{},
	}

	output, err := GenerateAtom(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateAtom failed: %v", err)
	}

	var atom AtomFeed
	err = xml.Unmarshal([]byte(output), &atom)
	if err != nil {
		t.Fatalf("Invalid XML: %v", err)
	}

	// Check links
	if len(atom.Link) < 2 {
		t.Fatal("Expected at least 2 links")
	}

	// Check alternate link
	var foundAlternate, foundSelf bool
	for _, link := range atom.Link {
		if link.Rel == "alternate" && link.Href == "https://example.com" {
			foundAlternate = true
		}
		if link.Rel == "self" && link.Href == "https://example.com/feed" {
			foundSelf = true
		}
	}

	if !foundAlternate {
		t.Error("Missing alternate link")
	}
	if !foundSelf {
		t.Error("Missing self link")
	}
}

func TestGenerateAtom_DefaultID(t *testing.T) {
	// Test that items without GUID use link as ID
	data := &Data{
		Title: "Test Feed",
		Link:  "https://example.com",
		Item: []Item{
			{
				Title:       "Item without GUID",
				Link:        "https://example.com/item1",
				Description: "Description",
			},
		},
	}

	output, err := GenerateAtom(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateAtom failed: %v", err)
	}

	var atom AtomFeed
	err = xml.Unmarshal([]byte(output), &atom)
	if err != nil {
		t.Fatalf("Invalid XML: %v", err)
	}

	if atom.Entries[0].ID != "https://example.com/item1" {
		t.Errorf("Expected ID to be link, got '%s'", atom.Entries[0].ID)
	}
}

func TestFormatRFC3339(t *testing.T) {
	// Test RFC3339 date formatting
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	formatted := formatRFC3339(testTime)

	// Should be in RFC3339 format
	expected := "2024-01-01T12:00:00Z"
	if formatted != expected {
		t.Errorf("Expected '%s', got '%s'", expected, formatted)
	}
}
