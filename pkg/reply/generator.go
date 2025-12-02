package reply

import (
	"context"
	"strings"

	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/gmaps"
	"github.com/RichardoC/gMapsToOSM-mastodon-bot/pkg/osm"
	"go.uber.org/zap"
)

// Generator handles generating replies for Google Maps URLs
type Generator struct {
	extractor *gmaps.Extractor
	logger    *zap.SugaredLogger
}

// NewGenerator creates a new reply generator
func NewGenerator(extractor *gmaps.Extractor, logger *zap.SugaredLogger) *Generator {
	return &Generator{
		extractor: extractor,
		logger:    logger,
	}
}

// ConversionResult represents the result of converting a single URL
type ConversionResult struct {
	OriginalURL string
	OSMUrl      string
	Error       error
}

// GenerateReply processes the given text, extracts Google Maps URLs, and generates a reply
func (g *Generator) GenerateReply(ctx context.Context, text string) (string, error) {
	// Extract all Google Maps URLs from the text
	googleMapsURLs := gmaps.ExtractGoogleMapsURLs(text)

	if len(googleMapsURLs) == 0 {
		return "No Google Maps URLs found", nil
	}

	g.logger.Infow("Found Google Maps URLs", "count", len(googleMapsURLs), "urls", googleMapsURLs)

	// Convert each URL
	results := make([]ConversionResult, 0, len(googleMapsURLs))
	successCount := 0

	for _, url := range googleMapsURLs {
		coords, err := g.extractor.ExtractCoordinates(ctx, url)
		if err != nil {
			g.logger.Warnw("Failed to extract coordinates", "url", url, "error", err)
			results = append(results, ConversionResult{
				OriginalURL: url,
				Error:       err,
			})
			continue
		}

		osmURL := osm.MakeUrl(coords.Latitude, coords.Longitude)
		g.logger.Infow("Successfully converted URL", "googleMaps", url, "osm", osmURL)
		results = append(results, ConversionResult{
			OriginalURL: url,
			OSMUrl:      osmURL,
		})
		successCount++
	}

	// Generate the reply text
	return g.formatReply(results, successCount)
}

// formatReply formats the conversion results into a reply message
func (g *Generator) formatReply(results []ConversionResult, successCount int) (string, error) {
	if successCount == 0 {
		return "Couldn't convert Google Maps link(s) to OpenStreetMap", nil
	}

	// Build the reply with all successful conversions
	var sb strings.Builder

	for _, result := range results {
		if result.Error == nil {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(result.OSMUrl)
		}
	}

	return sb.String(), nil
}
