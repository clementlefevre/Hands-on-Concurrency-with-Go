package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type MassPoint struct {
	x, y, z, mass float64
}

func addMassPoint(a, b MassPoint) MassPoint {
	return MassPoint{a.x + b.x, a.y + b.y, a.z + b.z, a.mass + b.mass}
}
func avgMassPoint(a, b MassPoint) MassPoint {
	sum := addMassPoint(a, b)
	return MassPoint{sum.x / 2, sum.y / 2, sum.z / 2, sum.mass}
}

func toWeightedSubspace(a MassPoint) MassPoint {
	return MassPoint{a.x * a.mass, a.y * a.mass, a.z * a.mass, a.mass}
}

func fromWeightedSubspace(a MassPoint) MassPoint {
	return MassPoint{a.x / a.mass, a.y / a.mass, a.z / a.mass, a.mass}
}
func avgMassPointWeighted(a, b MassPoint) MassPoint {
	aWeighted := toWeightedSubspace(a)
	bWeighted := toWeightedSubspace(b)
	return fromWeightedSubspace(avgMassPoint(aWeighted, bWeighted))
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

func closeFile(fi *os.File) {
	err := fi.Close()
	handle(err)

}
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Incorrect number of arguments")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	handle(err)
	defer closeFile(file)

	var masspoints []MassPoint
	startLoading := time.Now()

	for {
		var newMassPoint MassPoint
		_, err = fmt.Fscanf(file, "%f:%f:%f:%f", &newMassPoint.x, &newMassPoint.y, &newMassPoint.z, &newMassPoint.mass)
		if err == io.EOF {
			break
		} else if err != nil {
			continue
		}

		masspoints = append(masspoints, newMassPoint)

	}
	fmt.Printf("Loaded %d values from file in %s.\n", len(masspoints), time.Since(startLoading))

	if len(masspoints) < 1 {
		handle(errors.New("Insufficient values."))
	}

	startCalculation := time.Now()

	for len(masspoints) != 1 {

		var newMassPoints []MassPoint
		for i := 0; i < len(masspoints)-1; i += 2 {
			newMassPoints = append(newMassPoints, avgMassPointWeighted(masspoints[i], masspoints[i+1]))
		}
		if len(masspoints)%2 != 0 {
			newMassPoints = append(newMassPoints, masspoints[len(masspoints)-1])
		}
		masspoints = newMassPoints

	}
	systemAverage := masspoints[0]

	fmt.Printf("System barycenter is at (%f, %f, %f) and the system's mass is %f.\n", systemAverage.x, systemAverage.y, systemAverage.z, systemAverage.mass)
	fmt.Printf("Calculation took %s.\n", time.Since(startCalculation))
}
