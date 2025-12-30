package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Graph struct {
	Devices map[string][]string
}

func countPaths(node, dst string, graph map[string][]string, memo map[string]map[string]int) int {
	if node == dst {
		return 1
	}

	if _, ok := memo[node]; !ok {
		memo[node] = make(map[string]int)
	}

	if count, exists := memo[node][dst]; exists {
		return count
	}

	total := 0
	for _, neighbor := range graph[node] {
		total += countPaths(neighbor, dst, graph, memo)
	}

	memo[node][dst] = total
	return total
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

	// Memoization map to store the number of paths between two nodes
	pathCount := make(map[string]map[string]int)

	// Part one
	pathCount["you"]["out"] = countPaths("you", "out", graph.Devices, pathCount)

	// Part two
	// svr -> dac -> fft -> out
	pathCount["svr"]["dac"] = countPaths("svr", "dac", graph.Devices, pathCount)
	pathCount["dac"]["fft"] = countPaths("dac", "fft", graph.Devices, pathCount)
	pathCount["fft"]["out"] = countPaths("fft", "out", graph.Devices, pathCount)

	// svr -> fft -> dac -> out
	pathCount["svr"]["fft"] = countPaths("svr", "fft", graph.Devices, pathCount)
	pathCount["fft"]["dac"] = countPaths("fft", "dac", graph.Devices, pathCount)
	pathCount["dac"]["out"] = countPaths("dac", "out", graph.Devices, pathCount)

	partTwoCount := (pathCount["svr"]["dac"] * pathCount["dac"]["fft"] * pathCount["fft"]["out"]) +
		(pathCount["svr"]["fft"] * pathCount["fft"]["dac"] * pathCount["dac"]["out"])

	fmt.Printf("Part one, number of paths leading you to out: %d\n", pathCount["you"]["out"])
	fmt.Printf("Part two, number of paths from svr to out through both dac and fft: %d\n", partTwoCount)
}
