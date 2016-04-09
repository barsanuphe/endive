package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestHelpersListEpubs(t *testing.T) {
	fmt.Println("+ Testing Helpers/ListEpubs()...")
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("Error getting current directory: ", err.Error())
	}

	epubsPaths, hashes, err := listEpubsInDirectory(currentDir)
	if err != nil {
		t.Errorf("Error listing epubs: ", err.Error())
	}

	// using epubs as defined in epub_test
	if len(epubsPaths) != len(hashes) {
		t.Errorf("Error listing epubs: paths and hashes should have same length.")
	}
	if len(epubsPaths) != len(epubs) {
		t.Errorf("Error listing epubs: expected 2 epubs, got %d.", len(epubsPaths))
	}

	for i, path := range epubsPaths {
		relativePath, err := filepath.Rel(currentDir, path)
		if err != nil {
			t.Errorf("Error: ", err.Error())
		}
		if epubs[i].filename != relativePath {
			t.Errorf("Error:  expected %s, got %s.", epubs[i].filename, relativePath)
		}
		if epubs[i].expectedSha256 != hashes[i] {
			t.Errorf("Error:  expected %s, got %s.", epubs[i].expectedSha256, hashes[i])
		}
	}
}
