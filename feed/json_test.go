package feed

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGenerateJSON(t *testing.T) {
	now := time.Now()

	data := &Data{
		Title:       "Test Feed",
		Link:        "https://example.com",
		Description: "Test Description",
		Language:    "en",
		Icon:        "https://example.com/icon.png",
		Author:      "Jane Doe",
		Item: []Item{
			{
				Title:       "Item 1",
				Link:        "https://example.com/item1",
				Description: "Item 1 description",
				PubDate:     now,
				Updated:     now,
				Author:      "John Doe",
				Category:    []string{"tech", "golang"},
				GUID:        "item1",
			},
		},
	}

	output, err := GenerateJSON(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	// Check it's valid JSON
	var jsonFeed JSONFeed
	err = json.Unmarshal([]byte(output), &jsonFeed)
	if err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify structure
	if jsonFeed.Version != "https://jsonfeed.org/version/1.1" {
		t.Errorf("Wrong version: %s", jsonFeed.Version)
	}
	if jsonFeed.Title != "Test Feed" {
		t.Errorf("Expected title 'Test Feed', got '%s'", jsonFeed.Title)
	}
	if jsonFeed.HomePageURL != "https://example.com" {
		t.Errorf("Wrong home page URL: %s", jsonFeed.HomePageURL)
	}
	if jsonFeed.FeedURL != "https://example.com/feed" {
		t.Errorf("Wrong feed URL: %s", jsonFeed.FeedURL)
	}
	if jsonFeed.Description != "Test Description" {
		t.Errorf("Wrong description: %s", jsonFeed.Description)
	}
	if jsonFeed.Icon != "https://example.com/icon.png" {
		t.Errorf("Wrong icon: %s", jsonFeed.Icon)
	}
	if len(jsonFeed.Authors) != 1 || jsonFeed.Authors[0].Name != "Jane Doe" {
		t.Error("Expected author 'Jane Doe'")
	}
	if len(jsonFeed.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(jsonFeed.Items))
	}

	// Check first item
	item := jsonFeed.Items[0]
	if item.ID != "item1" {
		t.Errorf("Expected ID 'item1', got '%s'", item.ID)
	}
	if item.URL != "https://example.com/item1" {
		t.Errorf("Wrong URL: %s", item.URL)
	}
	if item.Title != "Item 1" {
		t.Errorf("Expected title 'Item 1', got '%s'", item.Title)
	}
	if item.ContentHTML != "Item 1 description" {
		t.Errorf("Wrong content: %s", item.ContentHTML)
	}
	if len(item.Authors) != 1 || item.Authors[0].Name != "John Doe" {
		t.Error("Expected author 'John Doe'")
	}
	if len(item.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(item.Tags))
	}
}

func TestGenerateJSON_WithAttachments(t *testing.T) {
	data := &Data{
		Title: "Feed with Attachments",
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

	output, err := GenerateJSON(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	var jsonFeed JSONFeed
	err = json.Unmarshal([]byte(output), &jsonFeed)
	if err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	item := jsonFeed.Items[0]
	if len(item.Attachments) != 1 {
		t.Fatalf("Expected 1 attachment, got %d", len(item.Attachments))
	}

	attachment := item.Attachments[0]
	if attachment.URL != "https://example.com/audio.mp3" {
		t.Errorf("Wrong attachment URL: %s", attachment.URL)
	}
	if attachment.MIMEType != "audio/mpeg" {
		t.Errorf("Wrong MIME type: %s", attachment.MIMEType)
	}
	if attachment.SizeInBytes != 12345678 {
		t.Errorf("Wrong size: %d", attachment.SizeInBytes)
	}
}

func TestGenerateJSON_DefaultID(t *testing.T) {
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

	output, err := GenerateJSON(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	var jsonFeed JSONFeed
	err = json.Unmarshal([]byte(output), &jsonFeed)
	if err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if jsonFeed.Items[0].ID != "https://example.com/item1" {
		t.Errorf("Expected ID to be link, got '%s'", jsonFeed.Items[0].ID)
	}
}

func TestGenerateJSON_EmptyItems(t *testing.T) {
	data := &Data{
		Title: "Empty Feed",
		Link:  "https://example.com",
		Item:  []Item{},
	}

	output, err := GenerateJSON(data, "https://example.com/feed")
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	var jsonFeed JSONFeed
	err = json.Unmarshal([]byte(output), &jsonFeed)
	if err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if len(jsonFeed.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(jsonFeed.Items))
	}
}
