package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type floorCoord struct {
	content string
}

func getNeighborPaperRolls(grid [][]byte, x int, y int) int {
	rollsFound := 0
	for yPos := -1; yPos <= 1; yPos++ {
		for xPos := -1; xPos <= 1; xPos++ {
			if xPos == 0 && yPos == 0 {
				continue
			}

			if y+yPos < 0 || x+xPos < 0 || y+yPos == len(grid) || x+xPos == len(grid[0]) {
				continue
			}

			if string(grid[y+yPos][x+xPos]) == "@" {
				rollsFound++
			}
		}
	}

	return rollsFound
}

func sweepRoomToRemovePaperRolls(grid [][]byte, remove bool) int {
	accessiblePaperRolls := 0

	for y := 0; y < len(grid); y++ {
		for x := 0; x < len(grid[y]); x++ {
			if string(grid[y][x]) == "@" {
				surroundingPaperRolls := getNeighborPaperRolls(grid, x, y)

				if surroundingPaperRolls < 4 {
					accessiblePaperRolls++

					// Remove the paper roll if it can be removed
					if remove {
						gridRow := []byte(grid[y])
						gridRow[x] = '.'
						grid[y] = gridRow
					}
				}
			}
		}
	}

	return accessiblePaperRolls
}

func main() {
	path := filepath.Join("inputs/day04.txt")

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	grid := make([][]byte, 0)
	sc := bufio.NewScanner(f)

	for sc.Scan() {
		chars := []byte(sc.Text())
		grid = append(grid, chars)
	}

	partOneAccessiblePaperRolls := sweepRoomToRemovePaperRolls(grid, false)
	fmt.Printf("Part one accessible rolls of paper: %d\n", partOneAccessiblePaperRolls)

	partTwoRemovedPaperRolls := 0

	for {
		removedPaperRolls := sweepRoomToRemovePaperRolls(grid, true)
		partTwoRemovedPaperRolls += removedPaperRolls

		if removedPaperRolls == 0 {
			break
		}
	}

	fmt.Printf("Part two removed rolls of paper: %d\n", partTwoRemovedPaperRolls)
}
