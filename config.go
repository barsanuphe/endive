package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"launchpad.net/go-xdg"
)

const (
	endive        = "endive"
	xdgConfigPath = endive + "/" + endive + ".yaml"
)

var databaseFilename string = endive + ".json"

// "launchpad.net/go-xdg"

// Config holds all relevant information
type Config struct {
	Filename           string
	DatabaseFile       string
	LibraryRoot        string
	RetailSource       []string
	NonRetailSource    []string
	EpubFilenameFormat string
	AuthorAliases      map[string][]string
	EReaderTarget      string
}

// getConfigPath gets the default path for configuration.
func getConfigPath() (configFile string, err error) {
	configFile, err = xdg.Config.Find(xdgConfigPath)
	if err != nil {
		configFile, err = xdg.Config.Ensure(configFile)
		if err != nil {
			return
		}
		fmt.Println("Configuration file", configFile, "created. Populate it.")
	}
	return
}

// Load configuration file using viper.
func (c *Config) Load() (err error) {
	fmt.Println("Loading Config...")
	// TODO find in xdg folder
	conf := viper.New()
	conf.SetConfigName(filepath.Base(c.Filename))
	conf.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Dir(c.Filename))

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	c.LibraryRoot = viper.GetString("library_root")
	c.DatabaseFile = filepath.Join(c.LibraryRoot, databaseFilename)
	c.RetailSource = viper.GetStringSlice("retail_source")
	c.NonRetailSource = viper.GetStringSlice("nonretail_source")
	c.AuthorAliases = viper.GetStringMapStringSlice("author_aliases")
	c.EpubFilenameFormat = viper.GetString("epub_filename_format")
	c.EReaderTarget = viper.GetString("ereader_target")
	return
}

// Check if the paths in the configuration file are valid, and if the EpubFilename Format is ok.
func (c *Config) Check() (err error) {
	fmt.Println("Checking Config...")
	if !directoryExists(c.LibraryRoot) {
		return errors.New("Library root " + c.LibraryRoot + " does not exist")
	}
	// checking for sources, warnings only.
	for _, source := range c.RetailSource {
		if !directoryExists(source) {
			fmt.Println("Warning: retail source " + source + " does not exist.")
		}
	}
	for _, source := range c.NonRetailSource {
		if !directoryExists(source) {
			fmt.Println("Warning: non-retail source " + source + " does not exist.")
		}
	}
	return
}

// ListAuthorAliases from the configuration file.
func (c *Config) ListAuthorAliases() (aliases string, err error) {
	fmt.Println("Listing Author aliases...")
	return
}

// String displays all configuration information.
func (c *Config) String() (err error) {
	fmt.Println("Printing Config contents...")
	return
}
