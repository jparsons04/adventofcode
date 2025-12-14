package main

import (
	"bufio"
	"context"
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

type BoundarySegment struct {
	Segment                        LineSegment
	MinRow, MaxRow, MinCol, MaxCol float64
	IsVertical                     bool
}

func isOnBoundary(
	point TileCoord,
	boundaries []BoundarySegment,
	redTileMap map[TileCoord]bool) bool {
	if _, ok := redTileMap[point]; ok {
		return true
	}

	for _, boundary := range boundaries {
		if boundary.IsVertical {
			if boundary.Segment.Start.Col == point.Col && boundary.MinRow <= point.Row && point.Row <= boundary.MaxRow {
				return true
			}
		} else {
			if boundary.Segment.Start.Row == point.Row && boundary.MinCol <= point.Col && point.Col <= boundary.MaxCol {
				return true
			}
		}
	}

	return false
}

func isInside(point TileCoord, boundaries []BoundarySegment) bool {
	intersections := 0
	for _, segment := range boundaries {
		// Vertical line segments only
		if segment.IsVertical {
			segCol := segment.Segment.Start.Col
			if segCol > point.Col && segment.MinRow <= point.Row && point.Row < segment.MaxRow {
				intersections++
			}
		}
	}

	return intersections%2 == 1
}

func isValidRectangle(
	rect RectCandidate,
	boundaries []BoundarySegment,
	redTileMap map[TileCoord]bool) bool {
	perimeter := generatePerimeter([]TileCoord{rect.Corner1, rect.Corner2})

	for _, point := range perimeter {
		if !isOnBoundary(point, boundaries, redTileMap) {
			if !isInside(point, boundaries) {
				return false
			}
		}
	}

	return true
}

func getArea(tile1, tile2 TileCoord) float64 {
	return (math.Abs(tile2.Col-tile1.Col) + 1) * (math.Abs(tile2.Row-tile1.Row) + 1)
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

func buildBoundary(tiles []TileCoord) []BoundarySegment {
	segments := make([]BoundarySegment, len(tiles))
	for i := range tiles {
		nextIdx := (i + 1) % len(tiles)
		segments[i] = BoundarySegment{
			Segment:    LineSegment{Start: tiles[i], End: tiles[nextIdx]},
			MinRow:     math.Min(tiles[i].Row, tiles[nextIdx].Row),
			MaxRow:     math.Max(tiles[i].Row, tiles[nextIdx].Row),
			MinCol:     math.Min(tiles[i].Col, tiles[nextIdx].Col),
			MaxCol:     math.Max(tiles[i].Col, tiles[nextIdx].Col),
			IsVertical: tiles[i].Col == tiles[nextIdx].Col,
		}
	}
	return segments
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

	// boundaries are the line segments that make up the outline of the red and green tiles
	boundaries := buildBoundary(tiles)

	var largestArea float64
	var largestAreaInsideBoundaries float64

	rectCandidates := make([]RectCandidate, 0)

	// Iterate over pairs of tiles to evaluate all possible rectangle candidates
	for _, tile1 := range tiles {
		for _, tile2 := range tiles {
			if tile1 != tile2 {
				area := getArea(tile1, tile2)

				// For Part 1
				if area > largestArea {
					largestArea = area
				}

				// For Part 2
				rectCandidates = append(rectCandidates, RectCandidate{Corner1: tile1, Corner2: tile2, Area: area})
			}
		}
	}

	// Sort rectangle candidates by Area descending
	// The first worker to find a valid rectangle will signal all workers to stop
	// because that will be the largest valid rectangle found
	slices.SortFunc(rectCandidates, func(a, b RectCandidate) int {
		if b.Area > a.Area {
			return 1
		} else if b.Area < a.Area {
			return -1
		}
		return 0
	})

	numWorkers := 12
	candidateChan := make(chan RectCandidate, numWorkers)
	resultChan := make(chan float64, numWorkers)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	redTileMap := make(map[TileCoord]bool)
	for _, tile := range tiles {
		redTileMap[tile] = true
	}

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rect := range candidateChan {
				select {
				case <-ctx.Done():
					return
				default:
					if isValidRectangle(rect, boundaries, redTileMap) {
						resultChan <- rect.Area
						// Found valid rectangle, signal all workers to stop
						fmt.Println("Found valid rectangle, signaling all workers to stop")
						cancel()
						return
					}
				}
			}
		}()
	}

	fmt.Printf("Total rectangle candidates: %d\n", len(rectCandidates))

	// Send candidates to workers
	go func() {
		for i, candidate := range rectCandidates {
			select {
			case <-ctx.Done():
				close(candidateChan)
				return
			case candidateChan <- candidate:
				if i%1000 == 0 {
					fmt.Printf("Queued %d candidates...\n", i)
				}
			}
		}
		close(candidateChan)
	}()

	// Wait for workers and close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results for Part 2
	for area := range resultChan {
		if area > largestAreaInsideBoundaries {
			largestAreaInsideBoundaries = area
		}
	}

	fmt.Printf("Part one, largest area of any rectange: %.0f\n", largestArea)
	fmt.Printf("Part two, largest area of any rectange inside boundaries: %.0f\n", largestAreaInsideBoundaries)
}
