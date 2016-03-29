package main

import "fmt"

// use https://github.com/spf13/viper to parse config
// "launchpad.net/go-xdg"

// Config holds all relevant information
type Config struct {
	Filename           string
	DatabaseFile       string
	LibraryRoot        string
	RetailSource       string
	NonRetailSource    string
	EpubFilenameFormat string
	AuthorAliases      map[string]string
	EReaderTarget      string
}

// Parse configuration file using viper.
func (c *Config) Load() (err error) {
	fmt.Println("Loading Config...")
	return
}

// Check if the paths in the configuration file are valid, and if the EpubFilename Format is ok.
func (c *Config) Check() (err error) {
	fmt.Println("Checking Config...")
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
