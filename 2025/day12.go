package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Heavily inspired by https://github.com/lamasalah32/pentomino-tiling
// Adapted to work with the input's provided polyominos of present shapes. For
// Algorithm X, this implementation uses primary columns for present instances
// that must be covered exactly once, and secondary columns for grid positions
// that can be covered at most once.

// ========================
// Polyomino implementation

var allPresentTypes []Present

type Point struct {
	X int
	Y int
}

type Present struct {
	Index  int
	Points []Point
}

type Choice struct {
	Pos                  []int
	PresentInstanceIndex int
}

func Rotate(p Present) Present {
	newPoints := make([]Point, len(p.Points))
	for i, pt := range p.Points {
		newPoints[i] = Point{-pt.Y, pt.X}
	}

	return Present{Index: p.Index, Points: newPoints}
}

func Flip(p Present) Present {
	newPoints := make([]Point, len(p.Points))
	for i, pt := range p.Points {
		newPoints[i] = Point{-pt.X, pt.Y}
	}

	return Present{Index: p.Index, Points: newPoints}
}

// Normalize translates the present so that the minimum X and Y are both 0
func Normalize(p Present) Present {
	if len(p.Points) == 0 {
		return p
	}

	minX, minY := p.Points[0].X, p.Points[0].Y
	for _, pt := range p.Points {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
	}

	normalized := make([]Point, len(p.Points))
	for i, pt := range p.Points {
		normalized[i] = Point{X: pt.X - minX, Y: pt.Y - minY}
	}

	return Present{Index: p.Index, Points: normalized}
}

// CanonicalString creates a unique string representation for a normalized present
func CanonicalString(p Present) string {
	points := make([]Point, len(p.Points))
	copy(points, p.Points)

	// Sort points by Y first, then X for consistent ordering
	for i := range len(points) {
		for j := i + 1; j < len(points); j++ {
			if points[j].Y < points[i].Y || (points[j].Y == points[i].Y && points[j].X < points[i].X) {
				points[i], points[j] = points[j], points[i]
			}
		}
	}

	var s strings.Builder
	for _, pt := range points {
		s.WriteString(fmt.Sprintf("(%d,%d)", pt.X, pt.Y))
	}
	return s.String()
}

// GenOrientations generates all possible orientations of a present
// Each present can be rotated 90 degrees and each rotation can be flipped
// Orientations are deduplicated by normalizing and comparing canonical strings
func GenOrientations(p Present) []Present {
	var allOrientations []Present

	curr := p
	// Each present can be rotated 90 degrees (4 unique rotations) and each rotation can be flipped
	for range 4 {
		curr = Rotate(curr)
		allOrientations = append(allOrientations, curr)

		flipped := Flip(curr)
		allOrientations = append(allOrientations, flipped)
	}

	// Deduplicate orientations by normalizing and comparing canonical strings
	seen := make(map[string]bool)
	var uniqueOrientations []Present

	for _, orientation := range allOrientations {
		normalizedOrientation := Normalize(orientation)
		canonical := CanonicalString(normalizedOrientation)

		if !seen[canonical] {
			seen[canonical] = true
			uniqueOrientations = append(uniqueOrientations, orientation)
		}
	}

	return uniqueOrientations
}

// isValidPlacement checks if a present can be placed at a given position
// without exceeding the bounds of the region
func isValidPlacement(i, j, width, length int, p Present) bool {
	for pointIdx := range p.Points {
		xEnd := p.Points[pointIdx].X + i
		yEnd := p.Points[pointIdx].Y + j

		if xEnd < 0 || xEnd >= width || yEnd < 0 || yEnd >= length {
			return false
		}
	}

	return true
}

// FindPresents creates a list of presents to be used for the incidence matrix
// The present types are available in the allPresentTypes global variable
func FindPresents(presentCounts map[int]int) []Present {
	presents := make([]Present, 0)

	for presentIndex, count := range presentCounts {
		for range count {
			presents = append(presents, allPresentTypes[presentIndex])
		}
	}

	return presents
}

// GenChoices creates a list of choices to be used for the incidence matrix
// Choices are the possible placements of presents in the region
func GenChoices(width, length int, region Region) []Choice {
	choices := make([]Choice, 0)
	presents := FindPresents(region.PresentCount)

	// Cache orientations per present type to avoid redundant computation
	orientationCache := make(map[int][]Present)
	for presentIdx := range presents {
		presentType := presents[presentIdx].Index
		if _, exists := orientationCache[presentType]; !exists {
			orientationCache[presentType] = GenOrientations(presents[presentIdx])
		}
	}

	for rowIdx := range width {
		for colIdx := range length {
			for presentInstanceIdx := range presents {
				orientations := orientationCache[presents[presentInstanceIdx].Index]
				for orientationIdx := range orientations {
					if isValidPlacement(rowIdx, colIdx, width, length, orientations[orientationIdx]) {
						pos := make([]int, len(orientations[orientationIdx].Points))
						for pointIdx := range orientations[orientationIdx].Points {
							pos[pointIdx] = (orientations[orientationIdx].Points[pointIdx].Y+colIdx)*width + (orientations[orientationIdx].Points[pointIdx].X + rowIdx)
						}

						choices = append(choices, Choice{
							Pos:                  pos,
							PresentInstanceIndex: presentInstanceIdx,
						})
					}
				}
			}
		}
	}

	return choices
}

