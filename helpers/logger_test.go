package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestHelpersLogger(t *testing.T) {
	fmt.Println("+ Testing Helpers/GetLogger()...")
	logFilename := "../test/testing"
	err := GetLogger(logFilename)
	defer LogFile.Close()

	Logger.Error("Error")
	Logger.Infof("Test %d%% complete", 50)
	// should not be displayed
	Logger.Debug("Debug")

	// checking log file
	output, err := ioutil.ReadFile(logFilename)
	if os.IsNotExist(err) {
		t.Errorf("Error: cannot find log file")
		t.FailNow()
	} else if err != nil {
		t.Errorf("Error reading log file: %s", err.Error())
	}
	// checking 3 lines were written + 1 for setup + 1 return at end of file
	lines := strings.Split(string(output), "\n")
	if len(lines) != 5 {
		t.Errorf("Error checking log file: wrong number of lines: %d", len(lines))
	}
	// remove log file
	err = os.Remove(logFilename)
	if err != nil {
		t.Errorf("Error removing test log file")
	}
}
