package main

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//type SplitterNode struct {
//	Value                  int
//	LeftNode               *SplitterNode
//	RightNode              *SplitterNode
//	FoundSplitterNodeLater bool
//}

type SplitterNode struct {
	Value                  int
	Neighbors              []int
	FoundSplitterNodeLater bool
}

func bfs(splitterNodes map[int]*SplitterNode, root, target int) string {

	// check root node and target exist in the tree
	rootNode, rootExists := splitterNodes[root]
	_, targetExists := splitterNodes[target]
	if !rootExists || !targetExists {
		return ""
	}

	q := list.New()
	q.PushBack(rootNode)

	parents := make(map[int]int)
	parents[root] = -1

	for q.Len() > 0 {
		currentNode := q.Front().Value.(*SplitterNode)
		q.Remove(q.Front())

		if currentNode.Value == target {
			fmt.Printf("currentNode at target: %+v\n", currentNode)
			var route []string
			for currentNode.Value > -1 {
				route = append([]string{strconv.Itoa(currentNode.Value)}, route...)
				currentNode.Value = parents[currentNode.Value]
			}

			return strings.Join(route, "-")
		}

		for _, neighbor := range currentNode.Neighbors {
			// Will not track visited
			parents[neighbor] = currentNode.Value
			q.PushBack(splitterNodes[neighbor])
		}

		fmt.Printf("parents: %+v\n", parents)

	}

	return ""
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
	splitterNodes := map[int]*SplitterNode{}
	partOneSplitNum := 0
	var start int
	var destinations []int

	for sc.Scan() {
		line := []rune(sc.Text())

		var splitters []int

		for i, rune := range line {
			if rune == 'S' {
				fmt.Printf("Found S: %d\n", i)
				start = i
				if _, ok := tachyonBeamPositions[i]; !ok {
					tachyonBeamPositions[i] = true
					splitterNodes[i] = &SplitterNode{Value: i}
				}
			}

			if rune == '^' {
				if _, ok := tachyonBeamPositions[i]; ok {
					partOneSplitNum++
					delete(tachyonBeamPositions, i)

					if _, ok := tachyonBeamPositions[i-1]; !ok {
						tachyonBeamPositions[i-1] = true
					}

					if _, ok := tachyonBeamPositions[i+1]; !ok {
						tachyonBeamPositions[i+1] = true
					}

					if _, ok := splitterNodes[i-1]; !ok {
						splitterNodes[i-1] = &SplitterNode{Value: i}
						splitterNodes[i].Neighbors = append(splitterNodes[i].Neighbors, i-1)
						splitterNodes[i].FoundSplitterNodeLater = true
					}

					if _, ok := splitterNodes[i+1]; !ok {
						splitterNodes[i+1] = &SplitterNode{Value: i}
						splitterNodes[i].Neighbors = append(splitterNodes[i].Neighbors, i+1)
						splitterNodes[i].FoundSplitterNodeLater = true
					}

					splitters = append(splitters, i)
				}
			}
		}

		if len(splitters) > 0 {
			destinations = splitters
		}
	}

	fmt.Println(start)
	fmt.Println(destinations)

	bfs(splitterNodes, start, destinations[0])

	partTwoTimelines := 0

	fmt.Printf("Part one number of splits: %d\n", partOneSplitNum)
	fmt.Printf("Part two number of timelines: %d\n", partTwoTimelines)
}
