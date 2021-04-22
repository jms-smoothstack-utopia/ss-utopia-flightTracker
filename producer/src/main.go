package main

import (
	"fmt"
	"os"
	"plane-producer/src/domain"
)

func main() {
	a := domain.Aircraft{}
	origin := &domain.Airport{
		Iata: "ATL",
		Location: domain.Position{
			Latitude:  33.640411,
			Longitude: -84.419853,
		},
	}

	destination := &domain.Airport{
		Iata: "LAX",
		Location: domain.Position{
			Latitude:  33.942791,
			Longitude: -118.410042,
		},
	}

	a.Init("AB-123", "F123", origin, destination)

	r, err := a.Report()
	if err != nil {
		fmt.Print("Error marshalling.")
	}
	_, err = os.Stdout.Write(r)
	if err != nil {
		return
	}
}
