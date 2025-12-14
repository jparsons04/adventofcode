package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type JunctionBoxPos struct {
	X float64
	Y float64
	Z float64
}

type JunctionBox struct {
	Position JunctionBoxPos
}

type JunctionBoxPair struct {
	JunctionBox1 JunctionBoxPos
	JunctionBox2 JunctionBoxPos
	Distance     float64
	Connected    bool
}

func findPosInCircuit(pos JunctionBoxPos, circuits [][]JunctionBoxPos) int {
	var circuitIdx int

	for i, c := range circuits {
		circuitIdx = slices.IndexFunc(c, func(j JunctionBoxPos) bool { return j == pos })
		if circuitIdx != -1 {
			// Capture the outer index of the circuit that pos1 is in
			return i
		}
	}

	return -1
}

func modifyCircuits(
	circuits [][]JunctionBoxPos,
	circuitIdx1, circuitIdx2 int,
	junctionBoxPos1, junctionBoxPos2 JunctionBoxPos,
) [][]JunctionBoxPos {
	var circuit []JunctionBoxPos

	if circuitIdx1 == -1 && circuitIdx2 == -1 {
		// Neither in an existing circuit, create a new one
		circuit = append(circuit, junctionBoxPos1, junctionBoxPos2)
		circuits = append(circuits, circuit)
	} else if circuitIdx1 != -1 && circuitIdx2 == -1 {
		// pos1 found but not pos2, add pos2 to pos1's circuit
		existingCircuit := circuits[circuitIdx1]
		existingCircuit = append(existingCircuit, junctionBoxPos2)
		circuits[circuitIdx1] = existingCircuit
	} else if circuitIdx1 == -1 && circuitIdx2 != -1 {
		// pos2 found but not pos1, add pos1 to pos2's circuit
		existingCircuit := circuits[circuitIdx2]
		existingCircuit = append(existingCircuit, junctionBoxPos1)
		circuits[circuitIdx2] = existingCircuit
	} else {
		circuit1 := circuits[circuitIdx1]
		circuit2 := circuits[circuitIdx2]

		// If both are found in two different circuits, merge the two circuits
		// otherwise do nothing
		if !slices.Equal(circuit1, circuit2) {
			// Need to handle circuit index carefully - if circuitIdx1 > circuitIdx2,
			// after removing circuitIdx2, circuitIdx1 shifts down by 1
			if circuitIdx1 > circuitIdx2 {
				// remove the second circuit from circuits
				circuits = append(circuits[:circuitIdx2], circuits[circuitIdx2+1:]...)
				// add the second circuit to the first circuit
				circuit1 = append(circuit1, circuit2...)
				// Update the merged circuit (index shifted down by 1)
				circuits[circuitIdx1-1] = circuit1
			} else {
				// remove the second circuit from circuits
				circuits = append(circuits[:circuitIdx2], circuits[circuitIdx2+1:]...)
				// add the second circuit to the first circuit
				circuit1 = append(circuit1, circuit2...)
				// Update the merged circuit in the circuits slice
				circuits[circuitIdx1] = circuit1
			}
		}
	}

	return circuits

}

func partOne(junctionBoxPairs []JunctionBoxPair) ([][]JunctionBoxPos, int) {
	var circuits [][]JunctionBoxPos

	var numberofConnections int
	var nextJunctionBoxPair int

	for i := range len(junctionBoxPairs) {
		junctionBoxPos1 := junctionBoxPairs[i].JunctionBox1
		junctionBoxPos2 := junctionBoxPairs[i].JunctionBox2

		if !junctionBoxPairs[i].Connected {
			// Connect the two junction boxes together
			junctionBoxPairs[i].Connected = true

			// Also mark the reciprocal pair as connected
			for j := range junctionBoxPairs {
				if junctionBoxPairs[j].JunctionBox1 == junctionBoxPos2 &&
					junctionBoxPairs[j].JunctionBox2 == junctionBoxPos1 {
					junctionBoxPairs[j].Connected = true
					break
				}
			}

			circuitIdx1 := findPosInCircuit(junctionBoxPos1, circuits)
			circuitIdx2 := findPosInCircuit(junctionBoxPos2, circuits)

			circuits = modifyCircuits(circuits, circuitIdx1, circuitIdx2, junctionBoxPos1, junctionBoxPos2)

			// Count every connection attempt
			numberofConnections++
		}

		// Stop after 1000 connections
		if numberofConnections == 1000 {
			nextJunctionBoxPair = i + 1
			break
		}
	}

	sort.Slice(circuits, func(i, j int) bool {
		return len(circuits[i]) > len(circuits[j])
	})

	return circuits, nextJunctionBoxPair
}

