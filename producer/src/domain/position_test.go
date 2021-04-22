package domain

import (
	"math"
	"testing"
)

func TestPosition_CalcVector(t *testing.T) {
	scenarios := []struct {
		origin         Position
		destination    Position
		expectDistance float64
		expectBearing  float64
		maxDelta       float64
	}{
		{
			origin:         Position{Latitude: 0, Longitude: 0},
			destination:    Position{Latitude: 0, Longitude: 0},
			expectDistance: 0,
			expectBearing:  0,
			maxDelta:       0,
		},
		{
			origin:         Position{Latitude: 0, Longitude: 0},
			destination:    Position{Latitude: 1, Longitude: 1},
			expectDistance: 84.91,
			expectBearing:  28.38,
			maxDelta:       0.01,
		},
		{
			origin:         Position{Latitude: 1, Longitude: 1},
			destination:    Position{Latitude: 0, Longitude: 0},
			expectDistance: 84.91,
			expectBearing:  241.62,
			maxDelta:       0.01,
		},
		{
			origin:         Position{Latitude: 5, Longitude: 5},
			destination:    Position{Latitude: 6, Longitude: 6},
			expectDistance: 84.71,
			expectBearing:  62.63,
			maxDelta:       0.01,
		},
		{
			origin:         Position{Latitude: 5, Longitude: 5},
			destination:    Position{Latitude: -6, Longitude: -6},
			expectDistance: 933.27,
			expectBearing:  85.04,
			maxDelta:       .01,
		},
		{
			origin:         Position{Latitude: 5, Longitude: 5},
			destination:    Position{Latitude: -6, Longitude: 6},
			expectDistance: 663.16,
			expectBearing:  54.48,
			maxDelta:       .01,
		},
		{
			origin:         Position{Latitude: 5, Longitude: 5},
			destination:    Position{Latitude: 6, Longitude: -6},
			expectDistance: 660.12,
			expectBearing:  94.48,
			maxDelta:       .01,
		},
		{
			origin:         Position{Latitude: 5, Longitude: 5},
			destination:    Position{Latitude: 6, Longitude: 5},
			expectDistance: 60.04,
			expectBearing:  0,
			maxDelta:       .01,
		},
		{
			origin:         Position{Latitude: 5, Longitude: 5},
			destination:    Position{Latitude: 5, Longitude: 6},
			expectDistance: 59.81,
			expectBearing:  117.65,
			maxDelta:       .01,
		},
	}

	for i, s := range scenarios {
		gotDistance := s.origin.CalcDistance(&s.destination)
		gotBearing := s.origin.CalcBearing(&s.destination)

		deltaDistance := math.Abs(gotDistance - s.expectDistance)
		deltaBearing := math.Abs(gotBearing - s.expectBearing)

		if deltaDistance > s.maxDelta {
			t.Errorf("Failure on Scenario %d DISTANCE!\n"+
					"Origin: %v\tDestination: %v\n"+
					"Got: %f\tExpected: %f\tMax Delta: %f",
				i, s.origin, s.destination, gotDistance, s.expectDistance, s.maxDelta)
		}

		if deltaBearing > s.maxDelta {
			t.Errorf("Failure on Scenario %d BEARING!\n"+
					"Origin: %v\tDestination: %v\n"+
					"Got: %f\tExpected: %f\tMax Delta: %f",
				i, s.origin, s.destination, gotBearing, s.expectBearing, s.maxDelta)
		}
	}
}
