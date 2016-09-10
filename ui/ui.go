package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/op/go-logging"
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
		ui.Choice("Choose: (1) Local version (2) Remote version (3) Edit (4) Abort : ")
		scanner := bufio.NewReader(os.Stdin)
		choice, _ := scanner.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "4":
			err = errors.New("Abort")
			validChoice = true
		case "3":
			fmt.Print("Enter new value: ")
			choice, _ = scanner.ReadString('\n')
			choice = strings.TrimSpace(choice)
			if choice == "" {
				fmt.Println("Warning: Empty value detected.")
			}
			confirmed := ui.YesOrNo("Confirm: " + choice)
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

// askForNewValue from user
func (ui UI) updateValue(field, oldValue string) (newValue string, err error) {
	fmt.Printf(ui.BlueBold("Modifying %s:\n"), field)
	fmt.Printf("\t%s\n", oldValue)
	fmt.Printf(ui.GreenBold("Choose: (1) Keep Value (2) Edit "))
	validChoice := false
	errs := 0
	for !validChoice {
		scanner := bufio.NewReader(os.Stdin)
		choice, _ := scanner.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "2":
			fmt.Print("Enter new value: ")
			choice, _ = scanner.ReadString('\n')
			choice = strings.TrimSpace(choice)
			if choice == "" {
				fmt.Println("Warning: Empty value detected.")
			}
			confirmed := ui.YesOrNo("Confirm: " + choice)
			if confirmed {
				newValue = choice
				validChoice = true
			} else {
				fmt.Println("Manual entry not confirmed, trying again.")
			}
		case "1":
			newValue = oldValue
			validChoice = true
		default:
			fmt.Println("Invalid choice.")
			fmt.Printf(ui.GreenBold("Choose: (1) Keep Value (2) Edit "))
			errs++
			if errs > 10 {
				return "", errors.New("Too many invalid choices.")
			}
		}
	}
	return
}

// UpdateValues from candidates or from user input
func (ui UI) UpdateValues(field, oldValue string, candidates []string) (newValues []string, err error) {
	if len(candidates) == 0 {
		value, err := ui.updateValue(field, oldValue)
		if err != nil {
			return []string{}, err
		}
		candidates = append(candidates, value)
	}
	// cleanup
	for i := range candidates {
		candidates[i] = strings.TrimSpace(candidates[i])
	}
	newValues = candidates

	// show old_value => new_value
	newValuesString := strings.Join(newValues, "|")
	if oldValue != newValuesString {

		ui.Infof("Changing %s: \n%s\n\t=>\n%s\n", field, oldValue, newValuesString)
	} else {
		ui.Info("Nothing to change.")
	}
	return
}

// YesOrNo asks a question and returns the answer
func (ui UI) YesOrNo(question string) (yes bool) {
	fmt.Printf(ui.BlueBold("%s y/n? "), question)
	scanner := bufio.NewReader(os.Stdin)
	choice, _ := scanner.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "y", "Y", "yes":
		yes = true
	}
	return
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
		cmd.Run()
	}()

	// send through less
	fmt.Fprintf(stdin, output)
	// close stdin (result in pager to exit)
	stdin.Close()
	// wait for the pager to be finished
	<-c
}
