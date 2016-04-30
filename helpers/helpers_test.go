package helpers

import (
	"fmt"
	"testing"
)

func TestHelpersChoice(t *testing.T) {
	fmt.Println("+ Testing Helpers/GetChoice()...")
	candidates := []string{"one", "two"}
	idx, userInput, err := Choose(candidates...)
	if err.Error() != "Incorrect input" || idx != -1 || userInput != "" {
		t.Errorf("Expected GetChoice to fail without input")
	}
}

func TestHelpersStringInSlice(t *testing.T) {
	fmt.Println("+ Testing Helpers/StringInSlice()...")
	candidates := []string{"one", "two"}
	idx, isIn := StringInSlice("one", candidates)
	if !isIn || idx != 0 {
		t.Errorf("Error finding string in slice")
	}
	idx, isIn = StringInSlice("One", candidates)
	if isIn || idx != -1 {
		t.Errorf("Error finding string in slice")
	}
}

func TestHelpersCSContains(t *testing.T) {
	fmt.Println("+ Testing Helpers/CaseInsensitiveContains()...")
	if !CaseInsensitiveContains("TestString", "test") {
		t.Errorf("Error, substring in string")
	}
	if !CaseInsensitiveContains("TestString", "stSt") {
		t.Errorf("Error, substring in string")
	}
	if CaseInsensitiveContains("TestString", "teest") {
		t.Errorf("Error, substring not in string")
	}
}
