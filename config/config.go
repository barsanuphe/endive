/*
Package config is a subpackage of Endive.

It aims at reading and checking the Endive configuration file.
It also deals with the internal database of already imported files (tracked through their SHA256 hashes).
*/
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	h "github.com/barsanuphe/endive/helpers"

	"github.com/spf13/viper"
	"launchpad.net/go-xdg"
)

const (
	// Endive is the name of this program.
	Endive           = "endive"
	xdgConfigPath    = Endive + "/" + Endive + ".yaml"
	databaseFilename = Endive + ".json"
	// XdgLogPath is the path for the main log file.
	XdgLogPath = Endive + "/" + Endive + ".log"
	// XdgLockPath is the path for the db lock.
	XdgLockPath = Endive + "/" + Endive + ".lock"
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
	TagAliases         map[string][]string
	PublisherAliases   map[string][]string
	EReaderMountPoint  string
	GoodReadsAPIKey    string
}

// GetConfigPath gets the default path for configuration.
func GetConfigPath() (configFile string, err error) {
	configFile, err = xdg.Config.Find(xdgConfigPath)
	if err != nil {
		configFile, err = xdg.Config.Ensure(xdgConfigPath)
		if err != nil {
			return
		}
		h.Infof("Configuration file %s created. Populate it.", xdgConfigPath)
	}
	return
}

// SetLock sets the library lock.
func SetLock() (err error) {
	_, err = xdg.Data.Find(XdgLockPath)
	if err != nil {
		_, err = xdg.Data.Ensure(XdgLockPath)
		if err != nil {
			return
		}
	} else {
		err = errors.New("Cannot lock library, is it running somewhere already?")
	}
	return
}

// RemoveLock for current library.
func RemoveLock() (err error) {
	lockFile, err := xdg.Data.Find(XdgLockPath)
	if err != nil {
		return
	}
	return os.Remove(lockFile)
}

// Load configuration file using viper.
func (c *Config) Load() (err error) {
	h.Debugf("Loading Config %s...\n", c.Filename)
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
	c.TagAliases = conf.GetStringMapStringSlice("tag_aliases")
	c.PublisherAliases = conf.GetStringMapStringSlice("publisher_aliases")
	c.EpubFilenameFormat = conf.GetString("epub_filename_format")
	if c.EpubFilenameFormat == "" {
		c.EpubFilenameFormat = "$a [$y] $t"
	}
	c.EReaderMountPoint = conf.GetString("ereader_root")
	c.GoodReadsAPIKey = conf.GetString("goodreads_api_key")
	if c.GoodReadsAPIKey == "" {
		c.GoodReadsAPIKey = os.Getenv("GR_API_KEY")
		if c.GoodReadsAPIKey == "" {
			h.Warning("Warning: no GoodReads API key found! go to https://www.goodreads.com/api/keys to get one.")
		}
	}

	return
}

// Check if the paths in the configuration file are valid, and if the EpubFilename Format is ok.
func (c *Config) Check() (err error) {
	h.Debug("Checking Config...")
	if !h.DirectoryExists(c.LibraryRoot) {
		err = errors.New("Library root " + c.LibraryRoot + " does not exist")
		h.Error(err.Error())
		return err
	}
	// checking for sources, warnings only.
	for _, source := range c.RetailSource {
		if !h.DirectoryExists(source) {
			h.Warning("Warning: retail source " + source + " does not exist.")
		}
	}
	for _, source := range c.NonRetailSource {
		if !h.DirectoryExists(source) {
			h.Warning("Warning: non-retail source " + source + " does not exist.")
		}
	}
	return
}

// ListAuthorAliases from the configuration file.
func (c *Config) ListAuthorAliases() (allAliases string) {
	for mainalias, aliases := range c.AuthorAliases {
		allAliases += mainalias + " => " + strings.Join(aliases, ", ") + "\n"
	}
	return
}

// String displays all configuration information.
func (c *Config) String() string {
	fmt.Println("Printing Config contents...")
	var rows [][]string
	rows = append(rows, []string{"Library directory", c.LibraryRoot})
	rows = append(rows, []string{"Database file", c.DatabaseFile})
	rows = append(rows, []string{"Epub filename format", c.EpubFilenameFormat})
	if c.GoodReadsAPIKey != "" {
		rows = append(rows, []string{"Goodreads API Key", "present"})
	}
	rows = append(rows, []string{"E-Reader mount point", c.EReaderMountPoint})
	rows = append(rows, []string{"Retail sources", strings.Join(c.RetailSource, ", ")})
	rows = append(rows, []string{"Non-Retail sources", strings.Join(c.NonRetailSource, ", ")})
	for mainalias, aliases := range c.AuthorAliases {
		rows = append(rows, []string{"Author alias: " + mainalias, strings.Join(aliases, ", ")})
	}
	for mainalias, aliases := range c.TagAliases {
		rows = append(rows, []string{"Tag alias: " + mainalias, strings.Join(aliases, ", ")})
	}
	for mainalias, aliases := range c.PublisherAliases {
		rows = append(rows, []string{"Publisher alias: " + mainalias, strings.Join(aliases, ", ")})
	}
	return h.TabulateRows(rows, "Config", "Value")
}
