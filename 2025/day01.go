package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const (
	dialMax      = 100
	fullRotation = 100
)

type instruction struct {
	dir      string
	distance int
}

func main() {
	path := filepath.Join("inputs/day01.txt")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	instructions := []instruction{}

	for sc.Scan() {
		line := sc.Text()
		dir := string(line[0])
		distance, err := strconv.Atoi(line[1:])
		if err != nil {
			panic(err)
		}

		instructions = append(instructions, instruction{dir, distance})
	}

	dialValue := 50
	partOneZeroCount := 0
	partTwoZeroCount := 0

	for _, v := range instructions {
		dialStartsAtZero := dialValue == 0
		// Part 2: Ticks into zero n times after n full rotations
		partTwoZeroCount += v.distance / fullRotation

		switch v.dir {
		case "L":
			dialValue -= v.distance % fullRotation
		case "R":
			dialValue += v.distance % fullRotation
		}

		// Part 2: If the dial starts at zero on an iteration, zeroCount is covered
		// by the full rotations before distance % fullRotation, otherwise we cross or land
		// on zero and the count needs to tick up
		if !dialStartsAtZero && (dialValue == 0 || dialValue < 0 || dialValue >= dialMax) {
			partTwoZeroCount++
		}

		if dialValue < 0 {
			dialValue += dialMax
		} else if dialValue >= 100 {
			dialValue -= dialMax
		}

		if dialValue == 0 {
			partOneZeroCount++
		}
	}

	fmt.Printf("Part 1 zero count: %d\n", partOneZeroCount)
	fmt.Printf("Part 2 zero count: %d\n", partTwoZeroCount)
}
