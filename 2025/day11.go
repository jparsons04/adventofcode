package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

type Graph struct {
	Devices map[string][]string
	Mutex   sync.Mutex
	Count   int
}

func (g *Graph) Increment() {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()
	g.Count++
}

func (g *Graph) bfs(src, dst string, path []string) {
	var q []string
	q = append(q, src)

	for {
		path = append(path, q[0])
		q = q[1:]

		current := path[len(path)-1]

		if current == dst {
			g.Increment()
		}

		for _, neighbor := range g.Devices[current] {
			var newPath []string
			copy(newPath, path)
			newPath = append(newPath, neighbor)
			for _, p := range newPath {
				q = append(q, p)
			}
		}

		if len(path) == 0 {
			break
		}
	}
}

func (g *Graph) dfs(src, dst string, path []string, partTwo bool) {
	path = append(path, src)

	if src == dst {
		if partTwo {
			if slices.Contains(path, "dac") && slices.Contains(path, "fft") {
				g.Increment()
			}
		} else {
			g.Increment()
		}
	} else {
		for _, neighbor := range g.Devices[src] {
			g.dfs(neighbor, dst, path, partTwo)
		}
	}

	path = path[1:]
}

func main() {
	filePath := filepath.Join("inputs/day11.txt")
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	graph := Graph{Devices: make(map[string][]string)}

	for sc.Scan() {
		device := strings.Split(sc.Text(), ":")
		deviceName := device[0]
		deviceOutputs := strings.Split(strings.TrimSpace(device[1]), " ")

		graph.Devices[deviceName] = deviceOutputs
	}

	sinks := make([]string, 0)

	for name, outputs := range graph.Devices {
		if slices.Contains(outputs, "out") {
			sinks = append(sinks, name)
		}
	}

	var path []string

	for _, sink := range sinks {
		graph.dfs("you", sink, path, false)
	}

	fmt.Printf("Part one, number of paths leading you to out: %d\n", graph.Count)

	graph.Count = 0
	path = []string{}

	var wg sync.WaitGroup

	for _, sink := range sinks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			graph.dfs("svr", sink, path, true)
		}()
	}

	wg.Wait()

	fmt.Printf("Part two, number of paths from svr to out through both dac and fft: %d\n", graph.Count)
}
