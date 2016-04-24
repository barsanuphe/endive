package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var epubs = []struct {
	filename       string
	expectedSha256 string
}{
	{
		"pg16328.epub",
		"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03",
	},
	{
		"pg17989.epub",
		"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a",
	},
}

func TestHelpersListEpubs(t *testing.T) {
	fmt.Println("+ Testing Helpers/ListEpubs()...")
	testDir, err := os.Getwd()
	if err != nil {
		t.Errorf("Error getting current directory: %s", err.Error())
	}
	// go up, down to test
	testDir = filepath.Join(filepath.Dir(testDir), "test")

	epubsPaths, hashes, err := ListEpubsInDirectory(testDir)
	if err != nil {

		t.Errorf("Error listing epubs: %s", err.Error())
	}

	// using epubs as defined in epub_test
	if len(epubsPaths) != len(hashes) {
		t.Errorf("Error listing epubs: paths and hashes should have same length.")
	}
	if len(epubsPaths) != len(epubs) {
		t.Errorf("Error listing epubs: expected 2 epubs, got %d.", len(epubsPaths))
	}

	for i, path := range epubsPaths {
		relativePath, err := filepath.Rel(testDir, path)
		if err != nil {
			t.Errorf("Error: %s", err.Error())
		}
		if epubs[i].filename != relativePath {
			t.Errorf("Error:  expected %s, got %s.", epubs[i].filename, relativePath)
		}
		if epubs[i].expectedSha256 != hashes[i] {
			t.Errorf("Error:  expected %s, got %s.", epubs[i].expectedSha256, hashes[i])
		}
	}
}

func TestHelpersChoice(t *testing.T) {
	fmt.Println("+ Testing Helpers/GetChoice()...")
	candidates := []string{"one", "two"}
	idx, userInput, err := Choose(candidates...)
	if err.Error() != "Incorrect input" || idx != -1 || userInput != "" {
		t.Errorf("Expected GetChoice to fail without input")
	}

}
