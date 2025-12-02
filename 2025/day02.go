package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

func partOne(intFirst int, intSecond int) int {
	invalidIDSum := 0
	//fmt.Printf("First: %d, Second: %d\n", intFirst, intSecond)
	for i := intFirst; i <= intSecond; i++ {
		v := strings.TrimSpace(strconv.Itoa(i))

		if v[:len(v)/2] == v[len(v)/2:] {
			//fmt.Printf("Invalid ID found: %d\n", i)
			invalidIDSum += i
		}
	}

	return invalidIDSum
}

func isInvalid(value string, seqToCheck string, seqLength int) bool {
	for j := 0; j < len(value); j = j + seqLength {
		if j+seqLength > len(value) {
			return false
		}

		if value[j:j+seqLength] != seqToCheck {
			//fmt.Printf("Invalid not found: %v, len: %d\n", value, seqLength)
			return false
		}
	}

	//fmt.Printf("Invalid found: %v, len: %d\n", value, seqLength)

	return true
}

func partTwo(intFirst int, intSecond int) int {
	invalidIDSum := 0
	//fmt.Printf("First: %d, Second: %d\n", intFirst, intSecond)
	for i := intFirst; i <= intSecond; i++ {
		v := strings.TrimSpace(strconv.Itoa(i))

		for valLength := 1; valLength < len(v); valLength++ {
			seqToCheck := string(v[:valLength])

			isInvalid := isInvalid(v, seqToCheck, valLength)

			if isInvalid {
				invalidIDSum += i
				break
			}
		}
	}

	return invalidIDSum
}

func main() {
	path := filepath.Join("inputs/day02.txt")

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	contents := string(bytes)
	ranges := strings.Split(contents, ",")
	partOneSum := 0
	partTwoSum := 0

	for _, r := range ranges {
		ids := strings.Split(r, "-")

		strFirst := strings.TrimSpace(ids[0])
		intFirst, err := strconv.Atoi(strFirst)
		if err != nil {
			return
		}

		strSecond := strings.TrimSpace(ids[1])
		intSecond, err := strconv.Atoi(strSecond)
		if err != nil {
			return
		}

		// invalid IDs must have even lengths in part one
		if len(ids[0])%2 == 0 || len(ids[1])%2 == 0 {
			partOneSum += partOne(intFirst, intSecond)
		}

		partTwoSum += partTwo(intFirst, intSecond)
	}

	fmt.Printf("partOneSum: %d\n", partOneSum)
	fmt.Printf("partTwoSum: %d\n", partTwoSum)
}
