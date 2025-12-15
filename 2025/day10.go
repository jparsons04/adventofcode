package main

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Machine struct {
	LightState          map[int]bool
	DesiredLightState   map[int]bool
	Buttons             []Button
	JoltageCounters     map[int]int
	DesiredJoltageState map[int]int
}

type Button struct {
	LightsAffected []int
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
				lightsAffected := strings.Split(s[i+1:j], ",")
				intLightsAffected := make([]int, len(lightsAffected))

				for i, v := range lightsAffected {
					val, _ := strconv.Atoi(v)
					intLightsAffected[i] = int(val)
				}

				button := Button{LightsAffected: intLightsAffected}
				machineButtons = append(machineButtons, button)
			}

			if strings.Index(s[j+1:], "(") == -1 {
				break
			} else {
				s = s[j+1:]
			}
		}
	}

	return machineButtons
}

func getDesiredJoltageLevels(s string) map[int]int {
	desiredJoltageLevels := make(map[int]int)

	i := strings.Index(s, "{")
	if i >= 0 {
		j := strings.Index(s, "}")
		if j >= 0 {
			joltageLevels := strings.Split(s[i+1:j], ",")

			for i, v := range joltageLevels {
				val, _ := strconv.Atoi(v)
				desiredJoltageLevels[i] = val
			}
		}
	}

	return desiredJoltageLevels
}

func pressButtonForLights(machineLightState map[int]bool, buttonLights []int) map[int]bool {
	for _, light := range buttonLights {
		machineLightState[light] = !machineLightState[light]
	}

	return machineLightState
}

func pressButtonForJoltage(machineJoltageCounters map[int]int, buttonJoltages []int) map[int]int {
	for _, light := range buttonJoltages {
		machineJoltageCounters[light] += 1
	}

	return machineJoltageCounters
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

func main() {
	//path := filepath.Join("inputs/day10-example.txt")
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

		machine.LightState = machineLightState
		machine.DesiredLightState = desiredLightState
		machine.Buttons = getMachineButtons(machineLine)

		machine.DesiredJoltageState = getDesiredJoltageLevels(machineLine)
		machine.JoltageCounters = make(map[int]int)
		for i := range len(machine.DesiredJoltageState) {
			machine.JoltageCounters[i] = 0
		}

		// Part 1
		buttonChoose := 1

	LightLoop:
		for {
			buttonCombinations := buttonCombinationsWithRepetition(buttonChoose, machine.Buttons)
			initialLightState := make(map[int]bool)

			for _, buttons := range buttonCombinations {
				initialLightState = maps.Clone(machine.LightState)

				lightResult := make(map[int]bool)

				for _, button := range buttons {
					lightResult = pressButtonForLights(initialLightState, button.LightsAffected)
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

		// Part 2
		buttonChoose = 1

		fmt.Printf("Part 2 machine: %+v\n", machine)

	JoltageLoop:
		for {
			fmt.Printf("buttonChoose: %d\n", buttonChoose)
			buttonCombinations := buttonCombinationsWithRepetition(buttonChoose, machine.Buttons)
			initialJoltageCounters := make(map[int]int)

			for _, buttons := range buttonCombinations {
				initialJoltageCounters = maps.Clone(machine.JoltageCounters)

				joltageCounterResult := make(map[int]int)

				for _, button := range buttons {
					joltageCounterResult = pressButtonForJoltage(initialJoltageCounters, button.LightsAffected)
				}

				isDesiredJoltageCounters := true

				for i := range joltageCounterResult {
					if machine.DesiredJoltageState[i] != joltageCounterResult[i] {
						isDesiredJoltageCounters = false
						break
					}
				}

				if isDesiredJoltageCounters {
					fmt.Printf("Joltage counters match for buttonChoose: %d\n", buttonChoose)
					partTwoTotalButtonPresses += buttonChoose
					break JoltageLoop
				}
			}

			buttonChoose++
		}
	}

	fmt.Printf("Part one, fewest button presses to correctly configure lights: %d\n", partOneTotalButtonPresses)
	fmt.Printf("Part two, fewest button presses to correctly configure joltage: %d\n", partTwoTotalButtonPresses)
}
