package gmaps

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// Coordinates represents a latitude/longitude pair
type Coordinates struct {
	Latitude  float64
	Longitude float64
}

// HTTPClient interface for making HTTP requests (for testing and rate limiting)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Extractor handles extracting coordinates from Google Maps URLs
type Extractor struct {
	client HTTPClient
	logger *zap.SugaredLogger
}

// NewExtractor creates a new coordinate extractor
func NewExtractor(client HTTPClient, logger *zap.SugaredLogger) *Extractor {
	return &Extractor{
		client: client,
		logger: logger,
	}
}

// Common coordinate patterns in Google Maps URLs
var (
	// Matches @lat,lon,zoom or @lat,lon
	atCoordRegex = regexp.MustCompile(`@(-?\d+\.?\d*),(-?\d+\.?\d*)(?:,[\d.]+z)?`)

	// Matches /search/lat,lon or /search/lat,+lon (from redirected shortened URLs)
	searchCoordRegex = regexp.MustCompile(`/search/(-?\d+\.?\d*),\s*\+?\s*(-?\d+\.?\d*)`)

	// Matches ll=lat,lon query parameter
	llParamRegex = regexp.MustCompile(`[?&]ll=(-?\d+\.?\d*),(-?\d+\.?\d*)`)

	// Matches q=lat,lon query parameter
	qCoordRegex = regexp.MustCompile(`[?&]q=(-?\d+\.?\d*),(-?\d+\.?\d*)`)

	// Matches coordinates in the path like /maps/place/name/data=...!3d-12.345!4d67.890
	dataCoordRegex = regexp.MustCompile(`!3d(-?\d+\.?\d*)!4d(-?\d+\.?\d*)`)

	// Try to find coordinates anywhere in HTML meta tags or JSON
	htmlCoordRegex = regexp.MustCompile(`"(-?\d+\.?\d*),\s*(-?\d+\.?\d*)"`)
)

// ExtractCoordinates attempts to extract coordinates from a Google Maps URL
// It first tries to parse directly from the URL, then follows redirects if needed
func (e *Extractor) ExtractCoordinates(ctx context.Context, urlStr string) (*Coordinates, error) {
	// First try to extract directly from the URL
	coords, err := e.parseCoordinatesFromURL(urlStr)
	if err == nil {
		e.logger.Debugw("Extracted coordinates directly from URL", "url", urlStr, "coords", coords)
		return coords, nil
	}

	e.logger.Debugw("Could not extract from URL directly, following redirects", "url", urlStr, "error", err)

	// If that fails, follow the URL and try to extract from the final destination
	return e.extractByFollowingURL(ctx, urlStr)
}

// parseCoordinatesFromURL tries to extract coordinates directly from the URL string
func (e *Extractor) parseCoordinatesFromURL(urlStr string) (*Coordinates, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Try @lat,lon pattern (most common in modern Google Maps URLs)
	if match := atCoordRegex.FindStringSubmatch(urlStr); match != nil {
		return parseCoordMatch(match[1], match[2])
	}

	// Try /search/lat,lon pattern (common in redirected shortened URLs)
	if match := searchCoordRegex.FindStringSubmatch(urlStr); match != nil {
		return parseCoordMatch(match[1], match[2])
	}

	// Try !3d!4d pattern (in data= parameter)
	if match := dataCoordRegex.FindStringSubmatch(urlStr); match != nil {
		// Note: order is !3d (lat) !4d (lon)
		return parseCoordMatch(match[1], match[2])
	}

	// Try ll= query parameter
	if match := llParamRegex.FindStringSubmatch(urlStr); match != nil {
		return parseCoordMatch(match[1], match[2])
	}

	// Try q= query parameter
	if match := qCoordRegex.FindStringSubmatch(urlStr); match != nil {
		return parseCoordMatch(match[1], match[2])
	}

	// Check query parameters more thoroughly
	query := parsedURL.Query()

	// Check center parameter
	if center := query.Get("center"); center != "" {
		parts := strings.Split(center, ",")
		if len(parts) == 2 {
			return parseCoordMatch(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	return nil, fmt.Errorf("no coordinates found in URL")
}

// extractByFollowingURL makes an HTTP request to follow redirects and extract coordinates
func (e *Extractor) extractByFollowingURL(ctx context.Context, urlStr string) (*Coordinates, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set a reasonable User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; gMapsToOSM-bot/1.0)")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Try to extract from the final URL after redirects
	finalURL := resp.Request.URL.String()
	e.logger.Debugw("Followed redirects to final URL", "original", urlStr, "final", finalURL)

	coords, err := e.parseCoordinatesFromURL(finalURL)
	if err == nil {
		return coords, nil
	}

	// Last resort: try to extract from the HTML body
	// Read a limited amount of the body to avoid memory issues
	limitedReader := io.LimitReader(resp.Body, 1024*1024) // 1MB limit
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return e.extractFromHTML(string(body))
}

// extractFromHTML attempts to find coordinates in HTML content
func (e *Extractor) extractFromHTML(html string) (*Coordinates, error) {
	// Look for coordinates in meta tags or JSON structures
	// This is a best-effort attempt
	matches := htmlCoordRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			lat, err1 := strconv.ParseFloat(match[1], 64)
			lon, err2 := strconv.ParseFloat(match[2], 64)

			if err1 == nil && err2 == nil {
				// Sanity check: latitude should be -90 to 90, longitude -180 to 180
				if lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180 {
					e.logger.Debugw("Extracted coordinates from HTML", "lat", lat, "lon", lon)
					return &Coordinates{Latitude: lat, Longitude: lon}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no valid coordinates found in HTML")
}

// parseCoordMatch parses coordinate strings into a Coordinates struct
func parseCoordMatch(latStr, lonStr string) (*Coordinates, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}

	// Validate coordinate ranges
	if lat < -90 || lat > 90 {
		return nil, fmt.Errorf("latitude out of range: %f", lat)
	}
	if lon < -180 || lon > 180 {
		return nil, fmt.Errorf("longitude out of range: %f", lon)
	}

	return &Coordinates{Latitude: lat, Longitude: lon}, nil
}
