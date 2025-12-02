package ratelimit

import (
	"net/http"
	"time"
)

// RateLimitedClient wraps an HTTP client with rate limiting
type RateLimitedClient struct {
	client  *http.Client
	limiter <-chan time.Time
}

// NewRateLimitedClient creates a new rate-limited HTTP client
// requestsPerSecond determines how many requests are allowed per second
func NewRateLimitedClient(maxRedirects int, requestsPerSecond float64) *RateLimitedClient {
	// Calculate the interval between requests
	interval := time.Duration(float64(time.Second) / requestsPerSecond)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects automatically - we handle them manually
			// to read the Location header from the first redirect response
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}

	return &RateLimitedClient{
		client:  client,
		limiter: time.Tick(interval),
	}
}

// Do executes an HTTP request with rate limiting
func (c *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
	// Wait for the rate limiter to allow the request
	<-c.limiter
	return c.client.Do(req)
}
