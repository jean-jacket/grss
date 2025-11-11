package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/feed"
)

// GoogleAPI wraps the YouTube Data API v3 client with key rotation
type GoogleAPI struct {
	keys    []string
	current int
	mu      sync.Mutex
	client  *http.Client
}

// NewGoogleAPI creates a new Google API client with key rotation
func NewGoogleAPI() *GoogleAPI {
	if config.C.YouTube.Key == "" {
		return nil
	}

	keys := strings.Split(config.C.YouTube.Key, ",")
	validKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			validKeys = append(validKeys, key)
		}
	}

	if len(validKeys) == 0 {
		return nil
	}

	return &GoogleAPI{
		keys:    validKeys,
		current: 0,
		client:  &http.Client{},
	}
}

// getKey returns the current API key and rotates to the next one
func (g *GoogleAPI) getKey() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	key := g.keys[g.current%len(g.keys)]
	g.current++
	return key
}

// executeWithRetry executes a function with automatic key rotation on failure
func (g *GoogleAPI) executeWithRetry(fn func(key string) error) error {
	var lastErr error
	attempts := len(g.keys)

	for i := 0; i < attempts; i++ {
		key := g.getKey()
		err := fn(key)
		if err == nil {
			return nil
		}
		lastErr = err
	}

	return lastErr
}

// getChannelByID fetches channel information by channel ID
func (g *GoogleAPI) getChannelByID(channelID string, part string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := g.executeWithRetry(func(key string) error {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=%s&id=%s&key=%s",
			part, channelID, key)

		resp, err := g.client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
		}

		return json.Unmarshal(body, &result)
	})

	return result, err
}

// getChannelByUsername fetches channel information by username
func (g *GoogleAPI) getChannelByUsername(username string, part string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := g.executeWithRetry(func(key string) error {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/channels?part=%s&forUsername=%s&key=%s",
			part, username, key)

		resp, err := g.client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
		}

		return json.Unmarshal(body, &result)
	})

	return result, err
}

// getPlaylist fetches playlist information
func (g *GoogleAPI) getPlaylist(playlistID string, part string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := g.executeWithRetry(func(key string) error {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlists?part=%s&id=%s&key=%s",
			part, playlistID, key)

		resp, err := g.client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
		}

		return json.Unmarshal(body, &result)
	})

	return result, err
}

// getPlaylistItems fetches playlist items
func (g *GoogleAPI) getPlaylistItems(playlistID string, part string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := g.executeWithRetry(func(key string) error {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/playlistItems?part=%s&playlistId=%s&maxResults=50&key=%s",
			part, playlistID, key)

		resp, err := g.client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
		}

		return json.Unmarshal(body, &result)
	})

	return result, err
}

// getVideos fetches video information (supports comma-separated IDs)
func (g *GoogleAPI) getVideos(videoIDs string, part string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := g.executeWithRetry(func(key string) error {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=%s&id=%s&key=%s",
			part, videoIDs, key)

		resp, err := g.client.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
		}

		return json.Unmarshal(body, &result)
	})

	return result, err
}

