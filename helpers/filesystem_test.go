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

func TestHelpersIsDirEmpty(t *testing.T) {
	fmt.Println("+ Testing Helpers/Filesystem...")
	// testing on non-empty dir
	currentdir, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get current directory")
	}
	isEmpty, err := IsDirectoryEmpty(currentdir)
	if err != nil {
		t.Errorf("Error opening current directory")
	}
	if isEmpty {
		t.Errorf("Current directory is not empty")
	}
	if !DirectoryExists(currentdir) {
		t.Errorf("Current directory exists")
	}
	// testing on non existing dir
	nonExistingDir := filepath.Join(currentdir, "doesnotexist")
	isEmpty, err = IsDirectoryEmpty(nonExistingDir)
	if err == nil {
		t.Errorf("Non existing directory should have triggered error")
	}
	if DirectoryExists(nonExistingDir) {
		t.Errorf("Directory does not exist")
	}
	// testing on empty dir
	err = os.Mkdir(nonExistingDir, 0755)
	if err != nil {
		t.Errorf("Could not get create directory")
	}
	isEmpty, err = IsDirectoryEmpty(nonExistingDir)
	if err != nil {
		t.Errorf("Existing directory should not have triggered error")
	}
	if !isEmpty {
		t.Errorf("Directory should be empty")
	}
	if !DirectoryExists(nonExistingDir) {
		t.Errorf("Directory now exists")
	}
	err = os.Remove(nonExistingDir)
	if err != nil {
		t.Errorf("Could not get remove directory")
	}
}

var paths = []struct {
	path              string
	expectedCleanPath string
}{
	{
		`a/b\\j`,
		"a-b--j",
	},
	{
		".a/a",
		"a-a",
	},
}

func TestHelpersCleanForPath(t *testing.T) {
	fmt.Println("+ Testing Helpers/CleanForPath()...")
	for _, el := range paths {
		if CleanForPath(el.path) != el.expectedCleanPath {
			t.Errorf("Error cleaning path, got %s, expected %s", CleanForPath(el.path), el.expectedCleanPath)
		}
	}
}

func TestHelpersCopy(t *testing.T) {
	fmt.Println("+ Testing Helpers/Copy()...")

	// go up, down to test
	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get current directory")
		t.FailNow()
	}
	testDir := filepath.Join(filepath.Dir(currentDir), "test")

	for _, el := range epubs {
		// copy file to _test
		origFilename := filepath.Join(testDir, el.filename)
		copyFilename := filepath.Join(testDir, el.filename+"_test")
		err := CopyFile(origFilename, copyFilename)
		if err != nil {
			t.Errorf("Could not copy file %s: %s", origFilename, err.Error())
		}
		// check copy exists
		absolutePath, err := FileExists(copyFilename)
		if err != nil {
			t.Errorf("Copy file %s should exist, got %s", copyFilename, err.Error())
		}
		if absolutePath != copyFilename {
			t.Errorf("Copy path should be %s, got %s", copyFilename, absolutePath)
		}
		// check copy hash
		copyHash, err := CalculateSHA256(copyFilename)
		if err != nil {
			t.Errorf("Could not get hash from copy file %s, got %s", copyFilename, err.Error())
		}
		if copyHash != el.expectedSha256 {
			t.Errorf("Copy hash %s different from source %s", copyHash, el.expectedSha256)
		}
		// delete _test
		err = os.Remove(copyFilename)
		if err != nil {
			t.Errorf("Copy file %s could not be removed, got %s", copyFilename, err.Error())
		}
	}
}

func TestHelpersDeleteFolders(t *testing.T) {
	fmt.Println("+ Testing Helpers/DeleteEmptyFolders()...")
	// go up, down to test
	/*currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("Could not get current directory")
		t.FailNow()
	}
	testDir := filepath.Join(filepath.Dir(currentDir), "test")*/
	// TODO
	// create folders
	// test
	// remove all
}
