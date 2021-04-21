package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
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

// Status is a convenience byte type for enumerating the various states of an Aircraft.
// The value of each Status consists of a single character that should be decoded by any
// data consumers. A single byte is used to save space on Aircraft.Report records.
type Status byte

const (
	Idle            Status = 'i'
	Taxi            Status = 'x'
	TakeOff         Status = 't'
	Cruising        Status = 'c'
	AwaitingLanding Status = 'a'
	Landing         Status = 'l'
)

// Airport is an absolute location that can serve as a flight origin or destination.
// It consists only of the Airport IATA and the GPS Position of the Airport.
type Airport struct {
	iata     string
	location Position
}

// Aircraft stores dynamic information throughout a flight.
// Use Aircraft.Report for a JSON byte array of current store.
type Aircraft struct {
	tailNum  string
	flightId string

	origin      Airport
	destination Airport

	currentPos Position
	altitude   float64

	distanceToDest float64

	airspeed      float64
	groundSpeed   float64
	verticalSpeed float64

	bearing float64

	status Status
}

// Init initializes an Aircraft with the given information.
// Initial Status is set to Idle
// Aircraft.currentPos is initialized to the given origin
// Aircraft.bearing and Aircraft.distanceToDest are calculated and initialized with given arguments.
// All other fields are 0 initialized.
func (a *Aircraft) Init(tailNum, flightId string, origin, destination *Airport) {
	a.tailNum = tailNum
	a.flightId = flightId

	a.origin = *origin
	a.destination = *destination
	a.currentPos = origin.location

	bearing, distance := origin.location.CalcVector(&destination.location)
	a.bearing = bearing
	a.distanceToDest = distance

	a.status = Idle
}

// Report creates a JSON byte array consisting of Aircraft information to report to the Kinesis
// stream. Because only 1KB per record is allowed, some precision is purposefully dropped for float
// values. Additionally, each field name is shortened significantly to save on space.
// A timestamp is included as part of each Report
func (a *Aircraft) Report() ([]byte, error) {
	data := struct {
		time                   string
		tail, fId, or, dest    string
		cLat, cLong, alt, brng string
		dist, aSpd, gSpd, vSpd string
		sts                    Status
	}{
		time.Now().Format(time.RFC3339),
		a.tailNum,
		a.flightId, a.origin.iata,
		a.destination.iata,
		strconv.FormatFloat(a.currentPos.latitude, 'f', 5, 64),
		strconv.FormatFloat(a.currentPos.longitude, 'f', 5, 64),
		strconv.FormatFloat(a.altitude, 'f', 2, 64),
		strconv.FormatFloat(a.bearing, 'f', 2, 64),
		strconv.FormatFloat(a.distanceToDest, 'f', 2, 64),
		strconv.FormatFloat(a.airspeed, 'f', 2, 64),
		strconv.FormatFloat(a.groundSpeed, 'f', 2, 64),
		strconv.FormatFloat(a.verticalSpeed, 'f', 2, 64),
		a.status,
	}
	return json.Marshal(data)
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

const EarthRadiusMeters = 6371e3
const NauticalMilesPerMeter = 0.0005399565

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
