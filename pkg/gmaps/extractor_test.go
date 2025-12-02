package gmaps_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/gmaps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// mockHTTPClient is a simple mock that doesn't actually make requests
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.response != nil {
		return m.response, nil
	}

	// Return a default error response if nothing is configured
	return nil, http.ErrHandlerTimeout
}

func TestExtractCoordinates(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		expectLat float64
		expectLon float64
		shouldErr bool
	}{
		{
			name:      "At-coordinate format with zoom",
			url:       "https://www.google.com/maps/@37.7749,-122.4194,15z",
			expectLat: 37.7749,
			expectLon: -122.4194,
			shouldErr: false,
		},
		{
			name:      "At-coordinate format without zoom",
			url:       "https://www.google.com/maps/@-33.8688,151.2093",
			expectLat: -33.8688,
			expectLon: 151.2093,
			shouldErr: false,
		},
		{
			name:      "Place with at-coordinates",
			url:       "https://www.google.com/maps/place/Sydney+NSW/@-33.8688,151.2093,12z",
			expectLat: -33.8688,
			expectLon: 151.2093,
			shouldErr: false,
		},
		{
			name:      "Query parameter q with coordinates",
			url:       "https://maps.google.com/maps?q=51.5074,-0.1278",
			expectLat: 51.5074,
			expectLon: -0.1278,
			shouldErr: false,
		},
		{
			name:      "Query parameter ll with coordinates",
			url:       "https://maps.google.com/maps?ll=40.7128,-74.0060",
			expectLat: 40.7128,
			expectLon: -74.0060,
			shouldErr: false,
		},
		{
			name:      "Data format with !3d!4d",
			url:       "https://www.google.com/maps/place/Tokyo/@35.6762!3d35.6762!4d139.6503!5m1",
			expectLat: 35.6762,
			expectLon: 139.6503,
			shouldErr: false,
		},
		{
			name:      "Negative coordinates",
			url:       "https://www.google.com/maps/@-34.6037,-58.3816,15z",
			expectLat: -34.6037,
			expectLon: -58.3816,
			shouldErr: false,
		},
		{
			name:      "Zero coordinates",
			url:       "https://www.google.com/maps/@0,0,15z",
			expectLat: 0,
			expectLon: 0,
			shouldErr: false,
		},
		{
			name:      "Integer coordinates",
			url:       "https://www.google.com/maps/@12,-34,15z",
			expectLat: 12,
			expectLon: -34,
			shouldErr: false,
		},
		{
			name:      "Mussenden Temple - obfuscated URL with !3d!4d format",
			url:       "https://www.google.com/maps/place/Mussenden+Temple/data=!4m7!3m6!1s0x48602287e2f8db07:0xa0c2c065afc70175!8m2!3d55.1677806!4d-6.8108972!16zL20vMDZndzVw!19sChIJB9v44ociYEgRdQHHr2XAwqA?coh=277533&entry=tts&g_ep=EgoyMDI1MTExNy4wIPu8ASoASAFQAw%3D%3D&skid=6864092b-d6e1-4c04-8864-4ec8b7c0d841",
			expectLat: 55.1677806,
			expectLon: -6.8108972,
			shouldErr: false,
		},
		{
			name:      "No coordinates in URL",
			url:       "https://www.google.com/maps/search/restaurants",
			shouldErr: true,
		},
		{
			name:      "Invalid URL",
			url:       "not a url at all",
			shouldErr: true,
		},
		{
			name:      "Complex place URL with multiple parameters",
			url:       "https://www.google.com/maps/place/Golden+Gate+Bridge/@37.8199,-122.4783,17z/data=!3m1!4b1!4m6!3m5!1s0x808586deffffffb7:0xcb4ed812802cede8",
			expectLat: 37.8199,
			expectLon: -122.4783,
			shouldErr: false,
		},
		{
			name:      "Search format with plus-separated coordinates",
			url:       "https://www.google.com/maps/search/20.533907,+27.158833?entry=tts",
			expectLat: 20.533907,
			expectLon: 27.158833,
			shouldErr: false,
		},
		{
			name:      "Search format with space-separated coordinates",
			url:       "https://www.google.com/maps/search/54.375880, -5.551608?entry=tts",
			expectLat: 54.375880,
			expectLon: -5.551608,
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			mockClient := &mockHTTPClient{}
			extractor := gmaps.NewExtractor(mockClient, logger)

			ctx := context.Background()
			coords, err := extractor.ExtractCoordinates(ctx, tc.url)

			if tc.shouldErr {
				assert.Error(t, err, "Should return error for: %s", tc.url)
				assert.Nil(t, coords)
			} else {
				require.NoError(t, err, "Should not return error for: %s", tc.url)
				require.NotNil(t, coords)
				assert.InDelta(t, tc.expectLat, coords.Latitude, 0.0001, "Latitude should match")
				assert.InDelta(t, tc.expectLon, coords.Longitude, 0.0001, "Longitude should match")
			}
		})
	}
}

