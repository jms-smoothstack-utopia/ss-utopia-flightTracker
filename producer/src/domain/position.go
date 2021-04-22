package domain

import (
	"fmt"
	"math"
)

// Position consists of Latitude and Longitude GPS coordinates. This may be used to store an
// absolute point (such as the Location of an Airport) or the current Position of an Aircraft.
type Position struct {
	Latitude  float64
	Longitude float64
}

func (p Position) String() string {
	return fmt.Sprintf("Position:{Latitude: %f,Longitude: %f}", p.Latitude, p.Longitude)
}

// CalcVector calculates the bearing and distance from an origin point to a destination point.
// Given the Position consists of GPS coordinates of Latitude and Longitude, this is accomplished
// with the formulae found here: http://www.movable-type.co.uk/scripts/latlong.html
// Distance calculation is approximate as it does not account for changes in altitude (which is
// fine for this application).
//TODO validate lat/long are within expected ranges
func (p *Position) CalcVector(destination *Position) (bearing float64, distance float64) {
	bearing = p.CalcBearing(destination)
	distance = p.CalcDistance(destination)
	return
}

// CalcDistance calculates the distance between two Position structs.
// Resultant unit of measurement is nautical miles.
// Formula used is the `haversine` formula:
// a = sin²(Δφ/2) + cos φ1 ⋅ cos φ2 ⋅ sin²(Δλ/2)
func (p *Position) CalcDistance(destination *Position) float64 {
	sigma1 := p.Latitude * math.Pi / 180 // φ, λ in radians
	sigma2 := destination.Latitude * math.Pi / 180

	deltaSigma := (destination.Latitude - p.Latitude) * math.Pi / 180
	deltaLambda := (destination.Longitude - p.Longitude) * math.Pi / 180

	a := math.Sin(deltaSigma/2)*math.Sin(deltaSigma/2) +
			math.Cos(sigma1)*math.Cos(sigma2)*
					math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusMeters * c * NauticalMilesPerMeter
}

// CalcBearing calculates the directional bearing between two Position structs.
// Resultant unit of measurement is degrees.
// Formula used is the following:
// θ = atan2( sin Δλ ⋅ cos φ2 , cos φ1 ⋅ sin φ2 − sin φ1 ⋅ cos φ2 ⋅ cos Δλ )
func (p *Position) CalcBearing(destination *Position) float64 {
	y := math.Sin(destination.Longitude-p.Longitude) * math.Cos(destination.Latitude)

	x := math.Cos(p.Latitude)*math.Sin(destination.Latitude) -
			math.Sin(p.Latitude)*math.Cos(destination.Latitude)*
					math.Cos(destination.Longitude-p.Longitude)

	theta := math.Atan2(y, x)

	return math.Mod(theta*180/math.Pi+360, 360)
}

// DeterminePositionDelta determines the new Position from an existing Position given the distance
// travelled (in nautical miles) and the bearing of travel.
// Formula used is the following:
// φ2 = asin( sin φ1 ⋅ cos δ + cos φ1 ⋅ sin δ ⋅ cos θ )
// λ2 = λ1 + atan2( sin θ ⋅ sin δ ⋅ cos φ1, cos δ − sin φ1 ⋅ sin φ2 )
// where φ is Latitude, λ is Longitude, θ is the bearing (clockwise from north),
//			 δ is the angular distance d/R; d being the distance travelled, R the earth’s radius
func (p *Position) DeterminePositionDelta(distance, bearing float64) Position {
	// convert distance from nautical miles to meters
	distance /= NauticalMilesPerMeter

	angularDistance := distance / EarthRadiusMeters

	newLat := math.Asin(
		math.Sin(p.Latitude)*math.Cos(angularDistance) +
				math.Cos(p.Latitude)*math.Sin(angularDistance)*math.Cos(bearing),
	)

	newLong := p.Longitude + math.Atan2(
		math.Sin(bearing)*math.Sin(angularDistance)*math.Cos(p.Latitude),
		math.Cos(angularDistance)-math.Sin(p.Latitude)*math.Sin(newLat),
	)

	return Position{Latitude: newLat, Longitude: newLong}
}
