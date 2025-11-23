package gmaps

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatitudeInValidRange(t *testing.T) {
	testCases := []struct {
		name     string
		latitude float64
		expected bool
	}{
		{"InvalidLargeNegative", -4000, false},
		{"InvalidLargePositive", 4000, false},
		{"InvalidNegative", -91, false},
		{"InvalidPositive", 91, false},
		{"ValidNeutral", 0, true},
		{"ValidNegative", -10, true},
		{"ValidNegativeExtreme", -90, true},
		{"ValidPositive", 10, true},
		{"ValidPositiveExtreme", 90, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, latitudeInValidRange(tc.latitude), tc.expected)
		})
	}
}

func TestLongitudeInValidRange(t *testing.T) {
	testCases := []struct {
		name      string
		longitude float64
		expected  bool
	}{
		{"InvalidLargeNegative", -4000, false},
		{"InvalidLargePositive", 4000, false},
		{"InvalidNegative", 181, false},
		{"InvalidPositive", 181, false},
		{"ValidNeutral", 0, true},
		{"ValidNegative", -80, true},
		{"ValidNegativeExtreme", -180, true},
		{"ValidPositive", 80, true},
		{"ValidPositiveExtreme", 180, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, longitudeInValidRange(tc.longitude), tc.expected)
		})
	}
}

func TestObfusicatedLatLong(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected struct {
			coord coordinate
			err   error
		}
	}{
		{
			name: "obfuscatedUrl",
			url:  "https://www.google.com/maps/place/Mussenden+Temple/data=!4m7!3m6!1s0x48602287e2f8db07:0xa0c2c065afc70175!8m2!3d55.1677806!4d-6.8108972!16zL20vMDZndzVw!19sChIJB9v44ociYEgRdQHHr2XAwqA?coh=277533&entry=tts&g_ep=EgoyMDI1MTExNy4wIPu8ASoASAFQAw%3D%3D&skid=6864092b-d6e1-4c04-8864-4ec8b7c0d841",
			expected: struct {
				coord coordinate
				err   error
			}{
				coord: coordinate{55.1677806, -6.8108972},
				err:   nil,
			},
		},
		{
			name: "obviousUrl",
			url:  "https://www.google.com/maps/search/54.375880,+-5.551608?entry=tts&g_ep=EgoyMDI1MTExNy4wIPu8ASoASAFQAw%3D%3D&skid=b6f8004e-1d60-49c0-b344-2186a6902830",
			expected: struct {
				coord coordinate
				err   error
			}{
				coord: coordinate{54.375880, -5.551608},
				err:   nil,
			},
		},
		{
			name: "totallyInvalidUrl",
			url:  "https://www.yahoo.com/maps/search/fthjdfgsjkhsdtgjkhfstgjkhfjtghksjkmfghbjkh",
			expected: struct {
				coord coordinate
				err   error
			}{
				coord: coordinate{0, 0},
				err:   errors.New("no coordinates detected, latitude detection failed"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coord, err := detectLatLong(tc.url)
			assert.Equal(t, tc.expected.err, err)
			assert.Equal(t, tc.expected.coord, coord)
		})
	}
}
