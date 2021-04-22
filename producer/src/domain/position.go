package domain

import (
	"fmt"
	"math"
)

// Position consists of latitude and longitude GPS coordinates. This may be used to store an
// absolute point (such as the location of an Airport) or the current Position of an Aircraft.
type Position struct {
	latitude  float64
	longitude float64
}

func (p Position) String() string {
	return fmt.Sprintf("Latitude: %f\tLongitude: %f", p.latitude, p.longitude)
}

// CalcVector calculates the bearing and distance from an origin point to a destination point.
// Given the Position consists of GPS coordinates of latitude and longitude, this is accomplished
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
	sigma1 := p.latitude * math.Pi / 180 // φ, λ in radians
	sigma2 := destination.latitude * math.Pi / 180

	deltaSigma := (destination.latitude - p.latitude) * math.Pi / 180
	deltaLambda := (destination.longitude - p.longitude) * math.Pi / 180

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
	y := math.Sin(destination.longitude-p.longitude) * math.Cos(destination.latitude)

	x := math.Cos(p.latitude)*math.Sin(destination.latitude) -
			math.Sin(p.latitude)*math.Cos(destination.latitude)*
					math.Cos(destination.longitude-p.longitude)

	theta := math.Atan2(y, x)

	return math.Mod(theta*180/math.Pi+360, 360)
}

// DeterminePositionDelta determines the new Position from an existing Position given the distance
// travelled (in nautical miles) and the bearing of travel.
// Formula used is the following:
// φ2 = asin( sin φ1 ⋅ cos δ + cos φ1 ⋅ sin δ ⋅ cos θ )
// λ2 = λ1 + atan2( sin θ ⋅ sin δ ⋅ cos φ1, cos δ − sin φ1 ⋅ sin φ2 )
// where φ is latitude, λ is longitude, θ is the bearing (clockwise from north),
//			 δ is the angular distance d/R; d being the distance travelled, R the earth’s radius
func (p *Position) DeterminePositionDelta(distance, bearing float64) Position {
	// convert distance from nautical miles to meters
	distance /= NauticalMilesPerMeter

	angularDistance := distance / EarthRadiusMeters

	newLat := math.Asin(
		math.Sin(p.latitude)*math.Cos(angularDistance) +
				math.Cos(p.latitude)*math.Sin(angularDistance)*math.Cos(bearing),
	)

	newLong := p.longitude + math.Atan2(
		math.Sin(bearing)*math.Sin(angularDistance)*math.Cos(p.latitude),
		math.Cos(angularDistance)-math.Sin(p.latitude)*math.Sin(newLat),
	)

	return Position{latitude: newLat, longitude: newLong}
}
