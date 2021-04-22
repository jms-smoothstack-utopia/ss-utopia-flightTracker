package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Status is a convenience byte type for enumerating the various states of an Aircraft.
// The value of each Status consists of a single character that should be decoded by any
// data consumers. A single byte is used to save space on Aircraft.Report records.
type Status byte

const (
	Idle            Status = 'i'
	TaxiIn          Status = 'x'
	TakeOff         Status = 't'
	Cruising        Status = 'c'
	AwaitingLanding Status = 'a'
	Landing         Status = 'l'
	TaxiOut         Status = 'o'
)

// Airport is an absolute location that can serve as a flight origin or destination.
// It consists only of the Airport IATA and the GPS Position of the Airport.
type Airport struct {
	iata     string
	location Position
}

func (a *Airport) String() string {
	return fmt.Sprintf("Airport IATA: %v\t%v", a.iata, a.location)
}

// Aircraft stores dynamic information throughout a flight.
// Use Aircraft.Report for a JSON byte array of current store.
// Quick notes regarding speed:
// 1. All speeds are assumed in knots for simplicity sake
// 2. While airspeedKnots and groundSpeedKnots are technically different (but correlated) during flight, with
// 		airspeedKnots matching the speed of the plane and groundSpeedKnots matching the speed as observed on the
//    ground, for the purposes of this application airspeedKnots is for use during flight and groundSpeedKnots
//		is used during ground travel (ie taxi events)
// As soon as state transitions to TakeOff, groundSpeedKnots is set to 0. Likewise, as soon as state is set
// to Taxi, airspeedKnots is set to 0.
// These speeds are currently static values set during state transition.
type Aircraft struct {
	tailNum  string
	flightId string

	origin      Airport
	destination Airport

	currentPos Position
	altitude   float64

	// nautical miles
	nmiToDest    float64
	nmiTravelled float64

	// for simplicity, assume all speeds are in knots
	// airspeedKnots is for use during air travel only
	// groundSpeedKnots is for use during ground/hybrid travel
	airspeedKnots      float64
	groundSpeedKnots   float64
	verticalSpeedKnots float64

	bearing float64

	hasClearance bool // landing or takeoff dependent on status

	status Status
}

// Init initializes an Aircraft with the given information.
// Initial Status is set to Idle
// Aircraft.currentPos is initialized to the given origin
// Aircraft.bearing and Aircraft.nmiToDest are calculated and initialized with given arguments.
// All other fields are 0 initialized.
func (a *Aircraft) Init(tailNum, flightId string, origin, destination *Airport) {
	a.tailNum = tailNum
	a.flightId = flightId

	a.origin = *origin
	a.destination = *destination
	a.currentPos = origin.location

	bearing, distance := origin.location.CalcVector(&destination.location)
	a.bearing = bearing
	a.nmiToDest = distance

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
		strconv.FormatFloat(a.nmiToDest, 'f', 2, 64),
		strconv.FormatFloat(a.airspeedKnots, 'f', 2, 64),
		strconv.FormatFloat(a.groundSpeedKnots, 'f', 2, 64),
		strconv.FormatFloat(a.verticalSpeedKnots, 'f', 2, 64),
		a.status,
	}
	return json.Marshal(data)
}

// TransitionState transitions an Aircraft status (if possible)
// Status is modelled as a basic state machine, allowing transitions only to the according successive
// state. This is a loop in the form:
// Idle -> TaxiOut -> TakeOff -> Cruising -> AwaitingLanding -> Landing -> TaxiIn -> Idle
// Each state additionally has a gate that stops transition if the condition is not met.
// For example, transition from TaxiOut -> TakeOff is only allowed if the current distance of the
// Aircraft is greater than a specific distance from its origin Airport.
//
// NOTE: An Aircraft must be given clearance (by setting a.hasClearance to true)
// in order to initiate TakeOff.
func (a *Aircraft) TransitionState() {
	switch a.status {
	case Idle:
		a.setTaxi(true)
	case TaxiOut:
		a.setTakeOff()
	case TakeOff:
		a.setCruising()
	case Cruising:
		a.setAwaitingLanding()
	case AwaitingLanding:
		a.setLanding()
	case Landing:
		a.setTaxi(false)
	case TaxiIn:
		a.setIdle()
	}
}

func (a *Aircraft) setTaxi(out bool) {
	if !a.hasClearance {
		return
	}

	if out {
		a.status = TaxiOut
	} else {
		a.status = TaxiIn
	}

	a.airspeedKnots = 0
	a.groundSpeedKnots = taxiSpeed
	a.verticalSpeedKnots = 0
}

func (a *Aircraft) setTakeOff() {
	if a.currentPos.CalcDistance(&a.origin.location) < taxiDistanceFromOrigin {
		return
	}

	a.status = TakeOff
	a.airspeedKnots = takeoffAirspeed
	a.groundSpeedKnots = 0
	a.verticalSpeedKnots = takeoffVerticalSpeed
}

func (a *Aircraft) setCruising() {
	if a.altitude < cruisingAltitude {
		return
	}

	a.status = Cruising
	a.airspeedKnots = cruisingAirspeed
	a.groundSpeedKnots = 0
	a.verticalSpeedKnots = 0
}

func (a *Aircraft) setAwaitingLanding() {
	if a.nmiToDest > awaitingLandingDistance {
		return
	}

	a.status = AwaitingLanding
	a.awaitClearance()

	a.airspeedKnots = awaitingLandingAirspeed
	a.groundSpeedKnots = 0
	a.verticalSpeedKnots = 0
}

func (a *Aircraft) awaitClearance() {
	go func() {
		time.Sleep(time.Second * time.Duration(clearanceWaitSeconds))
		a.hasClearance = true
	}()
}

func (a *Aircraft) setLanding() {
	if a.hasClearance {
		return
	}

	a.status = Landing
	a.airspeedKnots = landingAirSpeed
	a.groundSpeedKnots = 0
	a.verticalSpeedKnots = landingVerticalSpeed
}

func (a *Aircraft) setIdle() {
	if a.nmiToDest > idleDistanceFromDestination {
		return
	}

	a.status = Idle
	a.airspeedKnots = 0
	a.groundSpeedKnots = 0
	a.verticalSpeedKnots = 0
}
