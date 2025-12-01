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

func instructionSplit(r rune) bool {
	return r == 'L' || r == 'R'
}

func main() {

	path := filepath.Join("./day1_input.txt")

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
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
	zeroCount := 0

	for _, v := range instructions {
		dialStartsAtZero := dialValue == 0
		// Part 2: Ticks into zero n times after n full rotations
		zeroCount += v.distance / 100

		switch v.dir {
		case "L":
			dialValue -= v.distance % 100
		case "R":
			dialValue += v.distance % 100
		}

		if !dialStartsAtZero && (dialValue == 0 || dialValue < 0 || dialValue >= 100) {
			zeroCount++
		}

		if dialValue < 0 {
			dialValue += 100
		} else if dialValue >= 100 {
			dialValue -= 100
		}

		// For Part 1
		//if dialValue == 0 {
		//	zeroCount++
		//}
	}

	fmt.Println(zeroCount)
}
