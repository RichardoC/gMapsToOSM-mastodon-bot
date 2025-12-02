package gmaps_test

import (
	"testing"

	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/gmaps"
	"github.com/stretchr/testify/assert"
)

func TestExtractGoogleMapsURLs(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "No URLs",
			text:     "This is just plain text with no URLs",
			expected: []string{},
		},
		{
			name:     "Single google.com/maps URL",
			text:     "Check out this place: https://www.google.com/maps/@37.7749,-122.4194,15z",
			expected: []string{"https://www.google.com/maps/@37.7749,-122.4194,15z"},
		},
		{
			name:     "Single maps.google.com URL",
			text:     "Location: https://maps.google.com/maps?q=37.7749,-122.4194",
			expected: []string{"https://maps.google.com/maps?q=37.7749,-122.4194"},
		},
		{
			name:     "Short goo.gl/maps URL",
			text:     "Visit https://goo.gl/maps/abc123def456",
			expected: []string{"https://goo.gl/maps/abc123def456"},
		},
		{
			name:     "New maps.app.goo.gl URL",
			text:     "Here: https://maps.app.goo.gl/XyZ123",
			expected: []string{"https://maps.app.goo.gl/XyZ123"},
		},
		{
			name: "Multiple URLs",
			text: "First location https://www.google.com/maps/@37.7749,-122.4194,15z and second https://goo.gl/maps/abc123",
			expected: []string{
				"https://www.google.com/maps/@37.7749,-122.4194,15z",
				"https://goo.gl/maps/abc123",
			},
		},
		{
			name:     "URL without www",
			text:     "Visit https://google.com/maps/@37.7749,-122.4194,15z",
			expected: []string{"https://google.com/maps/@37.7749,-122.4194,15z"},
		},
		{
			name:     "Country-specific domain",
			text:     "UK location: https://www.google.co.uk/maps/@51.5074,-0.1278,15z",
			expected: []string{"https://www.google.co.uk/maps/@51.5074,-0.1278,15z"},
		},
		{
			name:     "HTTP instead of HTTPS",
			text:     "Old link: http://maps.google.com/maps?q=37.7749,-122.4194",
			expected: []string{"http://maps.google.com/maps?q=37.7749,-122.4194"},
		},
		{
			name:     "Duplicate URLs",
			text:     "Same place https://goo.gl/maps/abc123 mentioned twice https://goo.gl/maps/abc123",
			expected: []string{"https://goo.gl/maps/abc123"}, // Should be deduplicated
		},
		{
			name:     "URL with place name",
			text:     "Restaurant: https://www.google.com/maps/place/Restaurant+Name/@37.7749,-122.4194,15z",
			expected: []string{"https://www.google.com/maps/place/Restaurant+Name/@37.7749,-122.4194,15z"},
		},
		{
			name:     "Mixed with other URLs",
			text:     "Check https://example.com and this map https://maps.google.com/maps?q=37.7749,-122.4194",
			expected: []string{"https://maps.google.com/maps?q=37.7749,-122.4194"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := gmaps.ExtractGoogleMapsURLs(tc.text)
			assert.ElementsMatch(t, tc.expected, result, "URLs should match")
		})
	}
}
