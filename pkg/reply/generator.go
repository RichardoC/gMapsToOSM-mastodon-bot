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
	OSMAppUrl   string
	Error       error
}

// GenerateReply processes the given texts, extracts Google Maps URLs, and generates a reply
func (g *Generator) GenerateReply(ctx context.Context, texts ...string) (string, error) {
	// Combine all texts and extract Google Maps URLs
	combinedText := strings.Join(texts, " ")
	googleMapsURLs := gmaps.ExtractGoogleMapsURLs(combinedText)

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

		osmAppURL := osm.MakeOSMAppUrl(coords.Latitude, coords.Longitude)
		g.logger.Infow("Successfully converted URL", "googleMaps", url, "osmApp", osmAppURL)
		results = append(results, ConversionResult{
			OriginalURL: url,
			OSMAppUrl:   osmAppURL,
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

	// Build the reply with all conversions (successful and failed)
	var sb strings.Builder

	sb.WriteString("Attempted to provide a link to OpenStreetMap for those Google Maps URLs:\n")

	for _, result := range results {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}

		if result.Error == nil {
			// Successful conversion
			sb.WriteString("Successfully converted ")
			sb.WriteString(result.OriginalURL)
			sb.WriteString(" to ")
			sb.WriteString(result.OSMAppUrl)
			sb.WriteString(" or ")
			sb.WriteString(result.OSMUrl)
		} else {
			// Failed conversion - inform the user
			sb.WriteString("Couldn't convert ")
			sb.WriteString(result.OriginalURL)
		}
	}

	return sb.String(), nil
}