func TestCoordinateValidation(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		{
			name:      "Valid latitude range (max)",
			url:       "https://www.google.com/maps/@90,0,15z",
			shouldErr: false,
		},
		{
			name:      "Valid latitude range (min)",
			url:       "https://www.google.com/maps/@-90,0,15z",
			shouldErr: false,
		},
		{
			name:      "Valid longitude range (max)",
			url:       "https://www.google.com/maps/@0,180,15z",
			shouldErr: false,
		},
		{
			name:      "Valid longitude range (min)",
			url:       "https://www.google.com/maps/@0,-180,15z",
			shouldErr: false,
		},
		{
			name:      "Invalid latitude (too high)",
			url:       "https://www.google.com/maps/@91,0,15z",
			shouldErr: true,
		},
		{
			name:      "Invalid latitude (too low)",
			url:       "https://www.google.com/maps/@-91,0,15z",
			shouldErr: true,
		},
		{
			name:      "Invalid longitude (too high)",
			url:       "https://www.google.com/maps/@0,181,15z",
			shouldErr: true,
		},
		{
			name:      "Invalid longitude (too low)",
			url:       "https://www.google.com/maps/@0,-181,15z",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()
			mockClient := &mockHTTPClient{}
			extractor := gmaps.NewExtractor(mockClient, logger)

			ctx := context.Background()
			coords, err := extractor.ExtractCoordinates(ctx, tc.url)

			if tc.shouldErr {
				assert.Error(t, err, "Should return error for invalid coordinates")
				assert.Nil(t, coords)
			} else {
				assert.NoError(t, err, "Should not return error for valid coordinates")
				assert.NotNil(t, coords)
			}
		})
	}
}

// mockRedirectHTTPClient simulates following HTTP redirects
type mockRedirectHTTPClient struct {
	redirectMap map[string]string
}

func (m *mockRedirectHTTPClient) Do(req *http.Request) (*http.Response, error) {
	originalURL := req.URL.String()

	// Check if we have a redirect for this URL
	if finalURL, ok := m.redirectMap[originalURL]; ok {
		// Parse the final URL
		parsedURL, err := url.Parse(finalURL)
		if err != nil {
			return nil, err
		}

		// Create a new request with the final URL (simulating the redirect being followed)
		finalReq := &http.Request{
			Method: "GET",
			URL:    parsedURL,
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("")),
			Request:    finalReq,
		}, nil
	}

	return nil, http.ErrHandlerTimeout
}

func TestExtractCoordinatesWithRedirect(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		redirectURL string
		expectLat   float64
		expectLon   float64
		shouldErr   bool
	}{
		{
			name:        "maps.app.goo.gl shortened URL redirects to /search/ format",
			url:         "https://maps.app.goo.gl/Cv5nHxys6A7YZhC58",
			redirectURL: "https://www.google.com/maps/search/20.533907,+27.158833?entry=tts&g_ep=EgoyMDI1MTEyMy4xIPu8ASoASAFQAw%3D%3D&skid=8ad50c34-482d-416a-b2d1-ff4f31fea460",
			expectLat:   20.533907,
			expectLon:   27.158833,
			shouldErr:   false,
		},
		{
			name:        "maps.app.goo.gl another shortened URL (20°32'02.1\"N 27°09'31.8\"W)",
			url:         "https://maps.app.goo.gl/prqsLMsGfWn9fXP76",
			redirectURL: "https://www.google.com/maps/search/20.533917,+-27.158833?entry=tts",
			expectLat:   20.533917,
			expectLon:   -27.158833,
			shouldErr:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t).Sugar()

			// Create a mock client that simulates the redirect
			mockClient := &mockRedirectHTTPClient{
				redirectMap: map[string]string{
					tc.url: tc.redirectURL,
				},
			}

			extractor := gmaps.NewExtractor(mockClient, logger)

			ctx := context.Background()
			coords, err := extractor.ExtractCoordinates(ctx, tc.url)

			if tc.shouldErr {
				assert.Error(t, err)
				assert.Nil(t, coords)
			} else {
				require.NoError(t, err)
				require.NotNil(t, coords)
				assert.InDelta(t, tc.expectLat, coords.Latitude, 0.0001, "Latitude should match")
				assert.InDelta(t, tc.expectLon, coords.Longitude, 0.0001, "Longitude should match")
			}
		})
	}
}
