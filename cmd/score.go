package cmd

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

func getAscendingScoresFromFile(fileName string) ([]int, string) {
	scoresBytes, err := ioutil.ReadFile(fileName)

	if err != nil {
		fmt.Printf("Could not open scores file - %v", err)
		os.Exit(1)
	}

	scoresString := string(scoresBytes)
	scoresStrings := strings.Split(scoresString, ",")

	scores := make([]int, len(scoresStrings))
	// convert scores to numeric
	for i, s := range scoresStrings {
		scores[i], err = strconv.Atoi(strings.Trim(s, " "))

		if err != nil {
			fmt.Printf("Could not convert score string to int - %v", err)
		}
	}

	// put scores in ascending order
	sort.Ints(scores)

	return scores, scoresString
}

func placeUser(score int, ascendingScores []int) {
	pos := 0
	for i, s := range ascendingScores {
		if s <= score {
			pos = i
		}
	}

	// position + 1 due to 0-based index
	percentile := 100 - (float64(pos+1)/float64(len(ascendingScores)))*100

	// round down to the nearest 5
	percentile = percentile - (math.Mod(percentile, 5))

	if percentile < 100 {
		percentile += 5
	}

	fmt.Printf("You got %v answer(s) correct and placed in the top %v%%", score, percentile)
}

func saveResult(fileContent string, score int) {
	err := ioutil.WriteFile(fileName, []byte(fileContent+","+strconv.Itoa(score)), 0666)
	if err != nil {
		fmt.Printf("Could not save score - %v", err)
	}
}
