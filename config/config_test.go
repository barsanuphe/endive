package config

import (
	"fmt"
	"os"
	"testing"

	h "github.com/barsanuphe/endive/helpers"
	"github.com/stretchr/testify/assert"
)

var configFile = "../test/config.yaml"

func TestMain(m *testing.M) {
	// init logger
	err := h.GetLogger("log_testing")
	if err != nil {
		panic(err)
	}
	// do the actual testing
	retCode := m.Run()
	// cleanup
	h.LogFile.Close()
	if err := os.Remove("log_testing"); err != nil {
		panic(err)
	}
	os.Exit(retCode)
}

func TestConfigLoad(t *testing.T) {
	fmt.Println("+ Testing Config.Load()...")
	assert := assert.New(t)
	c := Config{Filename: configFile}

	err := c.Load()
	assert.Nil(err, "Error loading configuration file")
	assert.Equal(c.LibraryRoot, "test_library", "Error loading library root")
	assert.Equal(c.DatabaseFile, "test_library/endive.json", "Error loading database filename")
	assert.Equal(len(c.RetailSource), 2, "Error: loading retail sources, expected 2")
	assert.Equal(len(c.NonRetailSource), 1, "Error: loading retail sources, expected 1")
	assert.Equal(len(c.AuthorAliases), 2, "Error: loading author aliases, expected 2")
	assert.Equal(len(c.AuthorAliases["China Miéville"]), 2, "Error: loading author aliases for china miéville, should have gotten 2")
	// checking library root, expecting error
	err = c.Check()
	assert.NotNil(err, "Error checking configuration file, library root should not exist.")
	// library root creation
	err = os.Mkdir(c.LibraryRoot, 0777)
	assert.Nil(err, "Error creating library root")
	// check should be ok
	err = c.Check()
	assert.Nil(err, "Error checking configuration file")
	// cleanup
	err = os.Remove(c.LibraryRoot)
	assert.Nil(err, "Error removing library root")
}