// ===============================
// Incidence matrix implementation
// Note that this matrix is used in Algorithm X to solve a slight variation of the exact cover problem
// Exact cover is not required here, so dancing links column headers are either marked as primary or secondary
// Columns for present instances must be covered exactly once
// Columns for grid positions can be covered at most once

// GenMatrix builds the incidence matrix for a given region
// The matrix is a boolean matrix where each row represents a choice and each column represents a present instance or a grid position
func GenMatrix(width, length int, region Region) [][]bool {
	choices := GenChoices(width, length, region)
	matrix := make([][]bool, len(choices))

	var presentCount int
	for _, count := range region.PresentCount {
		presentCount += count
	}

	for rowIdx := range matrix {
		matrix[rowIdx] = make([]bool, presentCount+width*length)
	}

	for choiceIdx := range choices {
		matrix[choiceIdx][choices[choiceIdx].PresentInstanceIndex] = true
		for posIdx := range choices[choiceIdx].Pos {
			matrix[choiceIdx][presentCount+choices[choiceIdx].Pos[posIdx]] = true
		}
	}

	return matrix
}

// ============================
// Dancing Links implementation
// Used for backtracking as part of utilizing Algorithm X

type Node struct {
	L, R *Node
	U, D *Node
	C    *Header
}

type Header struct {
	Node
	L, R      *Header
	S         int
	N         int
	IsPrimary bool
}

// ChooseColumn chooses the column with the smallest size that is primary
func ChooseColumn(h *Header) *Header {
	var c *Header
	smallestSize := int(^uint(0) >> 1)

	for col := h.R; col != h; col = col.R {
		if !col.IsPrimary {
			continue
		}

		if c == nil || col.S < smallestSize {
			c = col
			smallestSize = col.S
		}
	}

	return c
}

// Cover covers a column in the dancing links matrix
// It removes the column from the matrix by updating the pointers of the nodes in the column
// and the nodes in the rows that contain the column
func Cover(h *Header) {
	h.R.L = h.L
	h.L.R = h.R

	for row := h.D; row != &h.Node; row = row.D {
		for col := row.R; col != row; col = col.R {
			col.D.U = col.U
			col.U.D = col.D
			col.C.S--
		}
	}
}

// Uncover uncovers a column in the dancing links matrix
// It restores the column to the matrix by updating the pointers of the nodes in the column
// and the nodes in the rows that contain the column
func Uncover(h *Header) {
	for row := h.U; row != &h.Node; row = row.U {
		for col := row.L; col != row; col = col.L {
			col.C.S++
			col.D.U = col
			col.U.D = col
		}
	}

	h.R.L = h
	h.L.R = h
}

// BuildDLX builds the dancing links matrix for a given region
func BuildDLX(matrix [][]bool, presentLen int) *Header {
	numColumns := len(matrix[0])

	root := &Header{N: -1}
	root.L = root
	root.R = root
	root.U = &root.Node
	root.D = &root.Node
	root.C = root

	headers := make([]*Header, numColumns)
	prev := root

	for i := range numColumns {
		header := &Header{N: i}
		header.C = header
		header.S = 0
		header.U = &header.Node
		header.D = &header.Node
		header.L = prev
		header.R = root
		prev.R = header
		root.L = header
		headers[i] = header
		prev = header
		if i < presentLen {
			header.IsPrimary = true
		}
	}

	for _, row := range matrix {
		var first, last *Node

		for colIdx, val := range row {
			if val {
				header := headers[colIdx]
				node := &Node{C: header}

				node.D = &header.Node
				node.U = header.Node.U
				header.Node.U.D = node
				header.Node.U = node
				header.S++

				if first == nil {
					first = node
					last = node
					node.L = node
					node.R = node
				} else {
					node.L = last
					node.R = first
					last.R = node
					first.L = node
					last = node
				}
			}
		}
	}

	// Check for unsolvable constraints: any primary column with 0 choices
	for rootHeader := root.R; rootHeader != root; rootHeader = rootHeader.R {
		// If a required present has no valid placements, the region is unsolvable
		if rootHeader.IsPrimary && rootHeader.S == 0 {
			return nil
		}
	}

	return root
}

