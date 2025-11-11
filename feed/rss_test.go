package feed

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestGenerateRSS(t *testing.T) {
	now := time.Now()

	data := &Data{
		Title:       "Test Feed",
		Link:        "https://example.com",
		Description: "Test Description",
		Language:    "en",
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
			{
				Title:       "Item 2",
				Link:        "https://example.com/item2",
				Description: "Item 2 description",
				PubDate:     now.Add(-1 * time.Hour),
			},
		},
	}

	output, err := GenerateRSS(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateRSS failed: %v", err)
	}

	// Check XML header
	if !strings.HasPrefix(output, xml.Header) {
		t.Error("Expected XML header")
	}

	// Check it's valid XML
	var rss RSS
	err = xml.Unmarshal([]byte(output), &rss)
	if err != nil {
		t.Fatalf("Invalid XML: %v", err)
	}

	// Verify structure
	if rss.Version != "2.0" {
		t.Errorf("Expected version 2.0, got %s", rss.Version)
	}
	if rss.Channel.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", rss.Channel.Title)
	}
	if len(rss.Channel.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(rss.Channel.Items))
	}

	// Check first item
	item := rss.Channel.Items[0]
	if item.Title != "Item 1" {
		t.Errorf("Expected 'Item 1', got '%s'", item.Title)
	}
	if item.Author != "John Doe" {
		t.Errorf("Expected 'John Doe', got '%s'", item.Author)
	}
	if len(item.Category) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(item.Category))
	}
	if item.GUID.Value != "item1" {
		t.Errorf("Expected GUID 'item1', got '%s'", item.GUID.Value)
	}
}

func TestGenerateRSS_EmptyItems(t *testing.T) {
	data := &Data{
		Title:       "Empty Feed",
		Link:        "https://example.com",
		Description: "No items",
		Item:        []Item{},
	}

	output, err := GenerateRSS(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateRSS failed: %v", err)
	}

	var rss RSS
	err = xml.Unmarshal([]byte(output), &rss)
	if err != nil {
		t.Fatalf("Invalid XML: %v", err)
	}

	if len(rss.Channel.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(rss.Channel.Items))
	}
}

func TestGenerateRSS_WithEnclosure(t *testing.T) {
	data := &Data{
		Title: "Feed with Enclosure",
		Link:  "https://example.com",
		Item: []Item{
			{
				Title:           "Podcast Episode",
				Link:            "https://example.com/episode1",
				EnclosureURL:    "https://example.com/audio.mp3",
				EnclosureType:   "audio/mpeg",
				EnclosureLength: 12345678,
			},
		},
	}

	output, err := GenerateRSS(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateRSS failed: %v", err)
	}

	var rss RSS
	err = xml.Unmarshal([]byte(output), &rss)
	if err != nil {
		t.Fatalf("Invalid XML: %v", err)
	}

	item := rss.Channel.Items[0]
	if item.Enclosure == nil {
		t.Fatal("Expected enclosure, got nil")
	}
	if item.Enclosure.URL != "https://example.com/audio.mp3" {
		t.Errorf("Wrong enclosure URL: %s", item.Enclosure.URL)
	}
	if item.Enclosure.Type != "audio/mpeg" {
		t.Errorf("Wrong enclosure type: %s", item.Enclosure.Type)
	}
	if item.Enclosure.Length != 12345678 {
		t.Errorf("Wrong enclosure length: %d", item.Enclosure.Length)
	}
}

func TestGenerateRSS_WithImage(t *testing.T) {
	data := &Data{
		Title: "Feed with Image",
		Link:  "https://example.com",
		Image: "https://example.com/logo.png",
		Item:  []Item{},
	}

	output, err := GenerateRSS(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateRSS failed: %v", err)
	}

	var rss RSS
	err = xml.Unmarshal([]byte(output), &rss)
	if err != nil {
		t.Fatalf("Invalid XML: %v", err)
	}

	if rss.Channel.Image == nil {
		t.Fatal("Expected image, got nil")
	}
	if rss.Channel.Image.URL != "https://example.com/logo.png" {
		t.Errorf("Wrong image URL: %s", rss.Channel.Image.URL)
	}
}

func TestFormatRFC822(t *testing.T) {
	// Test RFC822 date formatting
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	formatted := formatRFC822(testTime)

	// Should be in RFC1123Z format
	if !strings.Contains(formatted, "2024") {
		t.Errorf("Expected year 2024 in formatted date: %s", formatted)
	}
	if !strings.Contains(formatted, "Jan") {
		t.Errorf("Expected month Jan in formatted date: %s", formatted)
	}
}
