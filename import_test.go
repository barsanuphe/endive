package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	en "github.com/barsanuphe/endive/endive"
	"github.com/barsanuphe/endive/mock"
)

func TestListEpubs(t *testing.T) {
	fmt.Println("+ Testing Import/ListEpubs()...")
	assert := assert.New(t)
	// getting test directory
	testDir, err := os.Getwd()
	require.Nil(t, err, "Error getting current directory")
	testDir = filepath.Join(testDir, "test")
	// using epubs as defined in epub_test
	epubsPaths, err := ListEpubs(testDir)
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

	candidates, err := GetCandidates(testDir, h, c)
	assert.Nil(err, "Error listing epubs")
	assert.Equal(2, len(candidates), "Error listing candidates: expected 2 epubs, got %d.", len(candidates))


	for _, candidate := range candidates {
		fmt.Println(candidate)
	}
	// TODO more.

}
