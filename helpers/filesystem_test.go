package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert := assert.New(t)
	// getting test directory
	testDir, err := os.Getwd()
	require.Nil(t, err, "Error getting current directory")
	testDir = filepath.Join(filepath.Dir(testDir), "test")
	// using epubs as defined in epub_test
	epubsPaths, hashes, err := ListEpubsInDirectory(testDir)
	assert.Nil(err, "Error listing epubs")
	assert.Equal(len(epubsPaths), len(hashes), "Error listing epubs: paths and hashes should have same length.")
	assert.Equal(len(epubsPaths), len(epubs), "Error listing epubs: expected 2 epubs, got %d.", len(epubsPaths))
	for i, path := range epubsPaths {
		relativePath, err := filepath.Rel(testDir, path)
		assert.Nil(err)
		assert.Equal(epubs[i].filename, relativePath, "Error getting path")
		assert.Equal(epubs[i].expectedSha256, hashes[i], "Error getting hash")
	}
}

func TestHelpersIsDirEmpty(t *testing.T) {
	fmt.Println("+ Testing Helpers/Filesystem...")
	assert := assert.New(t)
	// testing on non-empty dir
	currentdir, err := os.Getwd()
	require.Nil(t, err, "Error getting current directory")
	isEmpty, err := IsDirectoryEmpty(currentdir)
	assert.Nil(err, "Error opening current directory")
	assert.False(isEmpty, "Current directory is not empty")
	exists := DirectoryExists(currentdir)
	assert.True(exists, "Current directory exists")

	// testing on non existing dir
	nonExistingDir := filepath.Join(currentdir, "doesnotexist")
	isEmpty, err = IsDirectoryEmpty(nonExistingDir)
	assert.NotNil(err, "Non existing directory should have triggered error")
	exists = DirectoryExists(nonExistingDir)
	assert.False(exists, "Directory does not exist")

	// testing on empty dir
	err = os.Mkdir(nonExistingDir, 0755)
	require.Nil(t, err, "Could not get create directory")
	isEmpty, err = IsDirectoryEmpty(nonExistingDir)
	assert.Nil(err, "Non existing directory should not have triggered error")
	assert.True(isEmpty, "Directory should be empty")
	exists = DirectoryExists(nonExistingDir)
	assert.True(exists, "Directory now exists")

	// cleanup
	err = os.Remove(nonExistingDir)
	require.Nil(t, err, "Could not remove directory")
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
		assert.Equal(t, CleanForPath(el.path), el.expectedCleanPath, "Error cleaning path")
	}
}

func TestHelpersCopy(t *testing.T) {
	fmt.Println("+ Testing Helpers/Copy()...")
	assert := assert.New(t)
	// getting test directory
	testDir, err := os.Getwd()
	require.Nil(t, err, "Error getting current directory")
	testDir = filepath.Join(filepath.Dir(testDir), "test")

	for _, el := range epubs {
		// copy file to _test
		origFilename := filepath.Join(testDir, el.filename)
		copyFilename := filepath.Join(testDir, el.filename+"_test")
		err := CopyFile(origFilename, copyFilename)
		require.Nil(t, err, "Could not copy file "+origFilename)
		// check copy exists
		absolutePath, err := FileExists(copyFilename)
		assert.Nil(err, "Copy file %s should exist", copyFilename)
		assert.Equal(absolutePath, copyFilename, "Getting copy path")
		// check copy hash
		copyHash, err := CalculateSHA256(copyFilename)
		assert.Nil(err, "Could not get hash from copy file %s", copyFilename)
		assert.Equal(copyHash, el.expectedSha256, "Copy hash %s different from source %s", copyHash, el.expectedSha256)
		// cleanup
		err = os.Remove(copyFilename)
		require.Nil(t, err, "Copy file %s could not be removed", copyFilename)
	}
}

func TestHelpersDeleteFolders(t *testing.T) {
	fmt.Println("+ Testing Helpers/DeleteEmptyFolders()...")
	/*// getting test directory
	testDir, err := os.Getwd()
	require.Nil(err, "Error getting current directory")
	testDir = filepath.Join(filepath.Dir(testDir), "test")*/
	// TODO
	// create folders
	// test
	// remove all
}
