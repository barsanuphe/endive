package book

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var epubs = []struct {
	filename                string
	expectedTitle           string
	expectedAuthor          string
	expectedPublicationYear string
	expectedLanguage        string
	expectedSha256          string
	expectedJSONString      string
	expectedFormat1         string
	expectedFormat1Retail   string
	expectedFormat2         string
}{
	{
		"test/pg16328.epub",
		"Beowulf / An Anglo-Saxon Epic Poem",
		"Unknown",
		"2005",
		"en",
		"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03",
		`{"id":0,"retail":{"filename":"test/pg16328.epub","hash":"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03","replace":"false"},"nonretail":{"filename":"","hash":"","replace":""},"epub_metadata":{"title":"Beowulf / An Anglo-Saxon Epic Poem","original_title":"","image_url":"","num_pages":"","authors":null,"isbn":"","year":"2005","edition_year":"2005","description":"","series":null,"average_rating":"","tags":[{"name":"dragons -- poetry"}],"category":"Unknown","main_genre":"monsters -- poetry","language":"en","publisher":""},"metadata":{"title":"Beowulf / An Anglo-Saxon Epic Poem","original_title":"","image_url":"","num_pages":"","authors":null,"isbn":"","year":"2005","edition_year":"2005","description":"","series":null,"average_rating":"","tags":[{"name":"dragons -- poetry"}],"category":"Unknown","main_genre":"monsters -- poetry","language":"en","publisher":""},"progress":"unread","readdate":"","rating":"","review":""}`,
		"Unknown 2005 Beowulf - An Anglo-Saxon Epic Poem.epub",
		"Unknown 2005 Beowulf - An Anglo-Saxon Epic Poem [retail].epub",
		"en/Unknown/2005. [Unknown] (Beowulf - An Anglo-Saxon Epic Poem).epub",
	},
	{
		"test/pg17989.epub",
		"Le comte de Monte-Cristo, Tome I",
		"Alexandre Dumas",
		"2006",
		"fr",
		"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a",
		`{"id":1,"retail":{"filename":"test/pg17989.epub","hash":"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a","replace":"false"},"nonretail":{"filename":"","hash":"","replace":""},"epub_metadata":{"title":"Le comte de Monte-Cristo, Tome I","original_title":"","image_url":"","num_pages":"","authors":["Alexandre Dumas"],"isbn":"","year":"2006","edition_year":"2006","description":"","series":null,"average_rating":"","tags":[{"name":"revenge -- fiction"},{"name":"adventure stories"},{"name":"prisoners -- fiction"},{"name":"france -- history -- 19th century -- fiction"},{"name":"pirates -- fiction"},{"name":"dantès, edmond (fictitious character) -- fiction"}],"category":"Unknown","main_genre":"historical fiction","language":"fr","publisher":""},"metadata":{"title":"Le comte de Monte-Cristo, Tome I","original_title":"","image_url":"","num_pages":"","authors":["Alexandre Dumas"],"isbn":"","year":"2006","edition_year":"2006","description":"","series":null,"average_rating":"","tags":[{"name":"revenge -- fiction"},{"name":"adventure stories"},{"name":"prisoners -- fiction"},{"name":"france -- history -- 19th century -- fiction"},{"name":"pirates -- fiction"},{"name":"dantès, edmond (fictitious character) -- fiction"}],"category":"Unknown","main_genre":"historical fiction","language":"fr","publisher":""},"progress":"unread","readdate":"","rating":"","review":""}`,
		"Alexandre Dumas 2006 Le comte de Monte-Cristo, Tome I.epub",
		"Alexandre Dumas 2006 Le comte de Monte-Cristo, Tome I [retail].epub",
		"fr/Alexandre Dumas/2006. [Alexandre Dumas] (Le comte de Monte-Cristo, Tome I).epub",
	},
}

var parentDir string
var standardTestConfig cfg.Config
var isRetail = true

