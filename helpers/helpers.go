/*
Package helpers is a subpackage of Endive.

It is a mix of helper functions, for file manipulation, logging, remote API access, and display.

*/
package helpers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/barsanuphe/gotabulate"
)

// TimeTrack helps track the time taken by a function.
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logger.Debugf("-- %s in %s\n", name, elapsed)
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

// Choose displays a list of candidates and returns the user's pick
func Choose(localCandidate, remoteCandidate string) (chosenOne string, err error) {
	fmt.Printf("1. %s\n", localCandidate)
	fmt.Printf("2. %s\n", remoteCandidate)
	fmt.Printf(GreenBold("Choose: (1) Local version (2) Remote version (3) Edit (4) Abort "))

	validChoice := false
	errs := 0
	for !validChoice {
		scanner := bufio.NewReader(os.Stdin)
		choice, _ := scanner.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "4":
			err = errors.New("Abort")
			validChoice = true
		case "3":
			fmt.Printf("Enter new value: ")
			choice, _ = scanner.ReadString('\n')
			choice = strings.TrimSpace(choice)
			if choice == "" {
				fmt.Println("Warning: Empty value detected.")
			}
			confirmed := YesOrNo("Confirm: " + choice)
			if confirmed {
				chosenOne = choice
				validChoice = true
			} else {
				fmt.Println("Manual entry not confirmed, trying again.")
			}
		case "2":
			chosenOne = remoteCandidate
			validChoice = true
		case "1":
			chosenOne = localCandidate
			validChoice = true
		default:
			fmt.Println("Invalid choice.")
			errs++
			if errs > 10 {
				return "", errors.New("Too many invalid choices.")
			}
		}
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

// AskForNewValue from user
func AskForNewValue(field, oldValue string) (newValue string, err error) {
	fmt.Printf(BlueBold("Modifying %s: previous value: %s\n"), field, oldValue)
	fmt.Println(BlueBold("Enter new value: "))
	scanner := bufio.NewReader(os.Stdin)
	choice, err := scanner.ReadString('\n')
	if err != nil {
		return "", err
	}
	newValue = strings.TrimSpace(choice)
	return
}

// AssignNewValues from candidates or from user input
func AssignNewValues(field, oldValue string, candidates []string) (newValues []string, err error) {
	if len(candidates) == 0 {
		values, err := AskForNewValue(field, oldValue)
		if err != nil {
			return []string{}, err
		}
		if YesOrNo("Confirm") {
			candidates = append(candidates, values)
		} else {
			return []string{}, errors.New("No new value")
		}
	}
	// cleanup
	for i := range candidates {
		candidates[i] = strings.TrimSpace(candidates[i])
	}
	newValues = candidates
	// show old_value => new_value
	Infof("Changing %s: \n%s\n\t=>\n%s\n", field, oldValue, strings.Join(newValues, "|"))
	return
}

// ChooseVersion among two choices
func ChooseVersion(title, local, remote string) (choice string, err error) {
	Subpart(title + ":")
	choice, err = Choose(local, remote)
	return
}

// Display text through a pager if necessary.
func Display(output string) {
	// -e Causes less to automatically exit the second time it reaches end-of-file.
	// -F or --quit-if-one-screen  Causes less to automatically exit if the entire file can be displayed on the first screen.
	// -Q Causes totally "quiet" operation: the terminal bell is never rung.
	// -X or --no-init Disables sending the termcap initialization and deinitialization strings to the terminal. This is sometimes desirable if the deinitialization string does something unnecessary, like clearing the screen.
	cmd := exec.Command("less", "-e", "-F", "-Q", "-X")
	r, stdin := io.Pipe()
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// create a blocking chan, Run the pager and unblock once it is finished
	c := make(chan struct{})
	go func() {
		defer close(c)
		cmd.Run()
	}()

	// send through less
	fmt.Fprintf(stdin, output)
	// close stdin (result in pager to exit)
	stdin.Close()
	// wait for the pager to be finished
	<-c
}
