package domain

import (
	"plane-producer/src/utils/test_utils"
	"strconv"
	"testing"
)

func TestAircraft_Init(t *testing.T) {
	a := Aircraft{}
	tailNum := "ABC-123"
	flightId := "F1234"

	origin := &Airport{
		Iata: "ATL",
		Location: Position{
			Latitude:  33.640411,
			Longitude: -84.419853,
		},
	}

	destination := &Airport{
		Iata: "LAX",
		Location: Position{
			Latitude:  33.942791,
			Longitude: -118.410042,
		},
	}

	bearing, distance := origin.Location.CalcVector(&destination.Location)

	a.Init(tailNum, flightId, origin, destination)

	test_utils.ErrorIf(t, a.tailNum != tailNum, "tailNum", tailNum, a.tailNum)

	test_utils.ErrorIf(t, a.flightId != flightId, "flightId", flightId, a.flightId)

	test_utils.ErrorIf(t, a.origin != *origin, "origin", origin.String(), a.origin.String())

	test_utils.ErrorIf(
		t, a.destination != *destination, "destination", destination.String(), a.destination.String(),
	)

	test_utils.ErrorIf(
		t, a.currentPos != a.origin.Location, "currentPos", a.origin.String(), a.currentPos.String(),
	)

	test_utils.ErrorIf(
		t, a.bearing != bearing, "bearing",
		strconv.FormatFloat(bearing, 'f', 5, 64),
		strconv.FormatFloat(a.bearing, 'f', 5, 64),
	)

	test_utils.ErrorIf(
		t, a.nmiToDest != distance, "nmiToDest",
		strconv.FormatFloat(distance, 'f', 5, 64),
		strconv.FormatFloat(a.nmiToDest, 'f', 5, 64),
	)

	test_utils.ErrorIf(t, a.Status != Idle, "Status", string(Idle), string(a.Status))
}
