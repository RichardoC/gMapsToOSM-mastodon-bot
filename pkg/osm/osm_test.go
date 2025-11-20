package osm_test

import (
	"testing"

	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/osm"


	"github.com/stretchr/testify/assert"

)

func TestMakeUrl(t *testing.T) {
	testCases := []struct {
		name        string
		latitude    float64
		longitude   float64
		expectedURL string
	}{
		{"Negatives", -12.0, -12.0, "https://www.openstreetmap.org/?lat=-12&lon=-12"},
		{"Zero", -0, 0, "https://www.openstreetmap.org/?lat=0&lon=0"},
		{"Mixed", -12, 23.555, "https://www.openstreetmap.org/?lat=-12&lon=23.555"},

	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, osm.MakeUrl(tc.latitude, tc.longitude), tc.expectedURL)
		})
	}
}
