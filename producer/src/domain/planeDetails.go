package domain

import "time"

type PlaneDetails struct {
	tailNum string
	flightId string
	timestamp time.Time

	latitude float64
	longitude float64
	altitude float64

	airspeed float64
	groundSpeed float64
	verticalSpeed float64

	compass float64
	heading float64

	attitude float64
	bank float64
	rateOfTurn float64

	deviation struct {
		degrees float64
		miles float64
	}

	status Status
}

type Status uint8

const(
	Idle Status = iota
	Taxi
	TakeOff
	Cruising
	AwaitingLanding
	Landing
)
