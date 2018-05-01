package commute

import (
	"math"
)

type Point struct {
	Lat float64
	Lon float64
}
type Delta struct {
	Lat float64
	Lon float64
}

const earthRadiusMetres float64 = 6371000

func (p Point) Delta(point Point) Delta {
	return Delta{
		Lat: p.Lat - point.Lat,
		Lon: p.Lon - point.Lon,
	}
}

func (p Point) toRadians() Point {
	return Point{
		Lat: p.Lat * math.Pi / 180,
		Lon: p.Lon * math.Pi / 180,
	}
}

//DistanceBetwnPts function returns the haversine distance between two geo points. Instead of
//including package paultag/go-haversine, just copied the code since it is too small.
//Returns the distance in Meters.
func DistanceBetwnPts(origin, position Point) float64 {
	origin = origin.toRadians()
	position = position.toRadians()

	change := origin.Delta(position)

	a := math.Pow(math.Sin(change.Lat/2), 2) + math.Cos(origin.Lat)*math.Cos(position.Lat)*math.Pow(math.Sin(change.Lon/2), 2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return float64(earthRadiusMetres * c)
}
