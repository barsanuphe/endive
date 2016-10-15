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
	assert.Equal("test_library", c.LibraryRoot, "Error loading library root")
	assert.Equal("test_library/endive.json", c.DatabaseFile, "Error loading database filename")
	assert.Equal(2, len(c.RetailSource), "Error: loading retail sources, expected 2")
	assert.Equal(1, len(c.NonRetailSource), "Error: loading retail sources, expected 1")
	assert.Equal(5, len(c.AuthorAliases), "Error: loading author aliases, expected 2")
	assert.Equal(1, len(c.AuthorAliases["Richard K. Morgan"]), "Error: loading author aliases for richard morgan, should have gotten 1")
	assert.Equal(1, len(c.TagAliases), "Error: loading tag aliases, expected 1")
	assert.Equal(3, len(c.TagAliases["science-fiction"]), "Expected 3 aliases for SF")
	assert.Equal(1, len(c.PublisherAliases), "Error: loading publisher aliases, expected 1")
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