func TestMain(m *testing.M) {
	// init global variables
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parentDir = filepath.Dir(wd)
	tags := make(map[string][]string)
	tags["science-fiction"] = []string{"sci-fi"}
	standardTestConfig = cfg.Config{LibraryRoot: parentDir, TagAliases: tags}
	// init logger
	err = h.GetLogger("log_testing")
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

// TestBookJSON tests both JSON() and FromJSON().
func TestBookJSON(t *testing.T) {
	fmt.Println("+ Testing Book.JSON()...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, isRetail)
		info, err := e.MainEpub().ReadMetadata()
		assert.NotNil(err, "Error should be found (no ISBN in test epubs) for "+e.FullPath())
		if err != nil {
			assert.Equal("ISBN not found in epub", err.Error(), "Error should only mention missing ISBN")
		}
		e.EpubMetadata = info
		e.Metadata = info

		err = e.RetailEpub.GetHash()
		assert.Nil(err, "Error getting hash for "+e.FullPath())

		jsonString, err := e.JSON()
		assert.Nil(err, "Error exporting epub to JSON string: "+e.FullPath())
		assert.Equal(testEpub.expectedJSONString, jsonString, "JSON strings are different")

		// recreating new Epub object from Json string
		f := Book{}
		f.FromJSON([]byte(jsonString))
		// comparing a few fields
		assert.Equal(e.Metadata.Title(), f.Metadata.Title(), "Error rebuilt Epub and original are different")

		// exporting again to compare
		jsonString2, err := f.JSON()
		assert.Nil(err, "Error exporting rebuilt Epub to JSON string")
		assert.Equal(jsonString, jsonString2, "JSON strings are different")
	}
}

func TestBookNewName(t *testing.T) {
	fmt.Println("+ Testing Book.generateNewName()...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, !isRetail)
		info, err := e.MainEpub().ReadMetadata()
		assert.NotNil(err, "Error should be found (no ISBN in test epubs) for "+e.FullPath())
		if err != nil {
			assert.Equal("ISBN not found in epub", err.Error(), "Error should only mention missing ISBN")
		}
		e.EpubMetadata = info
		e.Metadata = info

		newName1, err := e.generateNewName("$a $y $t", !isRetail)
		assert.Nil(err, "Error generating new name")
		assert.Equal(newName1, testEpub.expectedFormat1, "Error getting new name")

		newName2, err := e.generateNewName("$l/$a/$y. [$a] ($t)", !isRetail)
		assert.Nil(err, "Error generating new name")
		assert.Equal(newName2, testEpub.expectedFormat2, "Error getting new name")

		e = NewBook(10+i, testEpub.filename, standardTestConfig, isRetail)
		info, err = e.MainEpub().ReadMetadata()
		assert.NotNil(err, "Error should be found (no ISBN in test epubs) for "+e.FullPath())
		if err != nil {
			assert.Equal("ISBN not found in epub", err.Error(), "Error should only mention missing ISBN")
		}
		e.EpubMetadata = info
		e.Metadata = info

		newName1, err = e.generateNewName("$a $y $t", isRetail)
		assert.Nil(err, "Error generating new name")
		assert.Equal(newName1, testEpub.expectedFormat1Retail, "Error getting new name")
	}
}

func TestBookRefresh(t *testing.T) {
	fmt.Println("+ Testing Book.Refresh()...")
	cfg := cfg.Config{EpubFilenameFormat: "$a $y $t", LibraryRoot: parentDir}
	assert := assert.New(t)
	for i, testEpub := range epubs {
		// copy testEpub.filename
		epubFilename := filepath.Base(testEpub.filename)
		epubDir := filepath.Dir(testEpub.filename)
		tempCopy := filepath.Join(parentDir, epubDir, "temp_"+epubFilename)

		err := h.CopyFile(filepath.Join(parentDir, testEpub.filename), tempCopy)
		assert.Nil(err, "Error copying")

		// creating Epub object
		e := NewBook(i, tempCopy, cfg, isRetail)
		info, err := e.MainEpub().ReadMetadata()
		assert.NotNil(err, "Error should be found (no ISBN in test epubs) for "+e.FullPath())
		if err != nil {
			assert.Equal("ISBN not found in epub", err.Error(), "Error should only mention missing ISBN")
		}
		e.EpubMetadata = info
		e.Metadata = info

		// refresh
		wasRenamed, newName, err := e.Refresh()
		assert.Nil(err, "Error generating new name")
		assert.True(wasRenamed[0], "Error renaming "+tempCopy)
		assert.False(wasRenamed[1], "Error: should not have rename non-existant non-retail epub.")
		assert.Equal(newName[0], testEpub.expectedFormat1Retail, "Error renaming %s "+tempCopy)

		// getting epub path relative to parent dir (ie simulated library root) for comparison
		filename, err := filepath.Rel(parentDir, e.FullPath())
		assert.Nil(err, "Error getting relative path")
		assert.Equal(newName[0], filename, "Error setting new name")

		//  cleanup
		if err != nil || !wasRenamed[0] {
			err = os.Remove(tempCopy)
			assert.Nil(err, "Error removing temp copy "+tempCopy)
		} else {
			err = os.Remove(filepath.Join(parentDir, newName[0]))
			assert.Nil(err, "Error removing temp copy "+newName[0])
		}
	}
}

// TestBookSetReadDate tests for SetReadDate and SetReadDateToday
func TestBookSetReadDate(t *testing.T) {
	fmt.Println("+ Testing Book.SetReadDate()...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		currentDate := time.Now().Local().Format("2006-01-02")
		e := NewBook(i, testEpub.filename, standardTestConfig, isRetail)
		err := e.SetReadDateToday()
		assert.Nil(err, "Error setting read date")
		assert.Equal(e.ReadDate, currentDate, "Error setting read date")
	}
}

// TestBookProgress tests for SetProgress
func TestBookProgress(t *testing.T) {
	fmt.Println("+ Testing Book.TestEpubProgress()...")
	assert := assert.New(t)
	e := NewBook(0, epubs[0].filename, standardTestConfig, isRetail)

	err := e.SetProgress("Shortlisted")
	assert.Nil(err, "Error setting progress Shortlisted")
	assert.Equal(e.Progress, "shortlisted", "Error setting progress")

	err = e.SetProgress("mhiuh")
	assert.NotNil(err, "Error setting progress should have failed")
	assert.Equal(e.Progress, "shortlisted", "Error setting progress")
}

// TestBookSearchOnline tests for SearchOnline
func TestBookSearchOnline(t *testing.T) {
	fmt.Println("+ Testing Book.SearchOnline()...")
	assert := assert.New(t)
	// get GR api key
	key := os.Getenv("GR_API_KEY")
	require.NotEqual(t, len(key), 0, "Cannot get Goodreads API key")
	standardTestConfig.GoodReadsAPIKey = key

	for i, testEpub := range epubs {
		fmt.Println(testEpub.filename)
		e := NewBook(i, testEpub.filename, standardTestConfig, isRetail)
		info, err := e.MainEpub().ReadMetadata()
		assert.NotNil(err, "Error should be found (no ISBN in test epubs) for "+e.FullPath())
		if err != nil {
			assert.Equal("ISBN not found in epub", err.Error(), "Error should only mention missing ISBN")
		}
		e.EpubMetadata = info
		e.Metadata = info

		err = e.SearchOnline()
		assert.NotNil(err, "Expected error searching online, missing user input")
	}
}
