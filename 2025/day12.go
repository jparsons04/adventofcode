package main

import (
	"bufio"
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

// ===============================
// Incidence matrix implementation
// Note that this matrix is used in Algorithm X to solve a slight variation of the exact cover problem
// Exact cover is not required here, so dancing links column headers are either marked as primary or secondary
// Columns for present instances must be covered exactly once
// Columns for grid positions can be covered at most once

// SparseRow represents a row in the incidence matrix, but instead of storing a
// boolean for each column, we store the index of the column that is true.
type SparseRow struct {
	TrueColumns   []int
	GridPositions []int
	PresentIdx    int
}

// ============================
// Dancing Links implementation
// Used for backtracking as part of utilizing Algorithm X

type Node struct {
	L, R     *Node
	U, D     *Node
	C        *Header
	RowIndex int
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

// BuildDLXSparse builds the dancing links matrix for a given list of sparse rows
// which are built in BuildDLXStreamed.
// The primary columns are the present instances, and the secondary columns are the
// grid positions that will be occupied.
func BuildDLXSparse(sparseRows []SparseRow, numPrimaryColumns, numColumns int) *Header {
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
		if i < numPrimaryColumns {
			header.IsPrimary = true
		}
	}

	for rowIdx, row := range sparseRows {
		var rowStart *Node

		for _, colIdx := range row.TrueColumns {
			node := &Node{
				C:        headers[colIdx],
				RowIndex: rowIdx,
			}

			node.U = headers[colIdx].U
			node.D = &headers[colIdx].Node
			node.U.D = node
			node.D.U = node
			headers[colIdx].S++

			if rowStart == nil {
				rowStart = node
				node.L = node
				node.R = node
			} else {
				node.L = rowStart.L
				node.R = rowStart
				node.L.R = node
				node.R.L = node
			}
		}
	}

	return root
}

var globalOrientationCache map[int][]Present

func InitializeOrientationCache() {
	globalOrientationCache = make(map[int][]Present)
	for _, present := range allPresentTypes {
		globalOrientationCache[present.Index] = GenOrientations(present)
	}
}

// BuildDLXStreamed builds the dancing links matrix for a given region For each
// valid placement of a present, a sparse row is added to the matrix. The
// primary column is the present instance, and the secondary columns are the
// grid positions that will be occupied.
func BuildDLXStreamed(region Region) (*Header, []SparseRow) {
	width := region.Width
	length := region.Length

	sparseRows := make([]SparseRow, 0)
	presents := FindPresents(region.PresentCount)

	for yIdx := range length {
		for xIdx := range width {
			for presentInstanceIdx := range presents {
				orientations := globalOrientationCache[presents[presentInstanceIdx].Index]
				for _, orientation := range orientations {
					if isValidPlacement(xIdx, yIdx, width, length, orientation) {
						trueColumns := make([]int, 0, len(orientation.Points)+1)
						trueColumns = append(trueColumns, presentInstanceIdx)

						gridPositions := make([]int, 0, len(orientation.Points))

						for pointIdx := range orientation.Points {
							gridPos := (orientation.Points[pointIdx].Y+yIdx)*width +
								(orientation.Points[pointIdx].X + xIdx)

							gridPositions = append(gridPositions, gridPos)
							trueColumns = append(trueColumns, len(presents)+gridPos)
						}

						sparseRow := SparseRow{
							TrueColumns:   trueColumns,
							PresentIdx:    presents[presentInstanceIdx].Index,
							GridPositions: gridPositions,
						}

						sparseRows = append(sparseRows, sparseRow)
					}
				}
			}
		}
	}

	presentCount := len(presents)
	numColumns := presentCount + (width * length)
	return BuildDLXSparse(sparseRows, presentCount, numColumns), sparseRows
}

// SearchState represents the state of the DLX search at a particular point
type SearchState struct {
	column     *Header
	currentRow *Node
	solution   []*Node
	phase      int // 0 = trying rows, 1 = backtracking
}

// SolveDLXIterative solves the dancing links matrix using Algorithm X iteratively
func SolveDLXIterative(h *Header) [][]*Node {
	// Base case: if all primary columns are covered, a solution has been found
	allPrimaryColsCovered := true
	for r := h.R; r != h; r = r.R {
		if r.IsPrimary {
			allPrimaryColsCovered = false
			break
		}
	}

	if allPrimaryColsCovered {
		return [][]*Node{{}}
	}

	// Initialize with first column choice
	initialColumn := ChooseColumn(h)
	if initialColumn == nil {
		return nil
	}

	Cover(initialColumn)
	stack := []SearchState{{
		column:     initialColumn,
		currentRow: initialColumn.D,
		solution:   []*Node{},
		phase:      0,
	}}

	for len(stack) > 0 {
		state := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if state.phase == 1 {
			// Backtracking phase: uncover row columns in reverse order
			for col := state.currentRow.L; col != state.currentRow; col = col.L {
				Uncover(col.C)
			}
			continue
		}

		// Check if we've exhausted all rows in this column
		if state.currentRow == &state.column.Node {
			// Uncover this column and backtrack
			Uncover(state.column)
			continue
		}

		// Try this row: add to solution and cover affected columns
		newSolution := append([]*Node(nil), state.solution...)
		newSolution = append(newSolution, state.currentRow)

		// Cover all other columns in this row (noting that state.column is already covered)
		for col := state.currentRow.R; col != state.currentRow; col = col.R {
			Cover(col.C)
		}

		// Check if solution is complete
		allPrimaryColsCovered = true
		for r := h.R; r != h; r = r.R {
			if r.IsPrimary {
				allPrimaryColsCovered = false
				break
			}
		}

		if allPrimaryColsCovered {
			// Found a solution, so uncover everything in reverse order
			for col := state.currentRow.L; col != state.currentRow; col = col.L {
				Uncover(col.C)
			}

			Uncover(state.column)

			return [][]*Node{newSolution}
		}

		// Push two states: backtrack state first, then the state to uncover the row columns we just covered.
		// 1. State to try next row in current column (for backtracking)
		stack = append(stack, SearchState{
			column:     state.column,
			currentRow: state.currentRow.D,
			solution:   state.solution,
			phase:      0,
		})

		// 2. State to uncover the row columns we just covered.
		stack = append(stack, SearchState{
			column:     state.column,
			currentRow: state.currentRow,
			solution:   state.solution,
			phase:      1,
		})

		// Choose next column and push its state (will run first due to stack order)
		nextColumn := ChooseColumn(h)
		if nextColumn != nil {
			Cover(nextColumn)
			stack = append(stack, SearchState{
				column:     nextColumn,
				currentRow: nextColumn.D,
				solution:   newSolution,
				phase:      0,
			})
		} else {
			// No column to choose - solution would be complete, but we already checked above
			// This shouldn't happen, but if it does, uncover what we covered
			for col := state.currentRow.L; col != state.currentRow; col = col.L {
				Uncover(col.C)
			}
		}
	}

	return nil
}

// ====================================================
// Reconstructing and rendering the solution from Nodes

type Placement struct {
	PresentIdx    int
	GridPositions []int
}

func ReconstructSolution(solution []*Node, sparseRows []SparseRow, regionNum int) []Placement {
	placements := make([]Placement, len(solution))

	for i, node := range solution {
		row := sparseRows[node.RowIndex]
		placements[i] = Placement{
			PresentIdx:    row.PresentIdx,
			GridPositions: row.GridPositions,
		}
	}

	return placements
}

// ANSI color codes
var colors = []string{
	"\033[31m", // Red
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[34m", // Blue
	"\033[35m", // Magenta
	"\033[36m", // Cyan
	"\033[91m", // Bright Red
	"\033[92m", // Bright Green
	"\033[93m", // Bright Yellow
	"\033[94m", // Bright Blue
	"\033[95m", // Bright Magenta
	"\033[96m", // Bright Cyan
}

const reset = "\033[0m"

func RenderSolution(placements []Placement, width, height int) {
	// Initialize the empty grid
	grid := make([][]string, height)
	for i := range grid {
		grid[i] = make([]string, width)
		for j := range grid[i] {
			grid[i][j] = "."
		}
	}

	runes := []string{"@", "#", "$", "%", "^", "&", "*", "+", "?", "/"}

	for i, placement := range placements {
		color := colors[i%len(colors)]
		symbol := fmt.Sprintf("%s%v%s", color, runes[placement.PresentIdx%len(runes)], reset)

		for _, gridPos := range placement.GridPositions {
			row := gridPos / width
			col := gridPos % width
			grid[row][col] = symbol
		}
	}

	// Print the grid
	for _, row := range grid {
		for _, cell := range row {
			fmt.Print(cell)
		}
		fmt.Println()
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
				index, err := strconv.Atoi(indexLine[0])
				if err != nil {
					panic(err)
				}
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
			regionWidth, err := strconv.Atoi(regionDimensions[0])
			if err != nil {
				panic(err)
			}
			regionLength, err := strconv.Atoi(regionDimensions[1])
			if err != nil {
				panic(err)
			}
			currentRegion.Width = regionWidth
			currentRegion.Length = regionLength

			regionPresentCounts := strings.Split(strings.TrimSpace(regionLine[1]), " ")
			for i, strCount := range regionPresentCounts {
				intCount, err := strconv.Atoi(strCount)
				if err != nil {
					panic(err)
				}
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

	// Initialize the present orientation cache before processing the regions
	InitializeOrientationCache()

	numWorkers := 4

	type result struct {
		regionNum   int
		hasSolution bool
		status      string
		solution    []*Node
		sparseRows  []SparseRow
		width       int
		length      int
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
				for presentIdx, count := range region.PresentCount {
					if count > 0 && presentIdx < len(allPresentTypes) {
						cellsPerPresent := len(allPresentTypes[presentIdx].Points)
						totalPolyominoCells += count * cellsPerPresent
					}
				}

				regionArea := width * length

				if totalPolyominoCells > regionArea {
					resultsChan <- result{i + 1, false, "Impossible (too many cells)", nil, nil, 0, 0}
					continue
				}

				root, sparseRows := BuildDLXStreamed(region)

				if root == nil {
					resultsChan <- result{i + 1, false, "Unsolvable (no valid placements)", nil, nil, 0, 0}
					continue
				}

				solutions := SolveDLXIterative(root)

				if len(solutions) > 0 {
					resultsChan <- result{i + 1, true, "Solution found", solutions[0], sparseRows, width, length}
				} else {
					resultsChan <- result{i + 1, false, "No solution", nil, nil, 0, 0}
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
			placements := ReconstructSolution(r.solution, r.sparseRows, r.regionNum)
			RenderSolution(placements, r.width, r.length)
		}
	}

	fmt.Printf("\nPart one, number of valid regions: %d\n", partOneValidRegions)
}
