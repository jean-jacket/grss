package github

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jean-jacket/grss/client"
	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/feed"
	"github.com/jean-jacket/grss/routes/registry"
)

// IssuesRoute defines the GitHub issues route
var IssuesRoute = registry.Route{
	Path:        "/issue/:user/:repo",
	Name:        "Repository Issues",
	Maintainers: []string{"example"},
	Example:     "/github/issue/golang/go",
	Parameters: map[string]interface{}{
		"user": "GitHub username",
		"repo": "Repository name",
	},
	Description: "Get latest issues from a GitHub repository",
	Handler:     issuesHandler,
}

type githubIssue struct {
	Title     string    `json:"title"`
	HTMLURL   string    `json:"html_url"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	State  string `json:"state"`
	Number int    `json:"number"`
}

func issuesHandler(c *gin.Context) (*feed.Data, error) {
	user := c.Param("user")
	repo := c.Param("repo")
	state := c.DefaultQuery("state", "open")

	// Construct API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=%s&per_page=30", user, repo, state)

	// Create HTTP client
	httpClient := client.New(config.C)

	// Fetch issues
	headers := map[string]string{
		"Accept": "application/vnd.github.v3+json",
	}

	data, err := httpClient.Get(apiURL, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	// Parse response
	var issues []githubIssue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issues: %w", err)
	}

	// Build feed data
	feedData := &feed.Data{
		Title:       fmt.Sprintf("%s/%s Issues", user, repo),
		Link:        fmt.Sprintf("https://github.com/%s/%s/issues", user, repo),
		Description: fmt.Sprintf("Latest issues from %s/%s repository", user, repo),
		Language:    "en",
		Item:        make([]feed.Item, 0, len(issues)),
	}

	// Convert issues to feed items
	for _, issue := range issues {
		// Extract labels
		labels := make([]string, len(issue.Labels))
		for i, label := range issue.Labels {
			labels[i] = label.Name
		}

		item := feed.Item{
			Title:       fmt.Sprintf("#%d: %s", issue.Number, issue.Title),
			Link:        issue.HTMLURL,
			Description: issue.Body,
			PubDate:     issue.CreatedAt,
			Author:      issue.User.Login,
			Category:    labels,
			GUID:        issue.HTMLURL,
		}

		feedData.Item = append(feedData.Item, item)
	}

	// Set latest issue date as feed pubDate
	if len(issues) > 0 {
		feedData.PubDate = issues[0].CreatedAt
	}

	return feedData, nil
}
