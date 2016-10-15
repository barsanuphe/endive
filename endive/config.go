package endive

import (
	"os"
	"path/filepath"
	"strings"

	"io/ioutil"

	"gopkg.in/yaml.v2"
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
	// XdgArchiveDir is the path where database archives are kept
	XdgArchiveDir = Endive + "/archives/"
	// index path
	xdgIndexPath string = Endive + "/" + Endive + ".index"
)

// Constant Error values which can be compared to determine the type of error
const (
	ErrorConfigFileCreated Error = iota
	ErrorCannotLockDB
	ErrorLibraryRootDoesNotExist
	ErrorLibraryRootUnknown
	ErrorBadFormat
	WarningGoodReadsAPIKeyMissing
	WarningRetailSourceDoesNotExist
	WarningNonRetailSourceDoesNotExist
)

var errorMessages = map[Error]string{
	ErrorConfigFileCreated:             "Configuration file " + xdgConfigPath + " created. Populate it.",
	ErrorCannotLockDB:                  "Cannot lock library, is it running somewhere already?",
	ErrorLibraryRootDoesNotExist:       "Library root does not exist",
	ErrorLibraryRootUnknown:            "Missing library_root from config file",
	ErrorBadFormat:                     "Error parsing configuration file: bad format",
	WarningGoodReadsAPIKeyMissing:      "GoodReads API key not found! go to https://www.goodreads.com/api/keys to get one.",
	WarningRetailSourceDoesNotExist:    "At least one retail source does not exist.",
	WarningNonRetailSourceDoesNotExist: "At least one non-retail source does not exist.",
}

// Error handles errors found in configuration
type Error int

// Error implements the error interface
func (e Error) Error() string {
	return errorMessages[e]
}

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

// GetArchiveUniqueName in the endive archive directory.
func GetArchiveUniqueName(filename string) (archive string, err error) {
	return GetUniqueTimestampedFilename(filepath.Join(xdg.Data.Dirs()[0], XdgArchiveDir), filename)
}

// GetConfigPath gets the default path for configuration.
func GetConfigPath() (configFile string, err error) {
	configFile, err = xdg.Config.Find(xdgConfigPath)
	if err != nil {
		configFile, err = xdg.Config.Ensure(xdgConfigPath)
		if err != nil {
			return
		}
		err = ErrorConfigFileCreated
	}
	return
}

// GetIndexPath gets the default index path
func GetIndexPath() (path string) {
	path, err := xdg.Cache.Find(xdgIndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			path = filepath.Join(xdg.Cache.Dirs()[0], xdgIndexPath)
		} else {
			panic(err)
		}
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
		err = ErrorCannotLockDB
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

func interfaceToStringSlice(in interface{}) ([]string, error) {
	out := []string{}
	switch in.(type) {
	case []interface{}:
		for _, e := range in.([]interface{}) {
			out = append(out, e.(string))
		}
	default:
		return out, ErrorBadFormat
	}
	return out, nil
}

func interfaceToMap(in interface{}) (map[string][]string, error) {
	out := make(map[string][]string)
	switch in.(type) {
	case map[interface{}]interface{}:
		for k, v := range in.(map[interface{}]interface{}) {
			slice, err := interfaceToStringSlice(v)
			if err != nil {
				return out, err
			}
			out[k.(string)] = slice
		}
	default:
		return out, ErrorBadFormat
	}
	return out, nil
}

// Load configuration file using viper.
func (c *Config) Load() (err error) {
	conf := make(map[interface{}]interface{})
	contents, err := ioutil.ReadFile(c.Filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(contents, &conf)
	if err != nil {
		return err
	}

	if val, ok := conf["library_root"]; ok {
		c.LibraryRoot = val.(string)
	} else {
		return ErrorLibraryRootUnknown
	}
	if val, ok := conf["database_filename"]; ok {
		c.DatabaseFile = filepath.Join(c.LibraryRoot, val.(string))
	} else {
		c.DatabaseFile = filepath.Join(c.LibraryRoot, databaseFilename)
	}
	if val, ok := conf["epub_filename_format"]; ok {
		c.EpubFilenameFormat = val.(string)
	} else {
		c.EpubFilenameFormat = "$a [$y] $t"
	}
	if val, ok := conf["ereader_root"]; ok {
		c.EReaderMountPoint = val.(string)
	}
	if val, ok := conf["goodreads_api_key"]; ok {
		c.GoodReadsAPIKey = val.(string)
	} else {
		c.GoodReadsAPIKey = os.Getenv("GR_API_KEY")
		if c.GoodReadsAPIKey == "" {
			return WarningGoodReadsAPIKeyMissing
		}
	}
	if val, ok := conf["retail_source"]; ok {
		c.RetailSource, err = interfaceToStringSlice(val)
		if err != nil {
			return err
		}
	}
	if val, ok := conf["nonretail_source"]; ok {
		c.NonRetailSource, err = interfaceToStringSlice(val)
		if err != nil {
			return err
		}
	}
	if val, ok := conf["author_aliases"]; ok {
		c.AuthorAliases, err = interfaceToMap(val)
		if err != nil {
			return err
		}
	}
	if val, ok := conf["tag_aliases"]; ok {
		c.TagAliases, err = interfaceToMap(val)
		if err != nil {
			return err
		}
	}
	if val, ok := conf["publisher_aliases"]; ok {
		c.PublisherAliases, err = interfaceToMap(val)
		if err != nil {
			return err
		}
	}
	return
}

// Check if the paths in the configuration file are valid, and if the EpubFilename Format is ok.
func (c *Config) Check() error {
	if !DirectoryExists(c.LibraryRoot) {
		return ErrorLibraryRootDoesNotExist
	}
	// checking for sources, warnings only.
	for _, source := range c.RetailSource {
		if !DirectoryExists(source) {
			return WarningRetailSourceDoesNotExist
		}
	}
	for _, source := range c.NonRetailSource {
		if !DirectoryExists(source) {
			return WarningNonRetailSourceDoesNotExist
		}
	}
	return nil
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
	return TabulateRows(rows, "Config", "Value")
}
