package endive

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// create a lot of epub files with random contents
const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func RandFile(root string) (filename string, err error) {
	filename = filepath.Join(root, string(RandBytes(64))+".epub")
	// create file, write random things inside.
	err = ioutil.WriteFile(filename, RandBytes(2048), 0777)
	return
}

func PrepareTestFiles(n int, root string) (testFiles []string, err error) {
	testFiles = make([]string, n)
	for i := 0; i < n; i++ {
		// NOTE: assuming here there are no filename collisions
		testFiles[i], err = RandFile(root)
		if err != nil {
			return
		}
	}
	return
}

func CleanupTestFiles(epubs []string) error {
	// remove files
	for _, f := range epubs {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

//----------------

func TestListEpubs(t *testing.T) {
	fmt.Println("+ Testing Import/ListEpubs()...")
	assert := assert.New(t)
	// using epubs as defined in epub_test
	epubsPaths, err := listEpubs("../test")
	assert.Nil(err, "Error listing epubs")
	assert.Equal(4, len(epubsPaths), "Error listing epubs: expected 4 epubs, got %d.", len(epubsPaths))
}

func TestScanForEpubs(t *testing.T) {
	fmt.Println("+ Testing Import/ScanForEpubs()...")
	assert := assert.New(t)
	// getting test directory
	testDir := "../test"
	// using epubs as defined in epub_test

	h := KnownHashes{Filename: "test/hashes.json"}
	// adding hash for pg16328.epub
	h.Add("dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03")
	var c Collection

	// prepare dummy test files
	testFiles, err := PrepareTestFiles(100, testDir)
	assert.Nil(err, "Error generating test files")
	defer CleanupTestFiles(testFiles)

	// non existing root
	_, err = ScanForEpubs("does not exist", h, c)
	assert.NotNil(err, "impossible to get candidates from inexistant directory")
	// inspecting test directory
	candidates, err := ScanForEpubs(testDir, h, c)
	assert.Nil(err, "Error listing epubs")
	assert.Equal(104, len(candidates), "Error listing candidates: expected 102 epubs, got %d.", len(candidates))

	for _, candidate := range candidates {
		if filepath.Base(candidate.Filename) == "pg16328.epub" {
			assert.Equal("dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03", candidate.Hash)
			assert.True(candidate.Imported, "hash has been imported")
			assert.False(candidate.ImportedButMissing, "is not missing")
		}
		if filepath.Base(candidate.Filename) == "pg17989.epub" {
			assert.Equal("acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a", candidate.Hash)
			assert.False(candidate.Imported, "hash has not been imported")
			assert.False(candidate.ImportedButMissing, "is not missing")
		}
	}

	// only one candidate is seen as already imported
	assert.Equal(103, len(EpubCandidates(candidates).New()))
	assert.Equal(0, len(EpubCandidates(candidates).Missing()))
	assert.Equal(103, len(EpubCandidates(candidates).Importable()))
}
