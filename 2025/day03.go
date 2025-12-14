package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func turnOnBatteriesInBank(bank string, numBatteriesToFind int) int {
	var bestRatings []int
	lastSelectedPos := -1

	for i := 0; i < numBatteriesToFind; i++ {
		digitsStillNeeded := numBatteriesToFind - i
		searchStartPos := lastSelectedPos + 1
		searchEndPos := len(bank) - digitsStillNeeded + 1

		bestRating := -1
		bestPos := -1

		for bankPos := searchStartPos; bankPos < searchEndPos; bankPos++ {
			intBankPos, _ := strconv.Atoi(string(bank[bankPos]))
			if intBankPos > bestRating {
				bestRating = intBankPos
				bestPos = bankPos
			}
		}

		bestRatings = append(bestRatings, bestRating)
		lastSelectedPos = bestPos
	}

	var resultStr string
	for _, rating := range bestRatings {
		resultStr += strconv.Itoa(rating)
	}

	intResult, _ := strconv.Atoi(resultStr)

	return intResult
}

func main() {
	path := filepath.Join("inputs/day03.txt")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	var partOneTotalOutputJoltage, partTwoTotalOutputJoltage int

	for sc.Scan() {
		bank := sc.Text()

		partOneTotalOutputJoltage += turnOnBatteriesInBank(bank, 2)
		partTwoTotalOutputJoltage += turnOnBatteriesInBank(bank, 12)
	}

	fmt.Printf("Part one total output joltage: %d\n", partOneTotalOutputJoltage)
	fmt.Printf("Part two total output joltage: %d\n", partTwoTotalOutputJoltage)
}
