package endive

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var configFile = "../test/config.yaml"

func TestConfigLoad(t *testing.T) {
	fmt.Println("+ Testing Config.Load()...")
	assert := assert.New(t)
	c := Config{Filename: configFile}

	err := c.Load()
	assert.Nil(err, "Error loading configuration file")
	assert.Equal("test_library", c.LibraryRoot,  "Error loading library root")
	assert.Equal("test_library/endive.json", c.DatabaseFile, "Error loading database filename")
	assert.Equal(2, len(c.RetailSource), "Error: loading retail sources, expected 2")
	assert.Equal(1, len(c.NonRetailSource), "Error: loading retail sources, expected 1")
	assert.Equal(2, len(c.AuthorAliases), "Error: loading author aliases, expected 2")
	assert.Equal(2, len(c.AuthorAliases["China Miéville"]), "Error: loading author aliases for china miéville, should have gotten 2")
	// checking library root, expecting error
	err = c.Check()
	assert.NotNil(err, "Error checking configuration file, library root should not exist.")
	// library root creation
	err = os.Mkdir(c.LibraryRoot, 0777)
	assert.Nil(err, "Error creating library root")
	// check should be ok
	err = c.Check()
	assert.NotEqual(ErrorLibraryRootDoesNotExist, err, "Library root should exist now")
	// cleanup
	err = os.Remove(c.LibraryRoot)
	assert.Nil(err, "Error removing library root")
}
