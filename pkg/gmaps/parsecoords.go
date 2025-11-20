package gmaps

import (
	"errors"
	"regexp"
	"strconv"
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

func obfusicatedLatLong(url string) (coordinate, error) {
	detectionRegex := regexp.MustCompile(`3d(-?\d*\.?\d*)!4d(-?\d*\.?\d*)!`)
	lat_long := detectionRegex.FindAllStringSubmatch(url, 2)
	if lat_long == nil || len(lat_long[0]) != 2 {
		return coordinate{0, 0}, errors.New("no coordinates detected")
	}
	lat, err := strconv.ParseFloat(lat_long[0][1], 64)
	if err != nil {
		return coordinate{0, 0}, errors.New("no coordinates detected, latitude invalid float")
	}
	long, err := strconv.ParseFloat(lat_long[1][2], 64)
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

func simpleLatLong(url string) (coordinate, error) {
	detectionRegex := regexp.MustCompile(`(-?\d*\.?\d*),\+(-?\d*\.?\d*)`)
	lat_long := detectionRegex.FindAllStringSubmatch(url, 2)

	if lat_long == nil || len(lat_long[0]) != 2 {
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

func detectLatLong(url string) (coordinate, error) {

	coord, err := obfusicatedLatLong(url)
	if err == nil {
		return coord, nil
	}

	coord, err = simpleLatLong(url)
	if err == nil {
		return coord, nil
	}
	return coordinate{0, 0}, err
}
