/*
Helpers is a subpackage of Endive.

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
	"time"

	"github.com/bndr/gotabulate"
)

// TimeTrack helps track the time taken by a function.
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("-- [%s in %s]\n", name, elapsed)
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
	t := gotabulate.Create(rows)
	t.SetHeaders([]string{firstHeader, secondHeader})
	t.SetEmptyString("N/A")
	t.SetAlign("left")
	return t.Render("simple")
}

// GetChoice displays a list of candidates and returns the user's pick
func GetChoice(candidates []string) (index int, err error) {
	for i, choice := range candidates {
		fmt.Printf("%d. %s\n", i+1, choice)
	}
	fmt.Printf("Choose: [1-%d], (A)bort? ", len(candidates))
	scanner := bufio.NewScanner(os.Stdin)
	// TODO: ask for: more choices, new search (prompt for author/title manually)
	var choice string
	for scanner.Scan() {
		choice = scanner.Text()
	}
	switch choice {
	case "a", "A", "abort":
		return -1, errors.New("Abort")
	default:
		index, err := strconv.Atoi(choice)
		if err == nil {
			index -= 1
		}
		fmt.Println(index)
	}
	return
}
