package gmaps

import (
	"regexp"
)

// Google Maps URL patterns we want to match
var (
	// Matches various Google Maps URL formats
	googleMapsRegex = regexp.MustCompile(`https?://(?:www\.)?(?:maps\.)?google\.(?:com|[a-z]{2}(?:\.[a-z]{2})?)/[^\s<>"]*`)

	// Shortened Google Maps URLs
	gooGlMapsRegex = regexp.MustCompile(`https?://goo\.gl/maps/[^\s<>"]*`)

	// New Google Maps share URLs
	mapsAppGooGlRegex = regexp.MustCompile(`https?://maps\.app\.goo\.gl/[^\s<>"]*`)
)

// ExtractGoogleMapsURLs finds all Google Maps URLs in the given text
func ExtractGoogleMapsURLs(text string) []string {
	urls := make([]string, 0)

	// Check for regular Google Maps URLs
	matches := googleMapsRegex.FindAllString(text, -1)
	urls = append(urls, matches...)

	// Check for goo.gl/maps shortened URLs
	matches = gooGlMapsRegex.FindAllString(text, -1)
	urls = append(urls, matches...)

	// Check for maps.app.goo.gl shortened URLs
	matches = mapsAppGooGlRegex.FindAllString(text, -1)
	urls = append(urls, matches...)

	// Deduplicate
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, url := range urls {
		if !seen[url] {
			seen[url] = true
			result = append(result, url)
		}
	}

	return result
}
