package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Heavily inspired by https://github.com/lamasalah32/pentomino-tiling
// Adapted to work with the input's provided polyominos of present shapes

// ========================
// Polyomino implementation

var allPresents []Present

var PresentIndexToIndex = make(map[int]int)
var IndexToPresentIndex = make(map[int]int)

type Point struct {
	X int
	Y int
}

type Present struct {
	Index  int
	Points []Point
}

type Choices struct {
	N   int
	Pos []int
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

func GenOrientations(p Present) []Present {
	var orientations []Present

	curr := p
	// Each present can be rotated 90 degrees and each rotation can be flipped
	for i := 0; i < 4; i++ {
		curr = Rotate(curr)
		orientations = append(orientations, curr)

		flipped := Flip(curr)
		orientations = append(orientations, flipped)
	}

	return orientations
}

func isValidPlacement(i, j, w, h int, o Present) bool {
	for k := range o.Points {
		xEnd := o.Points[k].X + i
		yEnd := o.Points[k].Y + j

		if xEnd < 0 || xEnd >= w || yEnd < 0 || yEnd >= h {
			return false
		}
	}

	return true
}

func FindPresents(presents []int) []Present {
	found := make([]Present, 0, len(presents))
	set := make(map[int]struct{})

	for _, name := range presents {
		set[name] = struct{}{}
	}

	k := 0
	for _, p := range allPresents {
		if _, ok := set[p.Index]; ok {
			found = append(found, p)
			PresentIndexToIndex[p.Index] = k
			IndexToPresentIndex[k] = p.Index
			k++
		}
	}

	return found
}

func GenChoices(w, h int, presents []int) []Choices {
	c := make([]Choices, 0)
	p := FindPresents(presents)

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			for k := range p {
				orientations := GenOrientations(p[k])
				for l := range orientations {
					if isValidPlacement(i, j, w, h, orientations[l]) {
						pos := make([]int, len(orientations[l].Points))
						for m := range orientations[l].Points {
							pos[m] = (orientations[l].Points[m].Y+j)*w + (orientations[l].Points[m].X + i)
						}

						c = append(c, Choices{N: orientations[l].Index, Pos: pos})
					}
				}
			}
		}
	}

	return c
}

// ===============================
// Incidence matrix implementation
// The matrix used in Algorithm X to solve for the exact cover problem

func GenMatrix(w, h int, presents []int) [][]bool {
	choices := GenChoices(w, h, presents)

	for i, c := range choices {
		fmt.Printf("choice %d: %+v\n", i, c)
	}

	matrix := make([][]bool, len(choices))

	for i := range matrix {
		matrix[i] = make([]bool, len(presents)+w*h)
	}

	for j := range choices {
		matrix[j][w*h+PresentIndexToIndex[choices[j].N]] = true
		for k := range choices[j].Pos {
			matrix[j][choices[j].Pos[k]] = true
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
	L, R *Header
	S    int
	N    int
}

func ChooseColumn(h *Header) *Header {
	c := h.R
	s := c.S

	for j := c.R; j != h; j = j.R {
		col := j

		if col.S < s {
			c = col
			s = col.S
		}
	}

	return c
}

func Cover(h *Header) {
	h.R.L = h.L
	h.L.R = h.R

	for i := h.D; i != &h.Node; i = i.D {
		for j := i.R; j != i; j = j.R {
			j.D.U = j.U
			j.U.D = j.D
			j.C.S--
		}
	}
}

func Uncover(h *Header) {
	for i := h.U; i != &h.Node; i = i.U {
		for j := i.L; j != i; j = j.L {
			j.C.S++
			j.D.U = j
			j.U.D = j
		}
	}

	h.R.L = h
	h.L.R = h
}

func BuildDLX(matrix [][]bool) *Header {
	w := len(matrix[0])

	root := &Header{N: -1}
	root.L = root
	root.R = root
	root.U = &root.Node
	root.D = &root.Node
	root.C = root

	headers := make([]*Header, w)
	prev := root

	for i := range w {
		h := &Header{N: i}
		h.C = h
		h.S = 0

		h.U = &h.Node
		h.D = &h.Node

		h.L = prev
		h.R = root
		prev.R = h
		root.L = h

		headers[i] = h
		prev = h
	}

	for _, row := range matrix {
		var first, last *Node

		for j, val := range row {
			if val {
				h := headers[j]
				n := &Node{C: h}

				n.D = &h.Node
				n.U = h.Node.U
				h.Node.U.D = n
				h.Node.U = n
				h.S++

				if first == nil {
					first = n
					last = n
					n.L = n
					n.R = n
				} else {
					n.L = last
					n.R = first
					last.R = n
					first.L = n
					last = n
				}
			}
		}
	}

	return root
}

func SolveDLX(h *Header, k int, solution []*Node) [][]*Node {
	if h.R == h {
		solCopy := make([]*Node, len(solution))
		copy(solCopy, solution)
		return [][]*Node{solCopy}
	}

	var res [][]*Node
	c := ChooseColumn(h)
	Cover(c)

	for r := c.D; r != &c.Node; r = r.D {
		solution = append(solution, r)

		for j := r.R; j != r; j = j.R {
			Cover(j.C)
		}

		res = append(res, SolveDLX(h, k+1, solution)...)
		solution = solution[:len(solution)-1]

		for j := r.L; j != r; j = j.L {
			Uncover(j.C)
		}
	}

	Uncover(c)
	return res
}

func PrintSolutions(width, height int, solutions [][]*Node) {
	region := make([][]string, height)
	for i := range region {
		region[i] = make([]string, width)
	}

	i := rand.Intn(len(solutions))
	sol := solutions[i]

	for _, node := range sol {
		var ch string
		for j := node; ; j = j.R {
			if j.C.N >= width*height {
				ch = strconv.Itoa(IndexToPresentIndex[j.C.N-width*height])
				break
			}

			if j.R == node {
				break
			}
		}

		for j := node; ; j = j.R {
			if j.C.N < width*height {
				pos := j.C.N
				x := pos % width
				y := pos / width
				region[y][x] = ch
			}

			if j.R == node {
				break
			}
		}
	}

	for i := range region {
		fmt.Println(region[i])
	}
}

// ==========================

type Region struct {
	Width        int
	Length       int
	PresentCount map[int]int
}

func main() {
	filePath := filepath.Join("inputs/day12-example.txt")
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

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
				allPresents = append(allPresents, currentPresent)
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

	width := 12
	height := 5
	testPresents := []int{0}
	matrix := GenMatrix(width, height, testPresents)
	root := BuildDLX(matrix)
	solutions := SolveDLX(root, 0, nil)

	fmt.Printf("len solutions: %d\n", len(solutions))

	if len(solutions) > 0 {
		PrintSolutions(width, height, solutions)
	}

	//for i, present := range presents {
	//	fmt.Printf("i: %d\n", i)
	//	fmt.Printf("present: %+v\n", present)
	//}

	//fmt.Printf("regions: %+v\n\n", regions)
}
