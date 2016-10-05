package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	en "github.com/barsanuphe/endive/endive"
	"github.com/barsanuphe/endive/mock"
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
	// getting test directory
	testDir, err := os.Getwd()
	require.Nil(t, err, "Error getting current directory")
	testDir = filepath.Join(testDir, "test")
	// using epubs as defined in epub_test
	epubsPaths, err := listEpubs(testDir)
	assert.Nil(err, "Error listing epubs")
	assert.Equal(2, len(epubsPaths), "Error listing epubs: expected 2 epubs, got %d.", len(epubsPaths))
}

func TestGetCandidates(t *testing.T) {
	fmt.Println("+ Testing Import/GetCandidates()...")
	assert := assert.New(t)
	// getting test directory
	testDir, err := os.Getwd()
	require.Nil(t, err, "Error getting current directory")
	testDir = filepath.Join(testDir, "test")
	// using epubs as defined in epub_test

	h := en.KnownHashes{Filename: "test/hashes.json"}
	// adding hash for pg16328.epub
	h.Add("dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03")
	c := &mock.Collection{}

	// prepare dummy test files
	testFiles, err := PrepareTestFiles(100, testDir)
	assert.Nil(err, "Error generating test files")
	defer CleanupTestFiles(testFiles)

	candidates, err := getCandidates(testDir, h, c)
	assert.Nil(err, "Error listing epubs")
	assert.Equal(102, len(candidates), "Error listing candidates: expected 102 epubs, got %d.", len(candidates))

	for _, candidate := range candidates {
		if filepath.Base(candidate.filename) == "pg16328.epub" {
			assert.Equal("dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03", candidate.hash)
			assert.True(candidate.imported, "hash has been imported")
			assert.False(candidate.importedButMissing, "is not missing")
		}
		if filepath.Base(candidate.filename) == "pg17989.epub" {
			assert.Equal("acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a", candidate.hash)
			assert.False(candidate.imported, "hash has not been imported")
			assert.False(candidate.importedButMissing, "is not missing")
		}
		//fmt.Println(candidate)
	}
	// TODO more.
}