func partTwo(junctionBoxPairs []JunctionBoxPair, nextJunctionBoxPair int, circuits [][]JunctionBoxPos) float64 {
	var circuit1X, circuit2X float64

	for i := nextJunctionBoxPair; i < len(junctionBoxPairs); i++ {
		junctionBoxPos1 := junctionBoxPairs[i].JunctionBox1
		junctionBoxPos2 := junctionBoxPairs[i].JunctionBox2

		circuit1X = junctionBoxPos1.X
		circuit2X = junctionBoxPos2.X

		if !junctionBoxPairs[i].Connected {
			// connect the two junction boxes together
			junctionBoxPairs[i].Connected = true

			// Also mark the reciprocal pair as connected
			for j := range junctionBoxPairs {
				if junctionBoxPairs[j].JunctionBox1 == junctionBoxPos2 &&
					junctionBoxPairs[j].JunctionBox2 == junctionBoxPos1 {
					junctionBoxPairs[j].Connected = true
					break
				}
			}

			circuitIdx1 := findPosInCircuit(junctionBoxPos1, circuits)
			circuitIdx2 := findPosInCircuit(junctionBoxPos2, circuits)

			circuits = modifyCircuits(circuits, circuitIdx1, circuitIdx2, junctionBoxPos1, junctionBoxPos2)

			if len(circuits) == 1 {
				break
			}
		}
	}

	return circuit1X * circuit2X
}

func main() {
	path := filepath.Join("inputs/day08.txt")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	junctionBoxes := make(map[JunctionBoxPos]JunctionBox)
	var junctionBoxPairs []JunctionBoxPair

	for sc.Scan() {
		line := strings.Split(sc.Text(), ",")
		pos := make([]int, len(line))
		posFloat := make([]float64, len(line))

		for i, v := range line {
			pos[i], _ = strconv.Atoi(v)
			posFloat[i] = float64(pos[i])
		}

		junctionBoxPos := JunctionBoxPos{X: posFloat[0], Y: posFloat[1], Z: posFloat[2]}
		junctionBoxes[junctionBoxPos] = JunctionBox{Position: junctionBoxPos}
	}

	// Create all junction box pairs out of all junction boxes and add them to junctionBoxPairs
	for pos1 := range junctionBoxes {
		for pos2 := range junctionBoxes {
			if pos1 != pos2 {
				x := pos1.X - pos2.X
				y := pos1.Y - pos2.Y
				z := pos1.Z - pos2.Z
				distance := math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2) + math.Pow(z, 2))
				junctionBoxPairs = append(junctionBoxPairs, JunctionBoxPair{
					JunctionBox1: pos1,
					JunctionBox2: pos2,
					Distance:     distance,
					Connected:    false,
				})
			}
		}
	}

	sort.Slice(junctionBoxPairs, func(i, j int) bool {
		return junctionBoxPairs[i].Distance < junctionBoxPairs[j].Distance
	})

	circuits, nextJunctionBoxPair := partOne(junctionBoxPairs)

	fmt.Printf("Part one: Product of the three largest circuits: %d\n", len(circuits[0])*len(circuits[1])*len(circuits[2]))

	// Add the unconnected junction boxes to circuits
	for i := range junctionBoxes {
		idx := -1
		for j, c := range circuits {
			idx = slices.IndexFunc(c, func(j JunctionBoxPos) bool { return j == i })

			if idx != -1 {
				// Capture the outer index of the circuit that pos1 is in
				idx = j
				break
			}
		}

		if idx == -1 {
			circuits = append(circuits, []JunctionBoxPos{i})
		}
	}

	productXCoordLastJunctionBox := partTwo(junctionBoxPairs, nextJunctionBoxPair, circuits)

	fmt.Printf("Part two: Product of X coordinate of last two connected junction boxes: %.0f\n", productXCoordLastJunctionBox)
}
