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
	Position                     JunctionBoxPos
	ClosestJunctionBox           JunctionBoxPos
	DistanceToClosestJunctionBox float64
	Connected                    bool
}

func main() {
	path := filepath.Join("inputs/day08.txt")

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	junctionBoxes := make(map[JunctionBoxPos]JunctionBox)

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

	for pos1 := range junctionBoxes {
		currentJunctionBox := junctionBoxes[pos1]

		var closestJunctionBox JunctionBoxPos
		closestDistance := float64(-1)

		// Get the shortest straight-line distance from the junction box at pos1
		// to all other junction boxes
		for pos2 := range junctionBoxes {
			if pos1 != pos2 {
				x := pos1.X - pos2.X
				y := pos1.Y - pos2.Y
				z := pos1.Z - pos2.Z

				distance := math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2) + math.Pow(z, 2))

				if closestDistance == -1 || distance < closestDistance {
					closestDistance = distance
					closestJunctionBox = pos2
				}
			}
		}

		currentJunctionBox.ClosestJunctionBox = closestJunctionBox
		currentJunctionBox.DistanceToClosestJunctionBox = closestDistance

		junctionBoxes[pos1] = currentJunctionBox
	}

	var sortedJunctionBoxes []JunctionBox

	for _, v := range junctionBoxes {
		sortedJunctionBoxes = append(sortedJunctionBoxes, v)
	}

	sort.Slice(sortedJunctionBoxes, func(i, j int) bool {
		return sortedJunctionBoxes[i].DistanceToClosestJunctionBox < sortedJunctionBoxes[j].DistanceToClosestJunctionBox
	})

	for _, v := range sortedJunctionBoxes {
		fmt.Printf("%+v\n", v)
	}

	var circuits [][]JunctionBoxPos

	for i := range 10 {
		var circuit []JunctionBoxPos

		fmt.Printf("Evaluating: %+v\n", sortedJunctionBoxes[i])
		junctionBoxPos1 := sortedJunctionBoxes[i].Position
		junctionBoxPos2 := sortedJunctionBoxes[i].ClosestJunctionBox

		if !junctionBoxes[junctionBoxPos1].Connected || !junctionBoxes[junctionBoxPos2].Connected {
			junctionBoxPos1Idx := slices.IndexFunc(sortedJunctionBoxes, func(j JunctionBox) bool { return j.Position == junctionBoxPos1 })
			junctionBoxPos2Idx := slices.IndexFunc(sortedJunctionBoxes, func(j JunctionBox) bool { return j.Position == junctionBoxPos2 })

			// connect the two junction boxes together
			sortedJunctionBoxes[junctionBoxPos1Idx].Connected = true
			sortedJunctionBoxes[junctionBoxPos2Idx].Connected = true

			fmt.Printf("After connection\n%+v\n%+v\n", sortedJunctionBoxes[junctionBoxPos1Idx], sortedJunctionBoxes[junctionBoxPos2Idx])

			circuitIdx1, circuitIdx2 := -1, -1

			// look for pos1 in circuits
			for _, c := range circuits {
				circuitIdx1 = slices.IndexFunc(c, func(j JunctionBoxPos) bool { return j == junctionBoxPos1 })
				//fmt.Printf("circuitIdx1 returned %d\n", circuitIdx1)
				if circuitIdx1 != -1 {
					break
				}
			}

			for _, c := range circuits {
				circuitIdx2 = slices.IndexFunc(c, func(j JunctionBoxPos) bool { return j == junctionBoxPos2 })
				//fmt.Printf("circuitIdx2 returned %d\n", circuitIdx2)
				if circuitIdx2 != -1 {
					break
				}
			}

			if circuitIdx1 == -1 && circuitIdx2 == -1 {

				// Neither in an existing circuit, create a new one
				circuit = append(circuit, junctionBoxPos1, junctionBoxPos2)
				circuits = append(circuits, circuit)
			} else if circuitIdx1 != -1 && circuitIdx2 == -1 {
				// pos1 found but not pos2, add pos2 to pos1's circuit
				existingCircuit := circuits[circuitIdx1]

				// briefly remove from circuits
				circuits = append(circuits[:circuitIdx1], circuits[circuitIdx1+1:]...)

				existingCircuit = append(existingCircuit, junctionBoxPos2)

				// add it back to circuits
				circuits = append(circuits, existingCircuit)

			} else if circuitIdx1 == -1 && circuitIdx2 != -1 {
				// pos2 found but not pos1, add pos1 to pos2's circuit
				existingCircuit := circuits[circuitIdx2]

				// briefly remove from circuits
				circuits = append(circuits[:circuitIdx2], circuits[circuitIdx1+2:]...)

				existingCircuit = append(existingCircuit, junctionBoxPos1)

				// add it back to circuits
				circuits = append(circuits, existingCircuit)
			}

			// if both junction boxes are found, they should both be in the same circuit already
			// in which case we do nothing
		}
	}

	sort.Slice(circuits, func(i, j int) bool {
		return len(circuits[i]) > len(circuits[j])
	})

	for i := range circuits {
		fmt.Printf("%d ", len(circuits[i]))
	}

	fmt.Printf("\n%+v\n", circuits)
}
