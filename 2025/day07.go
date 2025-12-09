package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

type Room map[int]map[int]rune

var calculatedPositions = make(map[int]map[int]int)

func (r Room) countTimelines(row, col int) int {
	if row == len(r)-1 {
		return 1
	}

	var count int

	if _, ok := calculatedPositions[row]; !ok {
		calculatedPositions[row] = make(map[int]int)
	}

	if r[row][col] == '^' {
		if _, ok := calculatedPositions[row][col]; ok {
			count += calculatedPositions[row][col]
		} else {
			count += r.countTimelines(row+1, col-1) + r.countTimelines(row+1, col+1)
		}
	}

	if r[row][col] == '.' || r[row][col] == 'S' {
		count += r.countTimelines(row+1, col)
	}

	calculatedPositions[row][col] = count
	return count
}

func main() {
	path := filepath.Join("inputs/day07.txt")

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	tachyonBeamPositions := map[int]bool{}
	room := make(Room)
	var startRow, startCol int
	partOneSplitNum := 0

	row := 0
	for sc.Scan() {
		line := []rune(sc.Text())

		if _, ok := room[row]; !ok {
			room[row] = map[int]rune{}
		}

		for col, r := range line {
			room[row][col] = r

			if r == 'S' {
				startRow = row
				startCol = col
				if _, ok := tachyonBeamPositions[col]; !ok {
					tachyonBeamPositions[col] = true
				}
			}

			if r == '^' {
				//splitterPositions = append(splitterPositions, SplitterPosition{Row: row, Col: col})
				if _, ok := tachyonBeamPositions[col]; ok {
					partOneSplitNum++
					delete(tachyonBeamPositions, col)

					if _, ok := tachyonBeamPositions[col-1]; !ok {
						tachyonBeamPositions[col-1] = true
					}

					if _, ok := tachyonBeamPositions[col+1]; !ok {
						tachyonBeamPositions[col+1] = true
					}
				}
			}
		}

		row++
	}

	partTwoTimelines := room.countTimelines(startRow, startCol)

	fmt.Printf("Part one number of splits: %d\n", partOneSplitNum)
	fmt.Printf("Part two number of timelines: %d\n", partTwoTimelines)
}
