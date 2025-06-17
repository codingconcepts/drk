package model

import "fmt"

// LatLon represents a map coordinate.
type LatLon struct {
	Lat float64
	Lon float64
}

func (ll LatLon) String() string {
	return fmt.Sprintf("Point(%f %f)", ll.Lon, ll.Lat)
}