// SolveDLXRecursive solves the dancing links matrix using Algorithm X
// It recursively covers and uncovers columns to find a solution
func SolveDLXRecursive(h *Header, k int, solution []*Node) [][]*Node {
	// Base case: if all primary columns are covered, a solution has been found
	allPrimaryColsCovered := true
	for r := h.R; r != h; r = r.R {
		if r.IsPrimary {
			allPrimaryColsCovered = false
			break
		}
	}

	if allPrimaryColsCovered {
		solCopy := make([]*Node, len(solution))
		copy(solCopy, solution)
		return [][]*Node{solCopy}
	}

	c := ChooseColumn(h)
	Cover(c)

	for row := c.D; row != &c.Node; row = row.D {
		solution = append(solution, row)

		for col := row.R; col != row; col = col.R {
			Cover(col.C)
		}

		res := SolveDLXRecursive(h, k+1, solution)
		if len(res) > 0 {
			Uncover(c)
			return res
		}

		solution = solution[:len(solution)-1]

		for col := row.L; col != row; col = col.L {
			Uncover(col.C)
		}
	}

	Uncover(c)
	return nil
}

// SolveDLX solves the dancing links matrix using Algorithm X
func SolveDLX(h *Header) ([][]*Node, bool) {
	resultChan := make(chan [][]*Node, 1)

	go func() {
		result := SolveDLXRecursive(h, 0, nil)
		resultChan <- result
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	select {
	case result := <-resultChan:
		return result, true
	case <-ctx.Done():
		return nil, false
	}
}

// ==========================

type Region struct {
	Width        int
	Length       int
	PresentCount map[int]int
}

// ParseInput parses the input file and returns a list of regions
// The input file is a list of present shapes and regions
// The present shapes are defined by a list of points that are marked with '#'
// The regions are defined by a width and length and a list of present counts
// The present counts are the number of times each present type must be present in the region
func parseInput(sc *bufio.Scanner) []Region {
	var regions []Region

	var currentPresent Present
	var presentShape []Point
	var presentRow int

	for sc.Scan() {
		line := sc.Text()

		if !strings.Contains(line, "x") {
			// Collect the present shapes first
			if strings.Contains(line, ":") {
				presentShape = []Point{}
				presentRow = 0
				indexLine := strings.Split(line, ":")
				index, _ := strconv.Atoi(indexLine[0])
				currentPresent.Index = index
			} else if strings.Contains(line, "#") {
				for i, r := range line {
					if r == '#' {
						presentShape = append(presentShape, Point{X: i, Y: presentRow})
					}
				}

				presentRow++
			} else if len(line) == 0 {
				currentPresent.Points = presentShape
				allPresentTypes = append(allPresentTypes, currentPresent)
			}
		} else {
			// Then collect the regions
			var currentRegion Region
			currentRegionPresentCounts := make(map[int]int)

			regionLine := strings.Split(line, ":")
			regionDimensions := strings.Split(regionLine[0], "x")
			regionWidth, _ := strconv.Atoi(regionDimensions[0])
			regionLength, _ := strconv.Atoi(regionDimensions[1])
			currentRegion.Width = regionWidth
			currentRegion.Length = regionLength

			regionPresentCounts := strings.Split(strings.TrimSpace(regionLine[1]), " ")
			for i, strCount := range regionPresentCounts {
				intCount, _ := strconv.Atoi(strCount)
				currentRegionPresentCounts[i] = intCount
			}

			currentRegion.PresentCount = currentRegionPresentCounts
			regions = append(regions, currentRegion)
		}
	}

	return regions
}

func main() {
	filePath := filepath.Join("inputs/day12.txt")
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	regions := parseInput(sc)

	numWorkers := 6

	type result struct {
		regionNum   int
		hasSolution bool
		status      string
	}

	jobsChan := make(chan int, len(regions))
	resultsChan := make(chan result, len(regions))
	var wg sync.WaitGroup

	// Worker pool
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobsChan {
				region := regions[i]
				width := region.Width
				length := region.Length

				// Quick feasibility check: total polyomino cells must fit in region
				totalPolyominoCells := 0
				presentInstanceCount := 0
				for presentIdx, count := range region.PresentCount {
					if count > 0 && presentIdx < len(allPresentTypes) {
						cellsPerPresent := len(allPresentTypes[presentIdx].Points)
						totalPolyominoCells += count * cellsPerPresent
						presentInstanceCount += count
					}
				}

				regionArea := width * length

				if totalPolyominoCells > regionArea {
					resultsChan <- result{i + 1, false, "Impossible (too many cells)"}
					continue
				}

				matrix := GenMatrix(width, length, region)
				root := BuildDLX(matrix, presentInstanceCount)

				if root == nil {
					resultsChan <- result{i + 1, false, "Unsolvable (no valid placements)"}
					continue
				}

				solutions, completed := SolveDLX(root)

				if !completed {
					resultsChan <- result{i + 1, false, "Context cancelled"}
				} else if len(solutions) > 0 {
					resultsChan <- result{i + 1, true, "Solution found"}
				} else {
					resultsChan <- result{i + 1, false, "No solution"}
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for i := range regions {
			jobsChan <- i
		}
		close(jobsChan)
	}()

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var partOneValidRegions int
	for range regions {
		r := <-resultsChan
		fmt.Printf("Region %d: %s\n", r.regionNum, r.status)
		if r.hasSolution {
			partOneValidRegions++
		}
	}

	fmt.Printf("\nPart one, number of valid regions: %d\n", partOneValidRegions)
}
