package osm

import "fmt"

// MakeOSMAppUrl generates an OSMapp URL for the given coordinates
// Example: https://osmapp.org/51.558,2.218
func MakeOSMAppUrl(latitude float64, longitude float64) string {
	return fmt.Sprintf("https://osmapp.org/%g,%g", latitude, longitude)
}
