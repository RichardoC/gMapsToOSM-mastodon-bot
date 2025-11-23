package gmaps

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type coordinate struct {
	latitude  float64
	longitude float64
}

func latitudeInValidRange(latitude float64) bool {
	if (latitude < -90) || (latitude > 90) {
		return false
	}
	return true
}
func longitudeInValidRange(longitude float64) bool {
	if (longitude < -180) || (longitude > 180) {
		return false
	}
	return true
}

func obfuscatedLatLong(url string) (coordinate, error) {
	detectionRegex := regexp.MustCompile(`3d(-?\d*\.?\d*)!4d(-?\d*\.?\d*)!`)
	lat_long := detectionRegex.FindAllStringSubmatch(url, 2)
	if lat_long == nil || len(lat_long[0]) != 3 {
		return coordinate{0, 0}, errors.New("no coordinates detected")
	}
	lat, err := strconv.ParseFloat(lat_long[0][1], 64)
	if err != nil {
		return coordinate{0, 0}, errors.New("no coordinates detected, latitude invalid float")
	}
	long, err := strconv.ParseFloat(lat_long[0][2], 64)
	if err != nil {
		return coordinate{0, 0}, errors.New("no coordinates detected, longitude invalid float")
	}
	if !latitudeInValidRange(lat) {
		return coordinate{0, 0}, errors.New("no coordinates detected, latitude invalid")
	}
	if !longitudeInValidRange(long) {
		return coordinate{0, 0}, errors.New("no coordinates detected, longitude invalid")
	}
	return coordinate{lat, long}, nil
}

func simpleLatLong(urlA string) (coordinate, error) {
	// detectionRegex := regexp.MustCompile(`(-?\d*\.?\d*),\+(-?\d*\.?\d*)`)
	u, err := url.Parse(urlA)
	pathSegments := strings.Split(u.Path, "/")
	relevantPath := pathSegments[len(pathSegments)-1]
	if err != nil {
		return coordinate{0, 0}, errors.New("invalid url")
	}

	lat_segment := strings.SplitN(relevantPath, ",", 2)
	// fmt.Println(lat_segment[0])
	// fmt.Println("beep")

	if len(lat_segment) != 2 {
		return coordinate{0, 0}, errors.New("no coordinates detected, latitude detection failed")
	}

	lat, err := strconv.ParseFloat(lat_segment[0], 64)
	if err != nil {
		return coordinate{0, 0}, errors.New("no coordinates detected, latitude invalid float")
	}

	long_segment := strings.SplitN(relevantPath, "+", 2)
	if len(long_segment) != 2 {
		return coordinate{0, 0}, errors.New("no coordinates detected, longitude detection failed")
	}

	long, err := strconv.ParseFloat(long_segment[1], 64)
	if err != nil {
		return coordinate{0, 0}, errors.New("no coordinates detected, longitude invalid float")
	}

	if !latitudeInValidRange(lat) {
		return coordinate{0, 0}, errors.New("no coordinates detected, latitude invalid")
	}
	if !longitudeInValidRange(long) {
		return coordinate{0, 0}, errors.New("no coordinates detected, longitude invalid")
	}
	return coordinate{lat, long}, nil
}

func detectLatLong(url string) (coordinate, error) {

	coord, err := obfuscatedLatLong(url)
	if err == nil {
		return coord, nil
	}

	coord, err = simpleLatLong(url)
	if err == nil {
		return coord, nil
	}
	return coordinate{0, 0}, err
}
