package main

import (
	"fmt"
	"plane-producer/src/domain"
)

func main() {
	a := domain.Aircraft{}
	origin := &domain.Airport{
		Iata: "ATL",
		Location: domain.Position{
			Latitude:  0,
			Longitude: 0,
		},
	}

	destination := &domain.Airport{
		Iata: "LAX",
		Location: domain.Position{
			Latitude:  1,
			Longitude: 1,
		},
	}

	a.Init("AB-123", "F123", origin, destination)
	a.HasClearance = true

	ch := make(chan []byte, 1)

	for a.Status == domain.Idle {
		a.Travel(1, false, ch)
	}

	for a.Status != domain.Idle {
		a.Travel(1, false, ch)
		report := <-ch
		fmt.Println(string(report))
	}
}
