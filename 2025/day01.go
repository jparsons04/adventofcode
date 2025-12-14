package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
		ln := sc.Text()
		dir := string(ln[0])
		distance, _ := strconv.Atoi(ln[1:])

		instructions = append(instructions, instruction{dir, distance})
	}

	dialValue := 50
	partOneZeroCount := 0
	partTwoZeroCount := 0

	for _, v := range instructions {
		dialStartsAtZero := dialValue == 0
		// Part 2: Ticks into zero n times after n full rotations
		partTwoZeroCount += v.distance / 100

		switch v.dir {
		case "L":
			dialValue -= v.distance % 100
		case "R":
			dialValue += v.distance % 100
		}

		// Part 2: If the dial starts at zero on an iteration, zeroCount is covered
		// by the full rotations before distance % 100, otherwise we cross or land
		// on zero and the count needs to tick up
		if !dialStartsAtZero && (dialValue == 0 || dialValue < 0 || dialValue >= 100) {
			partTwoZeroCount++
		}

		if dialValue < 0 {
			dialValue += 100
		} else if dialValue >= 100 {
			dialValue -= 100
		}

		if dialValue == 0 {
			partOneZeroCount++
		}
	}

	fmt.Printf("Part 1 zero count: %d\n", partOneZeroCount)
	fmt.Printf("Part 2 zero count: %d\n", partTwoZeroCount)
}