// getDataByChannelID fetches feed data for a channel by ID using Google API
func (g *GoogleAPI) getDataByChannelID(channelID string, embed bool, filterShorts bool) (*feed.Data, error) {
	// Determine playlist ID
	var playlistID string

	if filterShorts {
		playlistID = getPlaylistWithShortsFilter(channelID, true)
	} else {
		// Fetch channel to get uploads playlist
		channelData, err := g.getChannelByID(channelID, "contentDetails")
		if err != nil {
			return nil, err
		}

		items, ok := channelData["items"].([]interface{})
		if !ok || len(items) == 0 {
			return nil, errors.New("channel not found")
		}

		item := items[0].(map[string]interface{})
		contentDetails := item["contentDetails"].(map[string]interface{})
		relatedPlaylists := contentDetails["relatedPlaylists"].(map[string]interface{})
		playlistID = relatedPlaylists["uploads"].(string)
	}

	// Fetch playlist items
	playlistData, err := g.getPlaylistItems(playlistID, "snippet")
	if err != nil {
		return nil, err
	}

	items, ok := playlistData["items"].([]interface{})
	if !ok {
		return nil, errors.New("invalid playlist data")
	}

	// Extract video IDs
	videoIDs := make([]string, 0, len(items))
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		snippet := itemMap["snippet"].(map[string]interface{})

		// Skip private/deleted videos
		title := snippet["title"].(string)
		if title == "Private video" || title == "Deleted video" {
			continue
		}

		resourceID := snippet["resourceId"].(map[string]interface{})
		videoID := resourceID["videoId"].(string)
		videoIDs = append(videoIDs, videoID)
	}

	if len(videoIDs) == 0 {
		return &feed.Data{
			Title: fmt.Sprintf("Channel %s - YouTube", channelID),
			Link:  fmt.Sprintf("https://www.youtube.com/channel/%s", channelID),
			Item:  []feed.Item{},
		}, nil
	}

	// Fetch video details for durations
	videosData, err := g.getVideos(strings.Join(videoIDs, ","), "contentDetails")
	if err != nil {
		return nil, err
	}

	videoDetailsMap := make(map[string]map[string]interface{})
	if videosData != nil {
		if videoItems, ok := videosData["items"].([]interface{}); ok {
			for _, videoItem := range videoItems {
				videoMap := videoItem.(map[string]interface{})
				id := videoMap["id"].(string)
				videoDetailsMap[id] = videoMap
			}
		}
	}

	// Build feed items
	feedItems := make([]feed.Item, 0, len(items))
	for _, item := range items {
		itemMap := item.(map[string]interface{})
		snippet := itemMap["snippet"].(map[string]interface{})

		title := snippet["title"].(string)
		if title == "Private video" || title == "Deleted video" {
			continue
		}

		resourceID := snippet["resourceId"].(map[string]interface{})
		videoID := resourceID["videoId"].(string)

		// Get thumbnail
		thumbnails := parseThumbnails(snippet["thumbnails"])
		img := getThumbnail(thumbnails)

		// Format description
		description := ""
		if desc, ok := snippet["description"].(string); ok {
			description = formatDescription(desc)
		}

		// Render description with embed/thumbnail
		renderedDesc := renderDescription(embed, videoID, img, description)

		// Parse publish date
		pubDateStr := snippet["publishedAt"].(string)
		pubDate, _ := parseISO8601(pubDateStr)

		// Get duration if available
		var durationSeconds int
		if videoDetail, ok := videoDetailsMap[videoID]; ok {
			if contentDetails, ok := videoDetail["contentDetails"].(map[string]interface{}); ok {
				if duration, ok := contentDetails["duration"].(string); ok {
					durationSeconds = parseDuration(duration)
				}
			}
		}

		// Get author
		author := ""
		if channelTitle, ok := snippet["videoOwnerChannelTitle"].(string); ok {
			author = channelTitle
		}

		feedItem := feed.Item{
			Title:       title,
			Description: renderedDesc,
			Link:        fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
			PubDate:     pubDate,
			Author:      author,
		}

		// Add media enclosure
		if img != nil {
			feedItem.Media = &feed.Media{
				Thumbnail: &feed.MediaThumbnail{
					URL: img.URL,
				},
			}
		}

		// Add video URL as enclosure
		if durationSeconds > 0 {
			feedItem.EnclosureURL = getVideoURL(videoID)
			feedItem.EnclosureType = "text/html"
		}

		feedItems = append(feedItems, feedItem)
	}

	// Get channel metadata
	channelData, err := g.getChannelByID(channelID, "snippet")
	if err == nil {
		if items, ok := channelData["items"].([]interface{}); ok && len(items) > 0 {
			item := items[0].(map[string]interface{})
			snippet := item["snippet"].(map[string]interface{})

			title := snippet["title"].(string)
			description := ""
			if desc, ok := snippet["description"].(string); ok {
				description = desc
			}

			thumbnails := parseThumbnails(snippet["thumbnails"])
			img := getThumbnail(thumbnails)
			imageURL := ""
			if img != nil {
				imageURL = img.URL
			}

			return &feed.Data{
				Title:       fmt.Sprintf("%s - YouTube", title),
				Link:        fmt.Sprintf("https://www.youtube.com/channel/%s", channelID),
				Description: description,
				Image:       imageURL,
				Item:        feedItems,
			}, nil
		}
	}

	// Fallback if channel metadata fetch fails
	return &feed.Data{
		Title: fmt.Sprintf("Channel %s - YouTube", channelID),
		Link:  fmt.Sprintf("https://www.youtube.com/channel/%s", channelID),
		Item:  feedItems,
	}, nil
}

