package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type TileCoord struct {
	Col float64
	Row float64
}

type LineSegment struct {
	Start TileCoord
	End   TileCoord
}

type RectCandidate struct {
	Corner1 TileCoord
	Corner2 TileCoord
	Area    float64
}

func getArea(tile1, tile2 TileCoord) float64 {
	return (math.Abs(tile2.Col-tile1.Col) + 1) * (math.Abs(tile2.Row-tile1.Row) + 1)
}

func isOnBoundary(point TileCoord, redTiles []TileCoord, boundaries []LineSegment) bool {
	if slices.Contains(redTiles, point) {
		return true
	}

	for _, boundary := range boundaries {
		minRow := math.Min(boundary.Start.Row, boundary.End.Row)
		minCol := math.Min(boundary.Start.Col, boundary.End.Col)
		maxRow := math.Max(boundary.Start.Row, boundary.End.Row)
		maxCol := math.Max(boundary.Start.Col, boundary.End.Col)

		if boundary.Start.Col == boundary.End.Col {
			// Vertical boundary
			if boundary.Start.Col == point.Col && minRow <= point.Row && point.Row <= maxRow {
				return true
			}
		} else {
			// Horizontal boundary
			if boundary.Start.Row == point.Row && minCol <= point.Col && point.Col <= maxCol {
				return true
			}
		}
	}

	return false
}

func isInside(point TileCoord, boundaries []LineSegment) bool {
	intersections := 0
	for _, segment := range boundaries {
		// Vertical line segments only
		if segment.Start.Col == segment.End.Col {
			segCol := segment.Start.Col
			minRow := math.Min(segment.Start.Row, segment.End.Row)
			maxRow := math.Max(segment.Start.Row, segment.End.Row)
			if segCol > point.Col && minRow <= point.Row && point.Row < maxRow {
				intersections++
			}
		}
	}

	return intersections%2 == 1
}

func generatePerimeter(corners []TileCoord) []TileCoord {
	perimeter := make([]TileCoord, 0)

	minCol := math.Min(corners[0].Col, corners[1].Col)
	minRow := math.Min(corners[0].Row, corners[1].Row)

	maxCol := math.Max(corners[0].Col, corners[1].Col)
	maxRow := math.Max(corners[0].Row, corners[1].Row)

	// Top edge (left to right)
	for col := minCol + 1; col <= maxCol; col++ {
		perimeter = append(perimeter, TileCoord{Col: col, Row: minRow})
	}

	// Right edge (top to bottom)
	for row := minRow + 1; row <= maxRow; row++ {
		perimeter = append(perimeter, TileCoord{Col: maxCol, Row: row})
	}

	// Bottom edge (right to left)
	for col := maxCol - 1; col >= minCol; col-- {
		perimeter = append(perimeter, TileCoord{Col: col, Row: maxRow})
	}

	// Left edge (bottom to top)
	for row := maxRow - 1; row >= minRow; row-- {
		perimeter = append(perimeter, TileCoord{Col: minCol, Row: row})
	}

	return perimeter
}

func buildBoundary(tiles []TileCoord) []LineSegment {
	segments := make([]LineSegment, len(tiles))
	for i := range tiles {
		nextIdx := (i + 1) % len(tiles)
		segments[i] = LineSegment{Start: tiles[i], End: tiles[nextIdx]}
	}
	return segments
}

func isValidRectangle(rect RectCandidate, tiles []TileCoord, boundaries []LineSegment) bool {
	perimeter := generatePerimeter([]TileCoord{rect.Corner1, rect.Corner2})

	for _, point := range perimeter {
		if !isOnBoundary(point, tiles, boundaries) {
			if !isInside(point, boundaries) {
				return false
			}
		}
	}

	return true
}

func main() {
	path := filepath.Join("inputs/day09.txt")

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	tiles := make([]TileCoord, 0)

	for sc.Scan() {
		tile := strings.Split(sc.Text(), ",")
		pos := make([]int, len(tile))
		posFloat := make([]float64, len(tile))

		for i, v := range tile {
			pos[i], _ = strconv.Atoi(v)
			posFloat[i] = float64(pos[i])
		}

		tiles = append(tiles, TileCoord{Col: posFloat[0], Row: posFloat[1]})
	}

	boundaries := buildBoundary(tiles)

	var largestArea float64
	var largestAreaInsideBoundaries float64

	rectCandidates := make([]RectCandidate, 0)

	for _, tile1 := range tiles {
		for _, tile2 := range tiles {
			if tile1 != tile2 {
				area := getArea(tile1, tile2)

				if area > largestArea {
					largestArea = area
				}

				rectCandidates = append(rectCandidates, RectCandidate{Corner1: tile1, Corner2: tile2, Area: area})
			}
		}
	}

	// Sort rectangle candidates by Area descending
	slices.SortFunc(rectCandidates, func(a, b RectCandidate) int {
		if b.Area > a.Area {
			return 1
		} else if b.Area < a.Area {
			return -1
		}
		return 0
	})

	numWorkers := 12
	candidateChan := make(chan RectCandidate, 100000)
	resultChan := make(chan float64, numWorkers)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rect := range candidateChan {
				if isValidRectangle(rect, tiles, boundaries) {
					resultChan <- rect.Area
					// Found valid rectangle, return from worker
					return
				}
			}
		}()
	}

	// Send candidates to workers
	go func() {
		for _, candidate := range rectCandidates {
			candidateChan <- candidate
		}
		close(candidateChan)
	}()

	// Wait for workers and close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for area := range resultChan {
		if area > largestAreaInsideBoundaries {
			largestAreaInsideBoundaries = area
		}
	}

	fmt.Printf("Part one, largest area of any rectange: %.0f\n", largestArea)
	fmt.Printf("Part two, largest area of any rectange inside boundaries: %.0f\n", largestAreaInsideBoundaries)
}
