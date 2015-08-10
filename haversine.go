package search

import (
	"math"
)

// HaversineDistance returns the approximate distance between two coordinates in
// km
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * dTr
	dLon := (lon2 - lon1) * dTr
	a := math.Sin(dLat/2.0)*math.Sin(dLat/2.0) + math.Cos(lat1*dTr)*math.Cos(lat2*dTr)*math.Sin(dLon/2.0)*math.Sin(dLon/2.0)
	return 6371 * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))
}
