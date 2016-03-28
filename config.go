package main

// use https://github.com/spf13/viper to parse config

// Config holds all relevant information
type Config struct {
	Filename           string
	LibraryRoot        string
	RetailSource       string
	NonRetailSource    string
	EpubFilenameFormat string
	AuthorAliases      map[string]string
	EReaderTarget      string
}

// Parse configuration file using viper.
func (c *Config) Parse() (err error) {
	return
}

// Check if the paths in the configuration file are valid, and if the EpubFilename Format is ok.
func (c *Config) Check() (err error) {
	return
}

// ListAuthorAliases from the configuration file.
func (c *Config) ListAuthorAliases() (err error) {
	return
}

// String displays all configuration information.
func (c *Config) String() (err error) {
	return
}
