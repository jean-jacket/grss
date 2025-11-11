package client

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/jean-jacket/grss/config"
	"github.com/jean-jacket/grss/utils"
)

var defaultUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

// Client wraps http.Client with retry logic and proxy support
type Client struct {
	httpClient *http.Client
	config     *config.Config
}

// New creates a new HTTP client
func New(cfg *config.Config) *Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Set proxy if configured and strategy is "all"
	if cfg.Proxy.Strategy == "all" && cfg.Proxy.URI != "" {
		if proxyURL, err := url.Parse(cfg.Proxy.URI); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   cfg.RequestTimeout,
			Transport: transport,
		},
		config: cfg,
	}
}

// Get performs a GET request with retry logic
func (c *Client) Get(reqURL string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set User-Agent
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	} else {
		// Use random user agent
		req.Header.Set("User-Agent", defaultUserAgents[rand.Intn(len(defaultUserAgents))])
	}

	// Perform request with retry
	return c.doWithRetry(req)
}

// Post performs a POST request with retry logic
func (c *Client) Post(reqURL string, body io.Reader, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set User-Agent
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	} else {
		req.Header.Set("User-Agent", defaultUserAgents[rand.Intn(len(defaultUserAgents))])
	}

	return c.doWithRetry(req)
}

// doWithRetry performs the request with retry logic
func (c *Client) doWithRetry(req *http.Request) ([]byte, error) {
	var lastErr error
	maxRetries := c.config.RequestRetry

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Use proxy on retry if strategy is "on_retry"
		if attempt > 0 && c.config.Proxy.Strategy == "on_retry" && c.config.Proxy.URI != "" {
			c.enableProxy()
		}

		startTime := time.Now()
		resp, err := c.httpClient.Do(req)
		duration := time.Since(startTime)

		// Log request
		utils.LogRequest(req.Method, req.URL.String(), resp, duration, err)

		if err != nil {
			lastErr = err
			// Exponential backoff
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
			}
			continue
		}

		// Check status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success - read body
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			return body, nil
		}

		// Check if we should retry based on status code
		if shouldRetry(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			// Exponential backoff
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
			}
			continue
		}

		// Non-retryable error
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
	}

	return nil, fmt.Errorf("request failed after %d retries", maxRetries)
}

// enableProxy enables proxy for the client
func (c *Client) enableProxy() {
	if c.config.Proxy.URI != "" {
		if proxyURL, err := url.Parse(c.config.Proxy.URI); err == nil {
			transport := c.httpClient.Transport.(*http.Transport)
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
}

// shouldRetry determines if a status code should trigger a retry
func shouldRetry(statusCode int) bool {
	// Retry on these status codes (matching RSSHub's ofetch behavior)
	retryableCodes := []int{400, 408, 409, 425, 429}

	// 5xx errors are always retryable
	if statusCode >= 500 {
		return true
	}

	for _, code := range retryableCodes {
		if statusCode == code {
			return true
		}
	}

	return false
}
