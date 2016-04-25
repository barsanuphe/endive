/*
Package helpers is a subpackage of Endive.

It is a mix of helper functions, for file manipulation, logging, remote API access, and display.

*/
package helpers

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bndr/gotabulate"
	"github.com/ttacon/chalk"
)

// TimeTrack helps track the time taken by a function.
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	Logger.Debugf("-- %s in %s\n", name, elapsed)
}

// StringInSlice checks if a string is in a []string.
func StringInSlice(a string, list []string) (index int, isIn bool) {
	for i, b := range list {
		if b == a {
			return i, true
		}
	}
	return -1, false
}

// TabulateRows of map[string]int.
func TabulateRows(rows [][]string, firstHeader string, secondHeader string) (table string) {
	if len(rows) == 0 {
		return
	}
	t := gotabulate.Create(rows)
	t.SetHeaders([]string{firstHeader, secondHeader})
	t.SetEmptyString("N/A")
	t.SetAlign("left")
	// wrapping
	t.SetMaxCellSize(80)
	t.SetWrapStrings(true)
	return t.Render("simple")
}

// TabulateMap of map[string]int.
func TabulateMap(input map[string]int, firstHeader string, secondHeader string) (table string) {
	if len(input) == 0 {
		return
	}
	// building first column list for sorting
	var keys []string
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	// building table
	var rows [][]string
	for _, key := range keys {
		rows = append(rows, []string{key, strconv.Itoa(input[key])})
	}
	return TabulateRows(rows, firstHeader, secondHeader)
}

// Choose displays a list of candidates and returns the user's pick
func Choose(candidates ...string) (index int, userInput string, err error) {
	for i, choice := range candidates {
		fmt.Printf("%d. %s\n", i+1, choice)
	}
	fmt.Printf(GreenBold("Choose: [1-%d], (E)nter manually, (A)bort? "), len(candidates))
	scanner := bufio.NewReader(os.Stdin)
	choice, _ := scanner.ReadString('\n')
	choice = strings.TrimSpace(choice)
	switch choice {
	case "a", "A", "abort":
		return -1, "", errors.New("Abort")
	case "e", "E", "enter":
		fmt.Printf("Enter new value: ")
		choice, _ = scanner.ReadString('\n')
		choice = strings.TrimSpace(choice)
		confirmed := YesOrNo("Confirm: " + choice)
		if confirmed {
			userInput = choice
			index = -1
		} else {
			fmt.Println("Manual entry not confirmed. Sticking with original value.")
			index = 0
		}
	default:
		index, err = strconv.Atoi(choice)
		if err != nil {
			err = errors.New("Incorrect input")
		}
		index--
	}
	return
}

// YesOrNo asks a question and returns the answer
func YesOrNo(question string) (yes bool) {
	fmt.Printf(BlueBold("%s (y)/(n)? "), question)
	scanner := bufio.NewReader(os.Stdin)
	choice, _ := scanner.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "y", "Y", "yes":
		yes = true
	}
	return
}

// BlueBold outputs a string in blue bold.
func BlueBold(in string) (out string) {
	return chalk.Bold.TextStyle(chalk.Blue.Color(in))
}

// GreenBold outputs a string in green bold.
func GreenBold(in string) (out string) {
	return chalk.Bold.TextStyle(chalk.Green.Color(in))
}
