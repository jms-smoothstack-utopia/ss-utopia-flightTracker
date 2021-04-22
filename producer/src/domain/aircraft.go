package domain

import (
	"encoding/json"
	"fmt"
	"log"
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

// Airport is an absolute Location that can serve as a flight origin or destination.
// It consists only of the Airport IATA and the GPS Position of the Airport.
type Airport struct {
	Iata     string
	Location Position
}

func (a *Airport) String() string {
	return fmt.Sprintf("Airport IATA: %v\t%v", a.Iata, a.Location)
}

// Aircraft stores dynamic information throughout a flight.
// Use Aircraft.Report for a JSON byte array of current store.
// All speeds are assumed in knots for simplicity sake.
// These speeds are currently static values set during state transition.
type Aircraft struct {
	tailNum  string
	flightId string

	origin      Airport
	destination Airport

	CurrentPos Position
	altitude   float64

	// nautical miles
	nmiToDest    float64
	nmiTravelled float64

	speedKnots     float64
	vSpeedFtPerSec float64

	bearing float64

	HasTakeOffClearance bool
	HasLandingClearance bool
	HasClearance bool // landing or takeoff dependent on Status

	Status Status
}

// Init initializes an Aircraft with the given information.
// Initial Status is set to Idle
// Aircraft.CurrentPos is initialized to the given origin
// Aircraft.bearing and Aircraft.nmiToDest are calculated and initialized with given arguments.
// All other fields are 0 initialized.
func (a *Aircraft) Init(tailNum, flightId string, origin, destination *Airport) {
	a.tailNum = tailNum
	a.flightId = flightId

	a.origin = *origin
	a.destination = *destination
	a.CurrentPos = origin.Location

	bearing, distance := origin.Location.CalcVector(&destination.Location)
	a.bearing = bearing
	a.nmiToDest = distance

	a.Status = Idle
}

// Report creates a JSON byte array consisting of Aircraft information to report to a Kinesis
// stream. A timestamp is included as part of each Report.
// Because only 1KB per record is allowed, some precision is purposefully dropped for float
// Additionally, each field name is shortened significantly to save on space.
// See FlightRecord
func (a *Aircraft) Report() ([]byte, error) {
	data := FlightRecord{
		time.Now().Format(time.RFC3339),
		a.tailNum,
		a.flightId, a.origin.Iata,
		a.destination.Iata,
		strconv.FormatFloat(a.CurrentPos.Latitude, 'f', 5, 64),
		strconv.FormatFloat(a.CurrentPos.Longitude, 'f', 5, 64),
		strconv.FormatFloat(a.altitude, 'f', 2, 64),
		strconv.FormatFloat(a.bearing, 'f', 2, 64),
		strconv.FormatFloat(a.nmiTravelled, 'f', 2, 64),
		strconv.FormatFloat(a.nmiToDest, 'f', 2, 64),
		strconv.FormatFloat(a.speedKnots, 'f', 2, 64),
		strconv.FormatFloat(a.vSpeedFtPerSec, 'f', 2, 64),
		string(a.Status),
	}

	return json.Marshal(data)
}

// FlightRecord is a data struct for Aircraft.Report records.
type FlightRecord struct {
	Time                   string
	Tail, FId, Or, Dest    string
	CLat, CLong, Alt, Brng string
	Trav, Dist, ASpd, VSpd string
	Sts                    string
}

// Travel simulates Aircraft travel in increments of one second.
// It needs to know how many seconds to simulate travel. During this time, all travel related fields
// will be updated on a per second basis. Additionally, the Aircraft.Status will attempt to transition
// state automatically. This allows dynamic speed determination.
// For "real" simulations, `wait` can be set to `true` and 1 second in realtime will elapse between
// updates. This can optionally be set to false if not needed.
// Once Travel is complete, the Report record will be placed in the given channel.
func (a *Aircraft) Travel(seconds int, wait bool, report chan<- []byte) {
	go func() {
		for i := 0; i < seconds; i++ {
			if a.Status != AwaitingLanding {
				//FIXME Either this or distance from destination is not being calculated correctly.
				travelled := a.speedKnots / 3600

				a.nmiTravelled += travelled
				a.altitude += a.vSpeedFtPerSec

				delta := a.CurrentPos.DeterminePositionDelta(travelled, a.bearing)
				a.CurrentPos = delta
				a.bearing, a.nmiToDest = a.CurrentPos.CalcVector(&a.destination.Location)
			}

			a.TransitionState()

			if wait {
				time.Sleep(time.Second)
			}
		}

		r, err := a.Report()
		if err != nil {
			log.Panicf("WARNING: Report failed for Aircraft with tailNum: %v", a.tailNum)
		}
		report <- r
	}()
}

// TransitionState transitions an Aircraft Status (if possible)
// Status is modelled as a basic state machine, allowing transitions only to the according successive
// state. This is a loop in the form:
// Idle -> TaxiOut -> TakeOff -> Cruising -> AwaitingLanding -> Landing -> TaxiIn -> Idle
// Each state additionally has a gate that stops transition if the condition is not met.
// For example, transition from TaxiOut -> TakeOff is only allowed if the current distance of the
// Aircraft is greater than a specific distance from its origin Airport.
//
// NOTE: An Aircraft must be given clearance (by setting a.HasClearance to true)
// in order to initiate TakeOff.
func (a *Aircraft) TransitionState() {
	switch a.Status {
	case Idle:
		a.setTaxiOut()
	case TaxiOut:
		a.setTakeOff()
	case TakeOff:
		a.setCruising()
	case Cruising:
		a.setAwaitingLanding()
	case AwaitingLanding:
		a.setLanding()
	case Landing:
		a.setTaxiIn()
	case TaxiIn:
		a.setIdle()
	}
}

func (a *Aircraft) setTaxiOut() {
	if !a.HasClearance {
		return
	}

	a.Status = TaxiOut
	a.speedKnots = taxiSpeed
	a.vSpeedFtPerSec = 0
}

func (a *Aircraft) setTakeOff() {
	if a.CurrentPos.CalcDistance(&a.origin.Location) < taxiDistanceFromOrigin {
		return
	}

	// reset take off clearance for eventual landing clearance.
	a.HasClearance = false

	a.Status = TakeOff
	a.speedKnots = takeoffAirspeed
	a.vSpeedFtPerSec = takeoffVerticalSpeed
}

func (a *Aircraft) setCruising() {
	if a.altitude < cruisingAltitude {
		return
	}

	a.Status = Cruising
	a.speedKnots = cruisingAirspeed
	a.vSpeedFtPerSec = 0
}

func (a *Aircraft) setAwaitingLanding() {
	if a.nmiToDest > awaitingLandingDistance {
		return
	}

	if a.Status != AwaitingLanding {
		//TODO: Refactor this to use a channel and switch while awaiting clearance
		a.awaitClearance()
	}
	a.Status = AwaitingLanding

	a.speedKnots = awaitingLandingAirspeed
	a.vSpeedFtPerSec = 0
}

func (a *Aircraft) awaitClearance() {
	go func() {
		time.Sleep(time.Second * time.Duration(ClearanceWaitSeconds))
		a.HasClearance = true
	}()
}

func (a *Aircraft) setLanding() {
	if !a.HasClearance {
		return
	}

	// reset landing clearance for eventual take off clearance
	a.HasClearance = false

	a.Status = Landing
	a.speedKnots = landingAirSpeed
	a.vSpeedFtPerSec = landingVerticalSpeed
}

func (a *Aircraft) setTaxiIn() {
	if a.altitude > 0 {
		return
	}

	a.Status = TaxiIn
	a.speedKnots = taxiSpeed
	a.vSpeedFtPerSec = 0
}

func (a *Aircraft) setIdle() {
	if a.nmiToDest > idleDistanceFromDestination {
		return
	}

	if a.Status != Idle {
		//TODO: Refactor this to use a channel and switch while awaiting clearance
		a.awaitClearance()
	}
	a.Status = Idle

	a.speedKnots = 0
	a.vSpeedFtPerSec = 0
}
