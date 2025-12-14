package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type freshRange struct {
	rangeStart int
	rangeEnd   int
}

func transposeRuneOperands(matrix [][]rune) [][]rune {
	rows := len(matrix)
	cols := len(matrix[0])

	result := make([][]rune, cols)
	for i := range result {
		result[i] = make([]rune, rows)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[j][i] = matrix[i][j]
		}
	}

	return result
}

func main() {
	path := filepath.Join("inputs/day06.txt")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	runeOperands := [][]rune{}
	operands := [][]int{}
	operators := []string{}

	for sc.Scan() {
		line := sc.Text()
		lineFields := strings.Fields(line)

		if slices.Contains(lineFields, "+") || slices.Contains(lineFields, "*") {
			operators = lineFields
			break
		}

		runeLine := []rune(line)
		slices.Reverse(runeLine)
		runeOperands = append(runeOperands, []rune(runeLine))

		intLineOperands := []int{}

		for _, operand := range lineFields {
			intOperand, _ := strconv.Atoi(operand)
			intLineOperands = append(intLineOperands, intOperand)
		}

		operands = append(operands, intLineOperands)
	}

	verticalTotals := make([]int, len(operands[0]), len(operands[0]))

	for i := 0; i < len(operands); i++ {
		for j := 0; j < len(operands[i]); j++ {
			if operators[j] == "+" {
				if i == 0 {
					verticalTotals[j] = operands[i][j]
				} else {
					verticalTotals[j] += operands[i][j]
				}
			} else if operators[j] == "*" {
				if i == 0 {
					verticalTotals[j] = operands[i][j]
				} else {
					verticalTotals[j] *= operands[i][j]
				}
			}
		}
	}

	var partOneTotal int
	for _, v := range verticalTotals {
		partOneTotal += v
	}

	runeOperands = transposeRuneOperands(runeOperands)

	slices.Reverse(operators)

	reverseCol := 0
	grabFirstValue := true
	localTotal := 0
	partTwoTotal := 0

	for i := range runeOperands {
		reversedOperand := string(runeOperands[i])
		intReversedOperand, _ := strconv.Atoi(strings.ReplaceAll(reversedOperand, " ", ""))

		if intReversedOperand != 0 {
			if i == len(runeOperands)-1 {
				if operators[reverseCol] == "+" {
					localTotal += intReversedOperand
				} else if operators[reverseCol] == "*" {
					localTotal *= intReversedOperand
				}

				partTwoTotal += localTotal
				break
			}

			if grabFirstValue == true {
				localTotal = intReversedOperand
				grabFirstValue = false
			} else if operators[reverseCol] == "+" {
				localTotal += intReversedOperand
			} else if operators[reverseCol] == "*" {
				localTotal *= intReversedOperand
			}
		} else {
			partTwoTotal += localTotal
			localTotal = 0
			grabFirstValue = true
			reverseCol++
		}
	}

	fmt.Printf("Part one grand total: %d\n", partOneTotal)
	fmt.Printf("Part two grand total: %d\n", partTwoTotal)
}