// getDataByUsername fetches feed data for a channel by username/handle using Google API
func (g *GoogleAPI) getDataByUsername(username string, embed bool, filterShorts bool) (*feed.Data, error) {
	// Check if username is a handle (starts with @)
	var channelID string
	var channelName string
	var channelDesc string
	var channelImage string

	if strings.HasPrefix(username, "@") {
		// Fetch page to extract channel ID from ytInitialData
		url := fmt.Sprintf("https://www.youtube.com/%s", username)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, err
		}

		// Extract ytInitialData from script tags
		var ytInitialData map[string]interface{}
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if strings.Contains(text, "ytInitialData") {
				// Extract JSON
				start := strings.Index(text, "ytInitialData = ")
				if start != -1 {
					start += len("ytInitialData = ")
					end := strings.Index(text[start:], "};")
					if end != -1 {
						jsonStr := text[start : start+end+1]
						json.Unmarshal([]byte(jsonStr), &ytInitialData)
					}
				}
			}
		})

		if ytInitialData == nil {
			return nil, errors.New("could not extract channel data")
		}

		// Extract channel metadata
		if metadata, ok := ytInitialData["metadata"].(map[string]interface{}); ok {
			if channelMeta, ok := metadata["channelMetadataRenderer"].(map[string]interface{}); ok {
				channelID = channelMeta["externalId"].(string)
				channelName = channelMeta["title"].(string)
				if desc, ok := channelMeta["description"].(string); ok {
					channelDesc = desc
				}
				if avatar, ok := channelMeta["avatar"].(map[string]interface{}); ok {
					if thumbnails, ok := avatar["thumbnails"].([]interface{}); ok && len(thumbnails) > 0 {
						thumb := thumbnails[0].(map[string]interface{})
						channelImage = thumb["url"].(string)
					}
				}
			}
		}
	} else {
		// Legacy username - use API
		channelData, err := g.getChannelByUsername(username, "snippet,contentDetails")
		if err != nil {
			return nil, err
		}

		items, ok := channelData["items"].([]interface{})
		if !ok || len(items) == 0 {
			return nil, errors.New("channel not found")
		}

		item := items[0].(map[string]interface{})
		channelID = item["id"].(string)

		snippet := item["snippet"].(map[string]interface{})
		channelName = snippet["title"].(string)
		if desc, ok := snippet["description"].(string); ok {
			channelDesc = desc
		}

		thumbnails := parseThumbnails(snippet["thumbnails"])
		img := getThumbnail(thumbnails)
		if img != nil {
			channelImage = img.URL
		}
	}

	// Now fetch channel data using channel ID
	feedData, err := g.getDataByChannelID(channelID, embed, filterShorts)
	if err != nil {
		return nil, err
	}

	// Override title and description with fetched metadata
	feedData.Title = fmt.Sprintf("%s - YouTube", channelName)
	feedData.Description = channelDesc
	feedData.Image = channelImage
	feedData.Link = fmt.Sprintf("https://www.youtube.com/channel/%s", channelID)

	return feedData, nil
}

