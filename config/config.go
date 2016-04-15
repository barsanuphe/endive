package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/barsanuphe/endive/helpers"
	"github.com/spf13/viper"
	"launchpad.net/go-xdg"
)

const (
	// Endive is the name of this program.
	Endive           = "endive"
	xdgConfigPath    = Endive + "/" + Endive + ".yaml"
	databaseFilename = Endive + ".json"
)

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

// GetConfigPath gets the default path for configuration.
func GetConfigPath() (configFile string, err error) {
	configFile, err = xdg.Config.Find(xdgConfigPath)
	if err != nil {
		configFile, err = xdg.Config.Ensure(xdgConfigPath)
		if err != nil {
			return
		}
		fmt.Println("Configuration file", xdgConfigPath, "created. Populate it.")
	}
	return
}

// Load configuration file using viper.
func (c *Config) Load() (err error) {
	fmt.Printf("Loading Config %s...\n", c.Filename)
	conf := viper.New()
	conf.SetConfigType("yaml")
	conf.SetConfigFile(c.Filename)

	err = conf.ReadInConfig()
	if err != nil {
		return
	}
	c.LibraryRoot = conf.GetString("library_root")
	db := conf.GetString("database_filename")
	if db == "" {
		c.DatabaseFile = filepath.Join(c.LibraryRoot, databaseFilename)
	} else {
		c.DatabaseFile = filepath.Join(c.LibraryRoot, db)
	}
	c.RetailSource = conf.GetStringSlice("retail_source")
	c.NonRetailSource = conf.GetStringSlice("nonretail_source")
	c.AuthorAliases = conf.GetStringMapStringSlice("author_aliases")
	c.EpubFilenameFormat = conf.GetString("epub_filename_format")
	if c.EpubFilenameFormat == "" {
		c.EpubFilenameFormat = "$a [$y] $t"
	}
	c.EReaderTarget = conf.GetString("ereader_target")
	return
}

// Check if the paths in the configuration file are valid, and if the EpubFilename Format is ok.
func (c *Config) Check() (err error) {
	fmt.Println("Checking Config...")
	if !helpers.DirectoryExists(c.LibraryRoot) {
		return errors.New("Library root " + c.LibraryRoot + " does not exist")
	}
	// checking for sources, warnings only.
	for _, source := range c.RetailSource {
		if !helpers.DirectoryExists(source) {
			fmt.Println("Warning: retail source " + source + " does not exist.")
		}
	}
	for _, source := range c.NonRetailSource {
		if !helpers.DirectoryExists(source) {
			fmt.Println("Warning: non-retail source " + source + " does not exist.")
		}
	}
	return
}

// ListAuthorAliases from the configuration file.
func (c *Config) ListAuthorAliases() (allAliases string) {
	fmt.Println("Listing Author aliases...")
	for mainalias, aliases := range c.AuthorAliases {
		allAliases += mainalias + " => " + strings.Join(aliases, ", ") + "\n"
	}
	return
}

// String displays all configuration information.
func (c *Config) String() (err error) {
	fmt.Println("Printing Config contents...")
	return
}
