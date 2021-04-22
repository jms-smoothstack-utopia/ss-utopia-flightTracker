package domain

const EarthRadiusMeters = 6371e3
const NauticalMilesPerMeter = 0.0005399565

const (
	taxiSpeed               = 15
	takeoffAirspeed         = 150
	takeoffVerticalSpeed    = 15
	cruisingAirspeed        = 300
	cruisingAltitude        = 35_000
	awaitingLandingAirspeed = 200
	landingAirSpeed         = takeoffAirspeed
	landingVerticalSpeed    = -takeoffVerticalSpeed
)

const (
	taxiDistanceFromOrigin      = 2
	awaitingLandingDistance     = 10
	idleDistanceFromDestination = 0
)

var clearanceWaitSeconds = 120