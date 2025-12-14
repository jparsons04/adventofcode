package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type freshRange struct {
	rangeStart int
	rangeEnd   int
}

func main() {
	path := filepath.Join("inputs/day05.txt")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	freshRanges := []freshRange{}

	// Get ranges first until the newline break
	for sc.Scan() {
		line := sc.Text()

		if len(line) == 0 {
			break
		}

		currentRange := strings.Split(line, "-")

		rangeStart, _ := strconv.Atoi(currentRange[0])
		rangeEnd, _ := strconv.Atoi(currentRange[1])

		freshRanges = append(freshRanges, freshRange{
			rangeStart: rangeStart,
			rangeEnd:   rangeEnd,
		})
	}

	freshCount := 0

	// Then evaluate ingredient IDs against all of the ranges
	for sc.Scan() {
		ingredientID, _ := strconv.Atoi(sc.Text())

		for _, freshRange := range freshRanges {
			if ingredientID >= freshRange.rangeStart && ingredientID <= freshRange.rangeEnd {
				freshCount++
				break
			}
		}
	}

	// Part 2: Sort the freshRanges first
	sort.Slice(freshRanges, func(i, j int) bool {
		return freshRanges[i].rangeStart < freshRanges[j].rangeStart
	})

	freshRangesNoOverlaps := []freshRange{}

	// Gradually build a new slice of non-overlapping ranges
	for i, _ := range freshRanges {
		if i == 0 {
			freshRangesNoOverlaps = append(freshRangesNoOverlaps, freshRanges[0])
			continue
		}

		noOverlapLen := len(freshRangesNoOverlaps)

		overlapFound := false

		// (e.g. existing range 3-6, new range 2-5, replace start, new range 2-6)
		if freshRanges[i].rangeStart <= freshRangesNoOverlaps[noOverlapLen-1].rangeStart && freshRanges[i].rangeEnd >= freshRangesNoOverlaps[noOverlapLen-1].rangeStart && freshRanges[i].rangeEnd <= freshRangesNoOverlaps[noOverlapLen-1].rangeEnd {
			overlapFound = true
			freshRangesNoOverlaps[noOverlapLen-1].rangeStart = freshRanges[i].rangeStart
		}

		// (e.g. existing range 3-6, new range 4-8, replace end, new range 3-8)
		if freshRanges[i].rangeEnd >= freshRangesNoOverlaps[noOverlapLen-1].rangeEnd && freshRanges[i].rangeStart >= freshRangesNoOverlaps[noOverlapLen-1].rangeStart && freshRanges[i].rangeStart <= freshRangesNoOverlaps[noOverlapLen-1].rangeEnd {
			overlapFound = true
			freshRangesNoOverlaps[noOverlapLen-1].rangeEnd = freshRanges[i].rangeEnd
		}

		// (e.g. existing range 3-6, new range 4-5, already covered by existing range, so skip)
		if freshRanges[i].rangeStart >= freshRangesNoOverlaps[noOverlapLen-1].rangeStart && freshRanges[i].rangeStart <= freshRangesNoOverlaps[noOverlapLen-1].rangeEnd && freshRanges[i].rangeEnd >= freshRangesNoOverlaps[noOverlapLen-1].rangeStart && freshRanges[i].rangeEnd <= freshRangesNoOverlaps[noOverlapLen-1].rangeEnd {
			overlapFound = true
		}

		// Unique range, append to non-overlapping ranges
		if !overlapFound {
			freshRangesNoOverlaps = append(freshRangesNoOverlaps, freshRanges[i])
		}
	}

	freshIngredientIDCount := 0
	for i := range freshRangesNoOverlaps {
		freshIngredientIDCount += (freshRangesNoOverlaps[i].rangeEnd - freshRangesNoOverlaps[i].rangeStart + 1)
	}

	fmt.Printf("Part one fresh ingredient count: %d\n", freshCount)
	fmt.Printf("Part two number of fresh ingredient IDs: %d\n", freshIngredientIDCount)
}