// getDataByPlaylistID fetches feed data for a playlist using Google API
func (g *GoogleAPI) getDataByPlaylistID(playlistID string, embed bool) (*feed.Data, error) {
	// Fetch playlist metadata
	playlistData, err := g.getPlaylist(playlistID, "snippet")
	if err != nil {
		return nil, err
	}

	items, ok := playlistData["items"].([]interface{})
	if !ok || len(items) == 0 {
		return nil, errors.New("playlist not found")
	}

	item := items[0].(map[string]interface{})
	snippet := item["snippet"].(map[string]interface{})

	playlistTitle := snippet["title"].(string)
	channelTitle := snippet["channelTitle"].(string)
	playlistDesc := ""
	if desc, ok := snippet["description"].(string); ok {
		playlistDesc = desc
	}

	// Fetch playlist items
	playlistItemsData, err := g.getPlaylistItems(playlistID, "snippet")
	if err != nil {
		return nil, err
	}

	playlistItems, ok := playlistItemsData["items"].([]interface{})
	if !ok {
		return nil, errors.New("invalid playlist items data")
	}

	// Extract video IDs
	videoIDs := make([]string, 0, len(playlistItems))
	for _, item := range playlistItems {
		itemMap := item.(map[string]interface{})
		snippet := itemMap["snippet"].(map[string]interface{})

		// Skip private/deleted videos
		title := snippet["title"].(string)
		if title == "Private video" || title == "Deleted video" {
			continue
		}

		resourceID := snippet["resourceId"].(map[string]interface{})
		videoID := resourceID["videoId"].(string)
		videoIDs = append(videoIDs, videoID)
	}

	if len(videoIDs) == 0 {
		return &feed.Data{
			Title:       fmt.Sprintf("%s by %s - YouTube", playlistTitle, channelTitle),
			Link:        fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlistID),
			Description: playlistDesc,
			Item:        []feed.Item{},
		}, nil
	}

	// Fetch video details for durations
	videosData, err := g.getVideos(strings.Join(videoIDs, ","), "contentDetails")
	if err != nil {
		return nil, err
	}

	videoDetailsMap := make(map[string]map[string]interface{})
	if videosData != nil {
		if videoItems, ok := videosData["items"].([]interface{}); ok {
			for _, videoItem := range videoItems {
				videoMap := videoItem.(map[string]interface{})
				id := videoMap["id"].(string)
				videoDetailsMap[id] = videoMap
			}
		}
	}

	// Build feed items
	feedItems := make([]feed.Item, 0, len(playlistItems))
	for _, item := range playlistItems {
		itemMap := item.(map[string]interface{})
		snippet := itemMap["snippet"].(map[string]interface{})

		title := snippet["title"].(string)
		if title == "Private video" || title == "Deleted video" {
			continue
		}

		resourceID := snippet["resourceId"].(map[string]interface{})
		videoID := resourceID["videoId"].(string)

		// Get thumbnail
		thumbnails := parseThumbnails(snippet["thumbnails"])
		img := getThumbnail(thumbnails)

		// Format description
		description := ""
		if desc, ok := snippet["description"].(string); ok {
			description = formatDescription(desc)
		}

		// Render description with embed/thumbnail
		renderedDesc := renderDescription(embed, videoID, img, description)

		// Parse publish date
		pubDateStr := snippet["publishedAt"].(string)
		pubDate, _ := parseISO8601(pubDateStr)

		// Get duration if available
		var durationSeconds int
		if videoDetail, ok := videoDetailsMap[videoID]; ok {
			if contentDetails, ok := videoDetail["contentDetails"].(map[string]interface{}); ok {
				if duration, ok := contentDetails["duration"].(string); ok {
					durationSeconds = parseDuration(duration)
				}
			}
		}

		// Get author
		author := ""
		if videoOwnerTitle, ok := snippet["videoOwnerChannelTitle"].(string); ok {
			author = videoOwnerTitle
		} else {
			author = channelTitle
		}

		feedItem := feed.Item{
			Title:       title,
			Description: renderedDesc,
			Link:        fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
			PubDate:     pubDate,
			Author:      author,
		}

		// Add media enclosure
		if img != nil {
			feedItem.Media = &feed.Media{
				Thumbnail: &feed.MediaThumbnail{
					URL: img.URL,
				},
			}
		}

		// Add video URL as enclosure
		if durationSeconds > 0 {
			feedItem.EnclosureURL = getVideoURL(videoID)
			feedItem.EnclosureType = "text/html"
		}

		feedItems = append(feedItems, feedItem)
	}

	return &feed.Data{
		Title:       fmt.Sprintf("%s by %s - YouTube", playlistTitle, channelTitle),
		Link:        fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlistID),
		Description: playlistDesc,
		Item:        feedItems,
	}, nil
}

// Helper function to parse thumbnails from API response
func parseThumbnails(thumbs interface{}) ThumbnailSet {
	thumbnailSet := ThumbnailSet{}

	if thumbMap, ok := thumbs.(map[string]interface{}); ok {
		if def, ok := thumbMap["default"].(map[string]interface{}); ok {
			thumbnailSet.Default = &Thumbnail{
				URL:    def["url"].(string),
				Width:  int(def["width"].(float64)),
				Height: int(def["height"].(float64)),
			}
		}
		if medium, ok := thumbMap["medium"].(map[string]interface{}); ok {
			thumbnailSet.Medium = &Thumbnail{
				URL:    medium["url"].(string),
				Width:  int(medium["width"].(float64)),
				Height: int(medium["height"].(float64)),
			}
		}
		if high, ok := thumbMap["high"].(map[string]interface{}); ok {
			thumbnailSet.High = &Thumbnail{
				URL:    high["url"].(string),
				Width:  int(high["width"].(float64)),
				Height: int(high["height"].(float64)),
			}
		}
		if standard, ok := thumbMap["standard"].(map[string]interface{}); ok {
			thumbnailSet.Standard = &Thumbnail{
				URL:    standard["url"].(string),
				Width:  int(standard["width"].(float64)),
				Height: int(standard["height"].(float64)),
			}
		}
		if maxres, ok := thumbMap["maxres"].(map[string]interface{}); ok {
			thumbnailSet.Maxres = &Thumbnail{
				URL:    maxres["url"].(string),
				Width:  int(maxres["width"].(float64)),
				Height: int(maxres["height"].(float64)),
			}
		}
	}

	return thumbnailSet
}

// Helper function to parse ISO8601 dates
func parseISO8601(dateStr string) (time.Time, error) {
	// Try RFC3339 format (most common for YouTube API)
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t, nil
	}

	// Try ISO8601 with milliseconds
	t, err = time.Parse("2006-01-02T15:04:05.000Z", dateStr)
	if err == nil {
		return t, nil
	}

	return time.Now(), err
}
