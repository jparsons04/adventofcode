package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type turnedOnBattery struct {
	bankPos    int
	joltRating int
}

func findBestJoltRatingInBank(
	bank string,
	startPos int,
	desiredRatingStart int,
	bankBatteriesTurnedOn []turnedOnBattery,
) turnedOnBattery {
	for rating := desiredRatingStart; rating > 0; rating-- {
		for i := startPos; i < len(bank); i++ {
			currentJoltRating, _ := strconv.Atoi(string(bank[i]))
			if currentJoltRating == rating {
				batteryFound := slices.ContainsFunc(bankBatteriesTurnedOn, func(battery turnedOnBattery) bool {
					return battery.bankPos == i
				})

				// Skip to the next battery in the bank if it's already been turned on
				if batteryFound {
					continue
				}

				return turnedOnBattery{bankPos: i, joltRating: currentJoltRating}
			}
		}
	}

	return turnedOnBattery{}
}

func turnOnBatteriesInBank(
	bank string,
	numBatteriesToFind int,
) []turnedOnBattery {

	turnedOnBatteries := make([]turnedOnBattery, 0, numBatteriesToFind)

	minRating := 0
	for rating := 9; rating > 0; rating-- {
		matches := strings.Count(bank, strconv.Itoa(rating))

		if matches >= numBatteriesToFind {
			minRating = rating
			break
		}
	}

	for rating := 9; rating >= minRating; rating-- {
		for i, _ := range bank {
			currentJoltRating, _ := strconv.Atoi(string(bank[i]))
			if currentJoltRating == rating {
				skipBattery := slices.ContainsFunc(turnedOnBatteries, func(battery turnedOnBattery) bool {
					return battery.bankPos == i || (battery.bankPos >= i && battery.joltRating >= currentJoltRating)
				})

				if !skipBattery {
					turnedOnBatteries = append(turnedOnBatteries, turnedOnBattery{bankPos: i, joltRating: currentJoltRating})

					if len(turnedOnBatteries) == numBatteriesToFind {
						return turnedOnBatteries
					}
				}
			}
		}
	}

	return nil
}

func getTotalJoltage(bankBatteriesTurnedOn []turnedOnBattery) int {
	sort.Slice(bankBatteriesTurnedOn, func(i, j int) bool {
		return bankBatteriesTurnedOn[i].bankPos < bankBatteriesTurnedOn[j].bankPos
	})

	var strJoltage string

	for _, battery := range bankBatteriesTurnedOn {
		strJoltage += strconv.Itoa(battery.joltRating)
	}

	intJoltage, _ := strconv.Atoi(strJoltage)

	//if len(bankBatteriesTurnedOn) == 12 {
	//	fmt.Printf("Sorted batteries: %+v\n", bankBatteriesTurnedOn)
	//	fmt.Println(intJoltage)
	//}

	return intJoltage
}

func main() {
	path := filepath.Join("inputs/day03.txt")

	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	sc := bufio.NewScanner(f)

	partOneTotalOutputJoltage := 0
	partTwoTotalOutputJoltage := 0
	tryTwoSum := 0

	for sc.Scan() {
		bank := sc.Text()
		bankBatteriesTurnedOn := make([]turnedOnBattery, 0, 12)

		bankBatteriesTurnedOn = append(bankBatteriesTurnedOn, findBestJoltRatingInBank(bank, 0, 9, bankBatteriesTurnedOn))

		if bankBatteriesTurnedOn[0].bankPos == len(bank)-1 {
			bankBatteriesTurnedOn = append(bankBatteriesTurnedOn, findBestJoltRatingInBank(bank, 0, bankBatteriesTurnedOn[0].joltRating-1, bankBatteriesTurnedOn))
		} else {
			bankBatteriesTurnedOn = append(bankBatteriesTurnedOn, findBestJoltRatingInBank(bank, bankBatteriesTurnedOn[0].bankPos+1, bankBatteriesTurnedOn[0].joltRating, bankBatteriesTurnedOn))
		}

		fmt.Printf("Part one batteries: %+v\n", bankBatteriesTurnedOn)

		var partOneBatteries []turnedOnBattery
		partOneBatteries = append(partOneBatteries, bankBatteriesTurnedOn...)
		partOneTotalOutputJoltage += getTotalJoltage(partOneBatteries)

		tryTwo := turnOnBatteriesInBank(bank, 2)
		fmt.Printf("Try Two Batteries : %+v\n", tryTwo)
		tryTwoSum += getTotalJoltage(tryTwo)

		// Part 2: Turn on 10 more batteries per bank
		//for i := 2; i < 12; i++ {
		//	if bankBatteriesTurnedOn[i-1].bankPos == len(bank)-1 {
		//		bankBatteriesTurnedOn = append(bankBatteriesTurnedOn, findBestJoltRatingInBank(bank, 0, bankBatteriesTurnedOn[i-1].joltRating-1, bankBatteriesTurnedOn))
		//	} else {
		//		bankBatteriesTurnedOn = append(bankBatteriesTurnedOn, findBestJoltRatingInBank(bank, bankBatteriesTurnedOn[i-1].bankPos+1, 9, bankBatteriesTurnedOn))
		//	}
		//}

		if bankBatteriesTurnedOn[1].bankPos == len(bank)-1 {
			bankBatteriesTurnedOn = append(bankBatteriesTurnedOn, findBestJoltRatingInBank(bank, 0, bankBatteriesTurnedOn[1].joltRating-1, bankBatteriesTurnedOn))
		} else {
			bankBatteriesTurnedOn = append(bankBatteriesTurnedOn, findBestJoltRatingInBank(bank, bankBatteriesTurnedOn[1].bankPos+1, bankBatteriesTurnedOn[1].joltRating, bankBatteriesTurnedOn))
		}

		//fmt.Printf("Batteries: %+v\n", bankBatteriesTurnedOn)

		//add := getTotalJoltage(bankBatteriesTurnedOn)
		//fmt.Println(add)

		//partTwoTotalOutputJoltage += add

		//partTwoTotalOutputJoltage += getTotalJoltage(bankBatteriesTurnedOn)
	}

	fmt.Printf("Part one total output joltage: %d\n", partOneTotalOutputJoltage)
	fmt.Printf("Part two total output joltage: %d\n", partTwoTotalOutputJoltage)
	fmt.Printf("Try two sum: %d\n", tryTwoSum)
}
