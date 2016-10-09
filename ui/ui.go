package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"io/ioutil"

	"github.com/op/go-logging"
)

const (
	editOrKeep    = "(1) Keep Value (2) Edit: "
	chooseOrEdit  = "(1) Local version (2) Remote version (3) Edit (4) Abort: "
	enterNewValue = "Enter new value: "
	invalidChoice = "Invalid choice."
	emptyValue    = "Empty value detected."
	notConfirmed  = "Manual entry not confirmed, trying again."
)

// UI implements endive.UserInterface
type UI struct {
	// Logger provides a logger to both stdout and a log file (for debug).
	logger *logging.Logger
	// LogFile is the pointer to the log file, to be closed by the main function.
	logFile *os.File
}

// Choose among two choices
func (ui UI) Choose(title, help, local, remote string) (string, error) {
	ui.SubPart(title)
	if help != "" {
		fmt.Println(ui.Green(help))
	}
	return ui.chooseVersion(local, remote)
}

// chooseVersion displays a list of candidates and returns the user's pick
func (ui UI) chooseVersion(localCandidate, remoteCandidate string) (chosenOne string, err error) {
	fmt.Printf("1. %s\n", localCandidate)
	fmt.Printf("2. %s\n", remoteCandidate)

	validChoice := false
	errs := 0
	for !validChoice {
		ui.Choice(chooseOrEdit)
		choice, scanErr := ui.GetInput()
		if scanErr != nil {
			return chosenOne, scanErr
		}
		switch choice {
		case "4":
			err = errors.New("Abort")
			validChoice = true
		case "3":
			ui.Choice(enterNewValue)
			choice, scanErr := ui.GetInput()
			if scanErr != nil {
				return chosenOne, scanErr
			}
			if choice == "" {
				ui.Warning(emptyValue)
			}
			confirmed := ui.Accept("Confirm: " + choice)
			if confirmed {
				chosenOne = choice
				validChoice = true
			} else {
				fmt.Println(notConfirmed)
			}
		case "2":
			chosenOne = remoteCandidate
			validChoice = true
		case "1":
			chosenOne = localCandidate
			validChoice = true
		default:
			ui.Warning(invalidChoice)
			errs++
			if errs > 10 {
				return "", errors.New(invalidChoice)
			}
		}
	}
	return
}

// askForNewValue from user
func (ui UI) updateValue(field, oldValue string, longField bool) (newValue string, err error) {
	ui.SubPart("Modifying " + field)
	fmt.Printf("Current value: %s\n", oldValue)
	ui.Choice(editOrKeep)
	validChoice := false
	errs := 0
	for !validChoice {
		choice, scanErr := ui.GetInput()
		if scanErr != nil {
			return newValue, scanErr
		}
		switch choice {
		case "2":
			var choice string
			var scanErr error
			if longField {
				choice, scanErr = ui.Edit(oldValue)
			} else {
				ui.Choice(enterNewValue)
				choice, scanErr = ui.GetInput()
			}
			if scanErr != nil {
				return newValue, scanErr
			}
			if choice == "" {
				ui.Warning(emptyValue)
			} else {
				fmt.Printf("New value:\n%s\n", choice)
			}
			confirmed := ui.Accept("Confirm")
			if confirmed {
				newValue = choice
				validChoice = true
			} else {
				fmt.Println(notConfirmed)
				ui.Choice(editOrKeep)
			}
		case "1":
			newValue = oldValue
			validChoice = true
		default:
			ui.Warning(invalidChoice)
			ui.Choice(editOrKeep)
			errs++
			if errs > 10 {
				return "", errors.New(invalidChoice)
			}
		}
	}
	return
}

// UpdateValues from candidates or from user input
func (ui UI) UpdateValues(field, oldValue string, candidates []string, longField bool) ([]string, error) {
	if len(candidates) == 0 {
		value, err := ui.updateValue(field, oldValue, longField)
		if err != nil {
			return []string{}, err
		}
		candidates = append(candidates, value)
	}
	// cleanup
	for i := range candidates {
		candidates[i] = strings.TrimSpace(candidates[i])
	}
	return candidates, nil
}

// GetInput from user
func (ui UI) GetInput() (string, error) {
	scanner := bufio.NewReader(os.Stdin)
	choice, scanErr := scanner.ReadString('\n')
	return strings.TrimSpace(choice), scanErr
}

// Accept asks a question and returns the answer
func (ui UI) Accept(question string) bool {
	fmt.Printf(ui.BlueBold("%s Y/N : "), question)
	input, err := ui.GetInput()
	if err == nil {
		switch input {
		case "y", "Y", "yes":
			return true
		}
	}
	return false
}

// Display text through a pager if necessary.
func (ui UI) Display(output string) {
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
		err := cmd.Run()
		if err != nil {
			ui.Error(err.Error())
			return
		}
	}()

	// send through less
	_, err := fmt.Fprintf(stdin, output)
	if err != nil {
		ui.Error(err.Error())
		return
	}
	// close stdin (result in pager to exit)
	err = stdin.Close()
	if err != nil {
		ui.Error(err.Error())
		return
	}
	// wait for the pager to be finished
	<-c
}

// Edit long value using external $EDITOR
func (ui *UI) Edit(oldValue string) (string, error) {
	// create temp file
	content := []byte(oldValue)
	tmpfile, err := ioutil.TempFile("", "edit")
	if err != nil {
		return oldValue, err
	}
	// clean up
	defer tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	// write input string inside
	if _, err := tmpfile.Write(content); err != nil {
		return oldValue, err
	}
	if err := tmpfile.Close(); err != nil {
		return oldValue, err
	}

	// find $EDITOR
	editor := os.Getenv("EDITOR")
	if editor == "" {
		ui.Warning("$EDITOR not set, falling back to nano")
		editor = "nano"
	}

	// open it with $EDITOR
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return oldValue, err
	}

	// read file back, set output string
	newContent, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		return oldValue, err
	}
	return strings.TrimSpace(string(newContent)), nil
}
