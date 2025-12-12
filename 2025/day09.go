package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type TileCoord struct {
	Col float64
	Row float64
}

func getArea(tile1, tile2 TileCoord) float64 {
	return (math.Abs(tile2.Col-tile1.Col) + 1) * (math.Abs(tile2.Row-tile1.Row) + 1)
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

	var largestArea float64

	for _, tile1 := range tiles {
		for _, tile2 := range tiles {
			if tile1 != tile2 {
				area := getArea(tile1, tile2)
				if area > largestArea {
					largestArea = area
				}
			}
		}
	}

	fmt.Printf("Part one, largest area of any rectange: %.0f\n", largestArea)
}
