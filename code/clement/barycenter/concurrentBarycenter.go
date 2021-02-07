package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sync"
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

func avgMassPointsWeightedAsync(a MassPoint, b MassPoint, c chan<- MassPoint) {
	c <- avgMassPointWeighted(a, b)
}

// The first thing we need to do is make an async version of the loading procedure.
// It takes a string to work on, a channel through which to send results, and a WaitGroup pointer.
func stringToPointAsync(s string, c chan<- MassPoint, wg *sync.WaitGroup) {
	// First off, we'll defer the WaitGroup finishing operation, so however this function exits
	// it will notify the WaitGroup that it's done.
	defer wg.Done()
	// We'll create a new MassPoint to hold the result
	var newMassPoint MassPoint
	// Then we'll use Sscanf to parse the line
	_, err := fmt.Sscanf(s, "%f:%f:%f:%f", &newMassPoint.x, &newMassPoint.y, &newMassPoint.z, &newMassPoint.mass)
	// If there's an error, just abort
	if err != nil {
		return
	}
	// If there wasn't an error, send the result through the channel
	c <- newMassPoint
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

	r := bufio.NewReader(file)
	// We also need a buffered channel for results
	masspointsChan := make(chan MassPoint, 128)
	// And a waitgroup for syncronization
	var wg sync.WaitGroup
	for {
		// To actually get a line, we'll use the ReadString function
		str, err := r.ReadString('\n')
		// If the result is empty or there's an error, there are no more lines to read
		if len(str) == 0 || err != nil {
			break
		}

		// Otherwise, we'll start off a goroutine to parse the line
		wg.Add(1)
		go stringToPointAsync(str, masspointsChan, &wg)
	}

	// Now we'll set up syncronization. We need a channel for this, unbuffered.
	syncChan := make(chan bool)
	// Then we'll run a goroutine which will send a value through this channel when
	// the WaitGroup finishes.
	go func() { wg.Wait(); syncChan <- false }()

	// Finally,  we'll receive the values in a loop
	// We'll have a boolean value telling us if the computations are still running
	run := true
	// If they're still running, or there are values in the channel, keep receiving values
	for run || len(masspointsChan) > 0 {
		select {
		// If a value is available, we'll put it in the masspoints buffer
		case val := <-masspointsChan:
			masspoints = append(masspoints, val)
			// If the computations are done, we'll toggle the switch off
		case _ = <-syncChan:
			run = false
		}
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
