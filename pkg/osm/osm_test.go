package osm_test

import (
	"testing"

	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/osm"

	"github.com/stretchr/testify/assert"
)

func TestMakeOSMAppUrl(t *testing.T) {
	testCases := []struct {
		name        string
		latitude    float64
		longitude   float64
		expectedURL string
	}{
		{"Negatives", -12.0, -12.0, "https://osmapp.org/-12,-12"},
		{"Zero", -0, 0, "https://osmapp.org/0,0"},
		{"Mixed", -12, 23.555, "https://osmapp.org/-12,23.555"},
		{"Example from user", 51.558, 2.218, "https://osmapp.org/51.558,2.218"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, osm.MakeOSMAppUrl(tc.latitude, tc.longitude), tc.expectedURL)
		})
	}
}
