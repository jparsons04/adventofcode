package main

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
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

type SpatialIndex struct {
	VerticalByCol       map[float64][]BoundarySegment // Vertical segments indexed by column
	HorizontalByRow     map[float64][]BoundarySegment // Horizontal segments indexed by row
	AllVerticalSegments []BoundarySegment             // Pre-filtered list of all vertical segments for ray casting
}

// isOnBoundary checks if a red tile is on the boundary of the loop
func isOnBoundary(
	point TileCoord,
	spatialIndex SpatialIndex,
	redTileMap map[TileCoord]bool) bool {
	if _, ok := redTileMap[point]; ok {
		return true
	}

	// Check vertical segments at this column
	if verticalSegs, ok := spatialIndex.VerticalByCol[point.Col]; ok {
		for _, boundary := range verticalSegs {
			if boundary.MinRow <= point.Row && point.Row <= boundary.MaxRow {
				return true
			}
		}
	}

	// Check horizontal segments at this row
	if horizontalSegs, ok := spatialIndex.HorizontalByRow[point.Row]; ok {
		for _, boundary := range horizontalSegs {
			if boundary.MinCol <= point.Col && point.Col <= boundary.MaxCol {
				return true
			}
		}
	}

	return false
}

// isInside checks if a red tile is inside the loop by ray casting
// from the point to the right and counting the number of intersections
func isInside(point TileCoord, spatialIndex SpatialIndex) bool {
	intersections := 0

	// Only check vertical segments to the right of the point
	for _, segment := range spatialIndex.AllVerticalSegments {
		segCol := segment.Segment.Start.Col
		if segCol > point.Col && segment.MinRow <= point.Row && point.Row < segment.MaxRow {
			intersections++
		}
	}

	return intersections%2 == 1
}

// isValidRectangle checks if a rectangle candidate is completely contained within the loop
func isValidRectangle(
	rect RectCandidate,
	spatialIndex SpatialIndex,
	redTileMap map[TileCoord]bool) bool {
	perimeter := generatePerimeter([]TileCoord{rect.Corner1, rect.Corner2})

	for _, point := range perimeter {
		if !isOnBoundary(point, spatialIndex, redTileMap) {
			if !isInside(point, spatialIndex) {
				return false
			}
		}
	}

	return true
}

func getArea(tile1, tile2 TileCoord) float64 {
	return (math.Abs(tile2.Col-tile1.Col) + 1) * (math.Abs(tile2.Row-tile1.Row) + 1)
}

// generatePerimeter generates the tile coordinates that make up the perimeter of a rectangle
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

// buildBoundary builds the boundary line segments that make up the outline of the red and green tiles
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

// buildSpatialIndex builds a spatial index for the loop's boundaries
// for faster boundary checks
func buildSpatialIndex(boundaries []BoundarySegment) SpatialIndex {
	index := SpatialIndex{
		VerticalByCol:       make(map[float64][]BoundarySegment),
		HorizontalByRow:     make(map[float64][]BoundarySegment),
		AllVerticalSegments: make([]BoundarySegment, 0),
	}

	for _, seg := range boundaries {
		if seg.IsVertical {
			col := seg.Segment.Start.Col
			index.VerticalByCol[col] = append(index.VerticalByCol[col], seg)
			index.AllVerticalSegments = append(index.AllVerticalSegments, seg)
		} else {
			row := seg.Segment.Start.Row
			index.HorizontalByRow[row] = append(index.HorizontalByRow[row], seg)
		}
	}

	return index
}

func main() {
	path := filepath.Join("inputs/day09.txt")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	tiles := make([]TileCoord, 0)

	for sc.Scan() {
		tile := strings.Split(sc.Text(), ",")
		posFloat := make([]float64, len(tile))

		for i, v := range tile {
			val, err := strconv.Atoi(v)
			if err != nil {
				panic(err)
			}
			posFloat[i] = float64(val)
		}

		tiles = append(tiles, TileCoord{Col: posFloat[0], Row: posFloat[1]})
	}

	// boundaries are the line segments that make up the outline of the red and green tiles
	boundaries := buildBoundary(tiles)

	// Build spatial index for faster boundary checks
	spatialIndex := buildSpatialIndex(boundaries)

	var largestArea float64
	var largestAreaInsideBoundaries float64

	rectCandidates := make([]RectCandidate, 0)

	// Iterate over pairs of tiles to evaluate all possible rectangle candidates
	for i, tile1 := range tiles {
		for j, tile2 := range tiles {
			if i <= j {
				continue
			}

			area := getArea(tile1, tile2)

			// For Part 1
			if area > largestArea {
				largestArea = area
			}

			// Check if any other red tile is strictly inside this rectangle
			hasInteriorRedTile := false
			minCol := math.Min(tile1.Col, tile2.Col)
			maxCol := math.Max(tile1.Col, tile2.Col)
			minRow := math.Min(tile1.Row, tile2.Row)
			maxRow := math.Max(tile1.Row, tile2.Row)

			for k, tile3 := range tiles {
				if k == i || k == j {
					// Skip the corners
					continue
				}

				// Check if tile3 is strictly inside the rectangle
				if minCol < tile3.Col && tile3.Col < maxCol && minRow < tile3.Row && tile3.Row < maxRow {
					hasInteriorRedTile = true
					// One interior red tile is enough to make the rectangle invalid
					break
				}
			}

			if !hasInteriorRedTile {
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

	numWorkers := runtime.NumCPU()
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
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rect := range candidateChan {
				select {
				case <-ctx.Done():
					return
				default:
					if isValidRectangle(rect, spatialIndex, redTileMap) {
						resultChan <- rect.Area
						// Found valid rectangle, signal all workers to stop
						cancel()
						return
					}
				}
			}
		}()
	}

	// Send candidates to workers
	go func() {
		for _, candidate := range rectCandidates {
			select {
			case <-ctx.Done():
				close(candidateChan)
				return
			default:
				candidateChan <- candidate
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
	fmt.Printf("Part two, largest area of any rectangle inside boundaries: %.0f\n", largestAreaInsideBoundaries)
}
