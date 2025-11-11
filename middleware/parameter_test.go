package middleware

import (
	"testing"
	"time"

	"github.com/jean-jacket/grss/feed"
)

func TestFilterItems(t *testing.T) {
	items := []feed.Item{
		{Title: "Bug fix in parser", Description: "Fixed a critical bug"},
		{Title: "New feature added", Description: "Added new functionality"},
		{Title: "Bug in authentication", Description: "Security issue"},
	}

	// Filter for "bug"
	filtered := filterItems(items, "(?i)bug", false)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 items with 'bug', got %d", len(filtered))
	}

	// Filter out "bug"
	filtered = filterItems(items, "(?i)bug", true)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 item without 'bug', got %d", len(filtered))
	}
	if filtered[0].Title != "New feature added" {
		t.Errorf("Wrong item filtered: %s", filtered[0].Title)
	}
}

func TestFilterByTitle(t *testing.T) {
	items := []feed.Item{
		{Title: "Bug fix", Description: "Some bug description"},
		{Title: "Feature", Description: "Bug mentioned in description"},
		{Title: "Bug report", Description: "Another description"},
	}

	// Filter by title only
	filtered := filterByTitle(items, "(?i)bug", false)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 items, got %d", len(filtered))
	}

	// Should not match description
	for _, item := range filtered {
		if item.Title == "Feature" {
			t.Error("Should not match description when filtering by title")
		}
	}
}

func TestFilterByDescription(t *testing.T) {
	items := []feed.Item{
		{Title: "Title with bug", Description: "Normal description"},
		{Title: "Normal title", Description: "Description with bug"},
	}

	filtered := filterByDescription(items, "(?i)bug", false)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 item, got %d", len(filtered))
	}
	if filtered[0].Title != "Normal title" {
		t.Errorf("Wrong item filtered: %s", filtered[0].Title)
	}
}

func TestFilterByTime(t *testing.T) {
	now := time.Now()

	items := []feed.Item{
		{Title: "Recent", PubDate: now.Add(-1 * time.Minute)},
		{Title: "Old", PubDate: now.Add(-2 * time.Hour)},
		{Title: "Very old", PubDate: now.Add(-1 * time.Hour)},
	}

	// Filter for items within last hour
	filtered := filterByTime(items, "3600")
	if len(filtered) != 1 {
		t.Errorf("Expected 1 item within last hour, got %d", len(filtered))
	}
	if filtered[0].Title != "Recent" {
		t.Errorf("Wrong item: %s", filtered[0].Title)
	}
}

func TestSortItems(t *testing.T) {
	now := time.Now()

	items := []feed.Item{
		{Title: "Middle", PubDate: now},
		{Title: "Newest", PubDate: now.Add(1 * time.Hour)},
		{Title: "Oldest", PubDate: now.Add(-1 * time.Hour)},
	}

	// Sort ascending
	sorted := sortItems(items, "asc")
	if sorted[0].Title != "Oldest" {
		t.Errorf("Expected 'Oldest' first in asc sort, got '%s'", sorted[0].Title)
	}
	if sorted[2].Title != "Newest" {
		t.Errorf("Expected 'Newest' last in asc sort, got '%s'", sorted[2].Title)
	}

	// Sort descending
	sorted = sortItems(items, "desc")
	if sorted[0].Title != "Newest" {
		t.Errorf("Expected 'Newest' first in desc sort, got '%s'", sorted[0].Title)
	}
	if sorted[2].Title != "Oldest" {
		t.Errorf("Expected 'Oldest' last in desc sort, got '%s'", sorted[2].Title)
	}
}

func TestFilterItems_InvalidRegex(t *testing.T) {
	items := []feed.Item{
		{Title: "Test"},
	}

	// Invalid regex should return all items unchanged
	filtered := filterItems(items, "[invalid(regex", false)
	if len(filtered) != len(items) {
		t.Error("Invalid regex should return all items")
	}
}

func TestFilterByTime_InvalidTime(t *testing.T) {
	items := []feed.Item{
		{Title: "Test"},
	}

	// Invalid time should return all items unchanged
	filtered := filterByTime(items, "invalid")
	if len(filtered) != len(items) {
		t.Error("Invalid time should return all items")
	}
}
