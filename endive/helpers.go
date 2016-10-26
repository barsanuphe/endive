package endive

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/barsanuphe/gotabulate"
	"github.com/moraes/isbn"
	"github.com/tj/go-spin"
)

// TimeTrack helps track the time taken by a function.
func TimeTrack(ui UserInterface, start time.Time, name string) {
	elapsed := time.Since(start)
	ui.Debugf("-- %s in %s\n", name, elapsed)
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

// RemoveDuplicates in []string
func RemoveDuplicates(options *[]string, otherStringsToClean ...string) {
	found := make(map[string]bool)
	// specifically remove other strings from values
	for _, o := range otherStringsToClean {
		found[o] = true
	}
	j := 0
	for i, x := range *options {
		if !found[x] && x != "" {
			found[x] = true
			(*options)[j] = (*options)[i]
			j++
		}
	}
	*options = (*options)[:j]
}

// StringInSliceCaseInsensitive checks if a string is in a []string, regardless of case.
func StringInSliceCaseInsensitive(a string, list []string) (index int, isIn bool) {
	for i, b := range list {
		if strings.ToLower(b) == strings.ToLower(a) {
			return i, true
		}
	}
	return -1, false
}

// CaseInsensitiveContains checks if a substring is in a string, regardless of case.
func CaseInsensitiveContains(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// TabulateRows of map[string]int.
func TabulateRows(rows [][]string, headers ...string) (table string) {
	if len(rows) == 0 {
		return
	}
	t := gotabulate.Create(rows)
	t.SetHeaders(headers)
	t.SetEmptyString("N/A")
	t.SetAlign("left")
	t.SetAutoSize(true)
	return t.Render("border")
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

// CleanISBN from a string
func CleanISBN(full string) (isbn13 string, err error) {
	// cleanup string, only keep numbers
	re := regexp.MustCompile("[0-9]+X?")
	candidate := strings.Join(re.FindAllString(strings.ToUpper(full), -1), "")

	// if start of isbn detected, try to salvage the situation
	if len(candidate) > 13 && strings.HasPrefix(candidate, "978") {
		candidate = candidate[:13]
	}

	// validate and convert to ISBN13 if necessary
	if isbn.Validate(candidate) {
		if len(candidate) == 10 {
			isbn13, err = isbn.To13(candidate)
			if err != nil {
				isbn13 = ""
			}
		}
		if len(candidate) == 13 {
			isbn13 = candidate
		}
	} else {
		err = errors.New("ISBN-13 not found")
	}
	return
}

// AskForISBN when not found in epub
func AskForISBN(ui UserInterface) (string, error) {
	if ui.Accept("Do you want to enter an ISBN manually") {
		validChoice := false
		errs := 0
		for !validChoice {
			fmt.Print("Enter ISBN: ")
			choice, scanErr := ui.GetInput()
			if scanErr != nil {
				return "", scanErr
			}
			// check valid ISBN
			isbnCandidate, err := CleanISBN(choice)
			if err != nil {
				errs++
				ui.Warning("Warning: Invalid value.")
			} else {
				confirmed := ui.Accept("Confirm: " + choice)
				if confirmed {
					validChoice = true
					return isbnCandidate, nil
				}
				errs++
				fmt.Println("Manual entry not confirmed, trying again.")
			}
			if errs > 5 {
				ui.Warning("Too many errors, continuing without ISBN.")
				break
			}
		}
	}
	return "", errors.New("ISBN not set")
}

//SpinWhileThingsHappen is a way to launch a function and display a spinner while it is being executed.
func SpinWhileThingsHappen(title string, f func() error) (err error) {
	c1 := make(chan bool)
	c2 := make(chan error)

	// first routine for the spinner
	ticker := time.NewTicker(time.Millisecond * 100)
	go func() {
		for range ticker.C {
			c1 <- true
		}
	}()
	// second routine deals with the function
	go func() {
		// run function
		c2 <- f()
	}()

	// await both of these values simultaneously,
	// dealing with each one as it arrives.
	functionDone := false
	s := spin.New()
	for !functionDone {
		select {
		case <-c1:
			fmt.Printf("\r%s... %s ", title, s.Next())
		case err := <-c2:
			if err != nil {
				fmt.Printf("\r%s... KO.\n", title)
				return err
			}
			fmt.Printf("\r%s... Done.\n", title)
			functionDone = true
		}
	}
	return
}
