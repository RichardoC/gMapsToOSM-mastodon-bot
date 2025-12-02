package osm

import "fmt"

// MakeUrl generates an OpenStreetMap URL for the given coordinates
// Example: https://www.openstreetmap.org/?lat=-12.345&lon=-12.345
func MakeUrl(latitude float64, longitude float64) string {
	return fmt.Sprintf("https://www.openstreetmap.org/?lat=%g&lon=%g", latitude, longitude)
}

// MakeOSMAppUrl generates an OSMapp URL for the given coordinates
// Example: https://osmapp.org/51.558,2.218
func MakeOSMAppUrl(latitude float64, longitude float64) string {
	return fmt.Sprintf("https://osmapp.org/%g,%g", latitude, longitude)
}
