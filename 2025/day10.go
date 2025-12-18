package main

import (
	"bufio"
	"fmt"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Machine struct {
	Buttons             []Button
	LightState          map[int]bool
	DesiredLightState   map[int]bool
	DesiredJoltageState []JoltageCounter
}

type Button struct {
	PositionsAffected []int
}

type JoltageCounter struct {
	TargetValue int
}

type LinearSolution struct {
	ButtonExpressions   []ButtonExpression
	FreeVariableIndices []int
}

type ButtonExpression struct {
	Constant     float64
	Coefficients map[int]float64
	IsFree       bool
}

const epsilon = 1e-9

func isZero(val float64) bool {
	return math.Abs(val) < epsilon
}

func isNonZero(val float64) bool {
	return math.Abs(val) >= epsilon
}

type CoefficientMatrix map[int]map[int]int
type AugmentedMatrix [][]float64

// buildCoefficientMatrix builds a coefficient matrix where each row is a
// joltage level counter and each column is a button that affects that joltage
// level counter position
func buildCoefficientMatrix(buttonList []Button, targetJoltageState []JoltageCounter) CoefficientMatrix {
	coefficientMatrix := make(CoefficientMatrix)

	// Each row is a joltage level counter
	for i := range targetJoltageState {
		coefficientMatrix[i] = make(map[int]int)

		// Each column is a button that affects the joltage level counter position
		for j, button := range buttonList {
			if slices.Contains(button.PositionsAffected, i) {
				coefficientMatrix[i][j] = 1
			} else {
				coefficientMatrix[i][j] = 0
			}
		}
	}

	return coefficientMatrix
}

// buildAugmentedMatrix transforms the coefficient matrix into an augmented matrix
// to prepare it for Gaussian elimination
func buildAugmentedMatrix(
	coefMatrix CoefficientMatrix,
	targets []JoltageCounter,
	buttonList []Button,
) AugmentedMatrix {
	unsolvedButtons := []Button{}
	unsatisfiedCounterIndices := []int{}
	originalToAugMatrixCol := make(map[int]int)

	// Build button mappings
	matrixCol := 0
	for i, button := range buttonList {
		unsolvedButtons = append(unsolvedButtons, button)
		originalToAugMatrixCol[i] = matrixCol
		matrixCol++
	}

	// Collect unsatisfied counter indices
	for i := range targets {
		unsatisfiedCounterIndices = append(unsatisfiedCounterIndices, i)
	}

	// Create the augmented matrix
	augmentedMatrix := make(AugmentedMatrix, len(unsatisfiedCounterIndices))
	for i := range augmentedMatrix {
		augmentedMatrix[i] = make([]float64, len(unsolvedButtons)+1)
	}

	// Fill augmented matrix from coefficient matrix
	for augRow, originalCounterIdx := range unsatisfiedCounterIndices {
		for originalButtonIdx, buttonIndex := range originalToAugMatrixCol {
			augmentedMatrix[augRow][buttonIndex] = float64(coefMatrix[originalCounterIdx][originalButtonIdx])
		}
		augmentedMatrix[augRow][len(unsolvedButtons)] = float64(targets[originalCounterIdx].TargetValue)
	}

	return augmentedMatrix
}

// findPivotRow finds the first row in the matrix that has a non-zero value in the
// given column. If no such row is found, it returns -1.
func findPivotRow(matrix AugmentedMatrix, startRow, col int) int {
	for row := startRow; row < len(matrix); row++ {
		if isNonZero(matrix[row][col]) {
			return row
		}
	}

	return -1
}

// scaleRow scales the given row so that the pivot value in the given column is 1
func scaleRow(matrix AugmentedMatrix, row, pivotCol int) {
	pivot := matrix[row][pivotCol]

	if isZero(pivot) {
		fmt.Printf("Pivot is zero, cannot scale row: %f\n", pivot)
		return
	}

	for col := range matrix[row] {
		matrix[row][col] /= pivot
	}
}

// eliminateBelow eliminates the entries below the pivot in the given column
// by subtracting a multiple of the pivot row from the target row
// to make the entry below the pivot zero
func eliminateBelow(matrix AugmentedMatrix, pivotRow, targetRow, col int) {
	multiplier := matrix[targetRow][col] / matrix[pivotRow][col]
	for col := range matrix[targetRow] {
		matrix[targetRow][col] -= (multiplier * matrix[pivotRow][col])

		// Clean up any rounding errors
		if isZero(matrix[targetRow][col]) {
			matrix[targetRow][col] = 0.0
		}
	}
}

func swapRows(matrix AugmentedMatrix, row1, row2 int) {
	matrix[row1], matrix[row2] = matrix[row2], matrix[row1]
}

// forwardElimination creates pivots in the augmented matrix by finding the first
// non-zero entry in each column and making it the pivot, which is then used to
// eliminate the entries below the pivot in that column
func forwardElimination(matrix AugmentedMatrix) []int {
	pivotColumns := []int{}
	pivotRow := 0
	// Exclude the target joltage value column on the right hand side
	numCols := len(matrix[0]) - 1

	for col := range numCols {
		// Find the pivot for this column
		pivotRowIdx := findPivotRow(matrix, pivotRow, col)

		if pivotRowIdx == -1 {
			// No pivot found, this column is a free variable
			continue
		}

		// Swap if needed
		if pivotRowIdx != pivotRow {
			swapRows(matrix, pivotRow, pivotRowIdx)
		}

		// Scale the pivot row so that the pivot = 1
		scaleRow(matrix, pivotRow, col)

		// Eliminate all entries below the pivot
		for row := pivotRow + 1; row < len(matrix); row++ {
			if isNonZero(matrix[row][col]) {
				eliminateBelow(matrix, pivotRow, row, col)
			}
		}

		// Record this column as having a pivot
		pivotColumns = append(pivotColumns, col)
		pivotRow++

		if pivotRow >= len(matrix) {
			break
		}
	}

	return pivotColumns
}

// eliminateAbove eliminates the entries above the pivot in the given column by
// subtracting a multiple of the pivot row from the target row to make the
// entry above the pivot zero
func eliminateAbove(matrix AugmentedMatrix, pivotRow, targetRow, col int) {
	eliminateBelow(matrix, pivotRow, targetRow, col)
}

// backSubstitution solves for the free variables of the augmented matrix by
// eliminating the entries above the pivots in each column. The resulting
// matrix after this step will be in Reduced Row Echelon Form (RREF).
func backSubstitution(matrix AugmentedMatrix, pivotColumns []int) {
	// Process pivot columns from right to left
	for i := len(pivotColumns) - 1; i >= 0; i-- {
		col := pivotColumns[i]
		pivotRow := i

		// Eliminate all entries above the pivot
		for row := range pivotRow {
			if isNonZero(matrix[row][col]) {
				eliminateAbove(matrix, pivotRow, row, col)
			}
		}
	}
}

// extractSolution extracts a linear solution from the augmented matrix by identifying
// the free variables and expressing the dependent variables in terms of the free variables
func extractSolution(matrix AugmentedMatrix, pivotColumns []int, numButtons int) LinearSolution {
	freeVariableIndices := []int{}
	buttonExpressions := make([]ButtonExpression, numButtons)

	for i := range numButtons {
		if !slices.Contains(pivotColumns, i) {
			freeVariableIndices = append(freeVariableIndices, i)
			buttonExpressions[i].IsFree = true
		}
	}

	for pivotRowIdx, pivotColIdx := range pivotColumns {
		buttonExpressions[pivotColIdx].IsFree = false
		buttonExpressions[pivotColIdx].Constant = matrix[pivotRowIdx][numButtons]
		buttonExpressions[pivotColIdx].Coefficients = make(map[int]float64)

		for _, freeVariableIndex := range freeVariableIndices {
			buttonExpressions[pivotColIdx].Coefficients[freeVariableIndex] = -matrix[pivotRowIdx][freeVariableIndex]
		}
	}

	return LinearSolution{
		ButtonExpressions:   buttonExpressions,
		FreeVariableIndices: freeVariableIndices,
	}
}

// calculateTotalButtonPresses calculates the total number of button presses required
// to reach the desired joltage levels by evaluating the linear solution and substituting
// the free variable values into the button expressions
func calculateTotalButtonPresses(
	solution LinearSolution,
	freeVariableValues map[int]int,
) int {
	totalButtonPresses := 0

	// Add the button presses from the button expressions in the augmented matrix
	for buttonIndex, buttonExpr := range solution.ButtonExpressions {
		if buttonExpr.IsFree {
			if freeVariableValues == nil {
				panic("Free variable found but no free variable values provided")
			}
			totalButtonPresses += freeVariableValues[buttonIndex]
		} else {
			pressCount := buttonExpr.Constant

			// Only evaluate coefficients if free variable values are provided
			if freeVariableValues != nil {
				for freeVariableIdx, coeff := range buttonExpr.Coefficients {
					pressCount += coeff * float64(freeVariableValues[freeVariableIdx])
				}
			}

			totalButtonPresses += int(math.Round(pressCount))
		}
	}

	return totalButtonPresses
}

func isValidSolution(solution LinearSolution, freeVariableValues map[int]int) bool {
	for buttonIndex, expr := range solution.ButtonExpressions {
		var pressCount float64

		if expr.IsFree {
			// Press count for free variables come directly from freeVariableValues
			pressCount = float64(freeVariableValues[buttonIndex])
		} else {
			// Press count for dependent variables is evaluated from the expression
			pressCount = expr.Constant
			for freeVariableIdx, coeff := range expr.Coefficients {
				// Skip coefficients that are effectively zero
				if !isZero(coeff) {
					pressCount += coeff * float64(freeVariableValues[freeVariableIdx])
				}
			}
		}

		// Does this button have a negative press count?
		if pressCount < -epsilon {
			return false
		}

		// Does this button have a non-integer press count?
		if math.Abs(pressCount-math.Round(pressCount)) > epsilon {
			return false
		}
	}

	return true
}

func enumerateRecursive(
	solution LinearSolution,
	buttonList []Button,
	freeVariableIndex int,
	freeVariableValues map[int]int,
) int {
	// Base case, when all free variables have been assigned values
	if freeVariableIndex == len(solution.FreeVariableIndices) {
		valid := isValidSolution(solution, freeVariableValues)

		// Validate that all button press counts are non-negative
		if !valid {
			return math.MaxInt
		}

		return calculateTotalButtonPresses(solution, freeVariableValues)
	}

	currentFreeVariable := solution.FreeVariableIndices[freeVariableIndex]
	minValue := 0
	maxValue := 0

	// For single free variable, calculate tight bounds
	// For multiple free variables, use conservative heuristic
	if len(solution.FreeVariableIndices) == 1 {
		// Calculate exact bounds for the single free variable
		for _, buttonExpr := range solution.ButtonExpressions {
			if !buttonExpr.IsFree {
				if coeff, exists := buttonExpr.Coefficients[currentFreeVariable]; exists && !isZero(coeff) {
					bound := -buttonExpr.Constant / coeff

					if coeff > 0 {
						// x >= -buttonExpr.Constant / coeff
						minValue = max(minValue, int(math.Ceil(bound)))
					} else if coeff < 0 {
						// x <= -buttonExpr.Constant / coeff
						if maxValue == 0 {
							maxValue = int(math.Floor(bound))
						} else {
							maxValue = min(maxValue, int(math.Floor(bound)))
						}
					}
				}
			}
		}

		// If no upper bound was set, use a fallback based on constants
		if maxValue == 0 || maxValue < minValue {
			absMax := 0
			for _, buttonExpr := range solution.ButtonExpressions {
				if !buttonExpr.IsFree {
					absConst := int(math.Ceil(math.Abs(buttonExpr.Constant)))
					if absConst > absMax {
						absMax = absConst
					}
				}
			}
			maxValue = max(minValue+absMax*2, 100)
		}
	} else {
		// Use conservative range based on constants for multiple free variables
		absMax := 0
		for _, buttonExpr := range solution.ButtonExpressions {
			if !buttonExpr.IsFree {
				absConst := int(math.Ceil(math.Abs(buttonExpr.Constant)))
				if absConst > absMax {
					absMax = absConst
				}
			}
		}

		// Scale by number of free variables to account for interdependencies
		maxValue = absMax * 2

		// Cap based on number of free variables to keep runtime reasonable
		switch len(solution.FreeVariableIndices) {
		case 2:
			maxValue = min(maxValue, 300)
		case 3:
			maxValue = min(maxValue, 100)
		default:
			maxValue = min(maxValue, 50)
		}

		// Ensure minimum search range
		maxValue = max(maxValue, 20)
	}

	minPresses := math.MaxInt

	for i := minValue; i <= maxValue; i++ {
		newFreeVariableValues := make(map[int]int)
		maps.Copy(newFreeVariableValues, freeVariableValues)
		newFreeVariableValues[currentFreeVariable] = i

		presses := enumerateRecursive(solution, buttonList, freeVariableIndex+1, newFreeVariableValues)
		if presses < minPresses {
			minPresses = presses
		}
	}

	return minPresses
}

// optimizeSolution optimizes the linear solution by enumerating all possible combinations
// of the ranges of the free variables to try to find the minimum number of button presses
// to satisfy the linear solution
func optimizeSolution(
	solution LinearSolution,
	buttonList []Button,
) int {
	// If there are no free variables to solve for, return the total number of button presses directly
	if len(solution.FreeVariableIndices) == 0 {
		return calculateTotalButtonPresses(solution, nil)
	}

	// Enumerate all possible combinations of the ranges of the free variables recursively
	// Even if bounds are invalid or can't be calculated, enumerateRecursive will use fallback ranges
	return enumerateRecursive(solution, buttonList, 0, make(map[int]int))
}

func getIndicatorLightDiagram(s string) string {
	i := strings.Index(s, "[")
	if i >= 0 {
		j := strings.Index(s, "]")
		if j >= 0 {
			return s[i+1 : j]
		}
	}

	return ""
}

func getMachineButtons(s string) []Button {
	machineButtons := make([]Button, 0)

	for {
		i := strings.Index(s, "(")
		if i >= 0 {
			j := strings.Index(s, ")")
			if j >= 0 {
				positionsAffected := strings.Split(s[i+1:j], ",")
				intPositionsAffected := make([]int, len(positionsAffected))

				for i, v := range positionsAffected {
					val, _ := strconv.Atoi(v)
					intPositionsAffected[i] = int(val)
				}

				button := Button{PositionsAffected: intPositionsAffected}
				machineButtons = append(machineButtons, button)
			}

			if !strings.Contains(s[j+1:], "(") {
				break
			} else {
				s = s[j+1:]
			}
		}
	}

	return machineButtons
}

func getDesiredJoltageLevels(s string) []JoltageCounter {
	targetJoltageLevels := make([]JoltageCounter, 0)

	i := strings.Index(s, "{")
	if i >= 0 {
		j := strings.Index(s, "}")
		if j >= 0 {
			joltageLevels := strings.Split(s[i+1:j], ",")

			for _, v := range joltageLevels {
				val, _ := strconv.Atoi(v)
				targetJoltageLevels = append(targetJoltageLevels, JoltageCounter{
					TargetValue: val,
				})
			}
		}
	}

	return targetJoltageLevels
}

func pressButtonForLights(machineLightState map[int]bool, buttonLights []int) map[int]bool {
	for _, light := range buttonLights {
		machineLightState[light] = !machineLightState[light]
	}

	return machineLightState
}

func buttonCombinationsWithRepetition(n int, buttonList []Button) [][]Button {
	if n == 0 {
		return [][]Button{nil}
	}

	if len(buttonList) == 0 {
		return nil
	}

	r := buttonCombinationsWithRepetition(n, buttonList[1:])
	for _, x := range buttonCombinationsWithRepetition(n-1, buttonList) {
		r = append(r, append(x, buttonList[0]))
	}

	return r
}

func solvePartOne(machine Machine) int {
	partOneTotalButtonPresses := 0
	buttonChoose := 1

LightLoop:
	for {
		buttonCombinations := buttonCombinationsWithRepetition(buttonChoose, machine.Buttons)
		initialLightState := make(map[int]bool)

		for _, buttons := range buttonCombinations {
			initialLightState = maps.Clone(machine.LightState)

			lightResult := make(map[int]bool)

			for _, button := range buttons {
				lightResult = pressButtonForLights(initialLightState, button.PositionsAffected)
			}

			isDesiredLightState := true

			for i := range lightResult {
				if machine.DesiredLightState[i] != lightResult[i] {
					isDesiredLightState = false
					break
				}
			}

			if isDesiredLightState {
				partOneTotalButtonPresses += buttonChoose
				break LightLoop
			}
		}

		buttonChoose++
	}

	return partOneTotalButtonPresses
}

// solvePartTwo solves Part Two using Gaussian elimination to express the
// system as a linear combination of free variables, then enumerates possible
// values to find the minimum number of button presses. Note that this solution
// does not simplify the matrix by finding forced variables before Gaussian
// elimination.  When a forced variable affects multiple columns, solving it
// greedily can produce globally suboptimal solutions.
func solvePartTwo(machine Machine) int {
	coefficientMatrix := buildCoefficientMatrix(machine.Buttons, machine.DesiredJoltageState)
	augmentedMatrix := buildAugmentedMatrix(coefficientMatrix, machine.DesiredJoltageState, machine.Buttons)
	pivotColumns := forwardElimination(augmentedMatrix)
	backSubstitution(augmentedMatrix, pivotColumns)
	linearSolution := extractSolution(augmentedMatrix, pivotColumns, len(augmentedMatrix[0])-1)
	minPresses := optimizeSolution(linearSolution, machine.Buttons)
	if minPresses == math.MaxInt {
		return 0
	}

	return minPresses
}

func main() {
	path := filepath.Join("inputs/day10.txt")
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	var partOneTotalButtonPresses, partTwoTotalButtonPresses int

	for sc.Scan() {
		var machine Machine
		machineLine := sc.Text()

		machineIndicatorLightDiagram := getIndicatorLightDiagram(machineLine)

		machineLightState := make(map[int]bool)
		desiredLightState := make(map[int]bool)

		for i, light := range machineIndicatorLightDiagram {
			machineLightState[i] = false

			if light == '#' {
				desiredLightState[i] = true
			} else {
				desiredLightState[i] = false
			}
		}

		machine.Buttons = getMachineButtons(machineLine)

		machine.DesiredLightState = desiredLightState
		machine.LightState = machineLightState

		machine.DesiredJoltageState = getDesiredJoltageLevels(machineLine)

		partOneTotalButtonPresses += solvePartOne(machine)
		partTwoTotalButtonPresses += solvePartTwo(machine)
	}

	fmt.Printf("Part one, fewest button presses to correctly configure lights: %d\n", partOneTotalButtonPresses)
	fmt.Printf("Part two, fewest button presses to correctly configure joltage: %d\n", partTwoTotalButtonPresses)
}
