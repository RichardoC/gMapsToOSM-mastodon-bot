package osm

import "fmt"

// example
// https://www.openstreetmap.org/?lat=-12.345&lon=-12.345
func MakeUrl(latitude float64, longitude float64) string {
	return fmt.Sprintf("https://www.openstreetmap.org/?lat=%g&lon=%g", latitude, longitude)
}
