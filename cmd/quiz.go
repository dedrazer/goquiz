/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type Question struct {
	Id                     int                 `json:"id"`
	Question               string              `json:"question"`
	Description            string              `json:"description"`
	Answers                map[string]string   `json:"answers"`
	MultipleCorrectAnswers string              `json:"multiple_correct_answers"`
	CorrectAnswers         map[string]string   `json:"correct_answers"`
	Explanation            string              `json:"explanation"`
	Tip                    string              `json:"tip"`
	Tags                   []map[string]string `json:"tags"`
	Category               string              `json:"category"`
	Difficulty             string              `json:"difficulty"`
}

// global variables
var quizURL string = "https://quizapi.io/api/v1/questions"
var apiToken string = "RohQw0KajLcOZHsBUSxwmPG1mULYRul2hUsm9hvZ"
var limit string = "3"
var fileName string = "scores.csv"

// quizCmd represents the quiz command
var quizCmd = &cobra.Command{
	Use:   "quiz",
	Short: "Generate a Quiz",
	Long:  `Generate a CLI quiz and collect the user's score.`,
	Run: func(cmd *cobra.Command, args []string) {
		// get a quiz from API
		quiz := getQuiz(getQuizBytes())

		// ensure singular answers
		for hasMultipleAnswers(quiz) {
			// if quiz has multiple answers get another
			// (the vast majority of questions have only 1 answer)
			quiz = getQuiz(getQuizBytes())
		}

		// load previous scores
		scores, fileContent := getAscendingScoresFromFile(fileName)

		// welcome user
		welcome()

		// start the quiz
		score := doQuiz(quiz)

		// place the user
		placeUser(score, scores)

		// save the result
		saveResult(fileContent, score)
	},
}

func init() {
	rootCmd.AddCommand(quizCmd)
}

func getQuizBytes() []byte {
	params := "apiKey=" + url.QueryEscape(apiToken) + "&" +
		"limit=" + limit
	path := fmt.Sprintf(quizURL+"?%s", params)

	response, err := http.Get(path)
	if err != nil {
		log.Printf("Could not get quiz - %v", err)
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Could not read response body - %v", err)
	}

	return responseBytes
}

func getQuiz(qb []byte) []Question {
	quiz := []Question{}

	if err := json.Unmarshal(qb, &quiz); err != nil {
		log.Printf("Could not unmarshal JSON - %v", err)
	}

	return quiz
}

func doQuiz(questions []Question) int {
	var result = make([]bool, len(questions))
	score := 0
	scanner := bufio.NewScanner(os.Stdin)

	for i, q := range questions {
		// print the question
		fmt.Println("Question", i+1, "-", q.Question+"\n")

		// answers
		c := 97
		for j := 0; j < len(q.Answers); j++ {
			char := string(c)
			a := q.Answers["answer_"+char]

			if len(a) > 0 {
				fmt.Println(char, a)
				c++
			}
		}

		scanner.Scan()
		input := scanner.Text()

		userAnswers := strings.Split(input, ",")

		if len(userAnswers) == 1 {
			// answers were not split
			// put the first element back into the slice
			userAnswers = []string{userAnswers[0]}
		}

		hasMultipleAnswers, err := strconv.ParseBool(q.MultipleCorrectAnswers)

		if err != nil {
			fmt.Printf("Could not check for multiple answers - %v", err)
			os.Exit(1)
		}

		if hasMultipleAnswers && (len(userAnswers) < 2) {
			// user did not get all the answers
			result[i] = false
			break
		}

		correctAnswers := []string{}
		for k, v := range q.CorrectAnswers {
			b, err := strconv.ParseBool(v)

			if err != nil {
				fmt.Printf("Could not parse correct answer - %v", err)
				os.Exit(1)
			}

			if b {
				// 8th char is answer letter
				correctAnswers = append(correctAnswers, string(k[7]))
			}
		}

		if len(userAnswers) != len(correctAnswers) {
			// user had an incorrect number of answers
			result[i] = false
			break
		}

		sort.Strings(correctAnswers)
		sort.Strings(userAnswers)

		fmt.Println("Correct Answer:", correctAnswers)

		correctAnswer := true
		for j, ca := range correctAnswers {
			if userAnswers[j] != ca {
				correctAnswer = false
				break
			}
		}

		if correctAnswer {
			score++
		}
		result[i] = correctAnswer
	}

	return score
}

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

	return sort.IntSlice(scores), scoresString
}

func placeUser(score int, scores []int) {
	pos := -1
	for i, s := range scores {
		if pos == -1 && s >= score {
			pos = i
		}
	}

	// user placed last
	if pos == -1 {
		pos = 0
	}

	// position + 1 due to 0-based index
	percentile := (float64(pos+1) / float64(len(scores))) * 100

	// round down to the nearest 5
	percentile = percentile - (math.Mod(percentile, 5))

	if percentile < 100 {
		percentile += 5
	}

	fmt.Printf("You placed in the top %v%%", percentile)
}

func saveResult(fileContent string, score int) {
	err := ioutil.WriteFile(fileName, []byte(fileContent+","+strconv.Itoa(score)), 0666)
	if err != nil {
		fmt.Printf("Could not save score - %v", err)
	}
}

func hasMultipleAnswers(q []Question) bool {
	for _, question := range q {
		b, err := strconv.ParseBool(question.MultipleCorrectAnswers)

		if err != nil {
			fmt.Printf("Could not check if question had multiple answers - %v", err)
		}

		if b {
			return true
		}
	}

	return false
}

func welcome() {
	fmt.Println("Welcome to the quiz!")
	fmt.Println("You will be asked a number of questions.")
	fmt.Println("Simply enter the letter which corresponds to your answer.")
	fmt.Println("< Press ENTER to continue >")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}
