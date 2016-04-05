package main

import (
	"fmt"
	"os"
	"testing"
)

var configFile string = "test/config.yaml"

func TestConfigLoad(t *testing.T) {
	fmt.Println("+ Testing Config.Load()...")
	c := Config{Filename: configFile}

	err := c.Load()
	if err != nil {
		t.Errorf("Error loading configuration file: %s", err.Error())
	}
	if c.LibraryRoot != "test_library" {
		t.Errorf("Error loading library root: %s instead of %s", c.LibraryRoot, "test_library")
	}
	if c.DatabaseFile != "test_library/endive.json" {
		t.Errorf("Error loading database filename: %s instead of %s", c.DatabaseFile, "test_library/endive.json")
	}
	if len(c.RetailSource) != 2 {
		t.Errorf("Error: loading retail sources, expected 2 instead of %d", len(c.RetailSource))
	}
	if len(c.NonRetailSource) != 1 {
		t.Errorf("Error: loading retail sources, expected 1 instead of %d", len(c.NonRetailSource))
	}
	if len(c.AuthorAliases) != 2 {
		t.Errorf("Error: loading author aliases, expected 2 instead of %d", len(c.AuthorAliases))
	}
	if len(c.AuthorAliases["China Miéville"]) != 2 {
		t.Errorf("Error: loading author aliases for china miéville, should have gotten 2 instead of %d", len(c.AuthorAliases["China Miéville"]))
	}
	// checking library root, expecting error
	err = c.Check()
	if err == nil {
		t.Errorf("Error checking configuration file, library root should not exist.")
	}
	// library root creation
	err = os.Mkdir(c.LibraryRoot, 0777)
	if err != nil {
		t.Errorf("Error creating library root")
	}
	// check should be ok
	err = c.Check()
	if err != nil {
		t.Errorf("Error checking configuration file: %s", err.Error())
	}
	// cleanup
	err = os.Remove(c.LibraryRoot)
	if err != nil {
		t.Errorf("Error removing library root")
	}
}
