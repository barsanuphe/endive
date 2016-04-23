package book

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	c "github.com/barsanuphe/endive/config"
	"github.com/barsanuphe/endive/helpers"
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
		`{"id":0,"retail":{"filename":"test/pg16328.epub","hash":"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03","replace":"false"},"nonretail":{"filename":"","hash":"","replace":""},"epub_metadata":{"title":"Beowulf / An Anglo-Saxon Epic Poem","original_title":"","image_url":"","num_pages":"","authors":null,"isbn":"","year":"2005","description":"","series":null,"average_rating":"","tags":[{"name":"Epic poetry, English (Old)"},{"name":"Monsters -- Poetry"},{"name":"Dragons -- Poetry"}]},"metadata":{"title":"Beowulf / An Anglo-Saxon Epic Poem","original_title":"","image_url":"","num_pages":"","authors":null,"isbn":"","year":"2005","description":"","series":null,"average_rating":"","tags":[{"name":"Epic poetry, English (Old)"},{"name":"Monsters -- Poetry"},{"name":"Dragons -- Poetry"}]},"progress":"unread","readdate":"","rating":"","review":""}`,
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
		`{"id":1,"retail":{"filename":"test/pg17989.epub","hash":"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a","replace":"false"},"nonretail":{"filename":"","hash":"","replace":""},"epub_metadata":{"title":"Le comte de Monte-Cristo, Tome I","original_title":"","image_url":"","num_pages":"","authors":["Alexandre Dumas"],"isbn":"","year":"2006","description":"","series":null,"average_rating":"","tags":[{"name":"Historical fiction"},{"name":"Revenge -- Fiction"},{"name":"Adventure stories"},{"name":"Prisoners -- Fiction"},{"name":"France -- History -- 19th century -- Fiction"},{"name":"Pirates -- Fiction"},{"name":"Dantès, Edmond (Fictitious character) -- Fiction"}]},"metadata":{"title":"Le comte de Monte-Cristo, Tome I","original_title":"","image_url":"","num_pages":"","authors":["Alexandre Dumas"],"isbn":"","year":"2006","description":"","series":null,"average_rating":"","tags":[{"name":"Historical fiction"},{"name":"Revenge -- Fiction"},{"name":"Adventure stories"},{"name":"Prisoners -- Fiction"},{"name":"France -- History -- 19th century -- Fiction"},{"name":"Pirates -- Fiction"},{"name":"Dantès, Edmond (Fictitious character) -- Fiction"}]},"progress":"unread","readdate":"","rating":"","review":""}`,
		"Alexandre Dumas 2006 Le comte de Monte-Cristo, Tome I.epub",
		"Alexandre Dumas 2006 Le comte de Monte-Cristo, Tome I [retail].epub",
		"fr/Alexandre Dumas/2006. [Alexandre Dumas] (Le comte de Monte-Cristo, Tome I).epub",
	},
}

var parentDir string
var standardTestConfig c.Config
var isRetail = true

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parentDir = filepath.Dir(wd)
	standardTestConfig = c.Config{LibraryRoot: parentDir}
}

// TestBookJSON tests both JSON() and FromJSON().
func TestBookJSON(t *testing.T) {
	fmt.Println("+ Testing Epub.JSON()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, isRetail)
		info, err := e.MainEpub().ReadMetadata()
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.FullPath(), err)
		}
		e.EpubMetadata = info
		e.Metadata = info
		err = e.RetailEpub.GetHash()
		if err != nil {
			t.Errorf("Error getting Hash for epub %s", e.FullPath())
		}
		jsonString, err := e.JSON()
		if err != nil {
			t.Errorf("Error exporting epub %s to JSON string", e.FullPath())
		}
		if jsonString != testEpub.expectedJSONString {
			t.Errorf("JSON(%s) returned:\n%s\nexpected:\n%s!", testEpub.filename, jsonString, testEpub.expectedJSONString)
		}
		// recreating new Epub object from Json string
		f := Book{}
		f.FromJSON([]byte(jsonString))
		// comparing a few fields
		if e.Metadata.Title() != f.Metadata.Title() {
			t.Errorf("Error rebuilt Epub and original are different")
		}
		// exporting again to compare
		jsonString2, err := f.JSON()
		if err != nil {
			t.Errorf("Error exporting rebuilt Epub to JSON string")
		}
		if jsonString != jsonString2 {
			t.Errorf("Error rebuilt Epub and original are different")
		}
	}
}

func TestBookNewName(t *testing.T) {
	fmt.Println("+ Testing Epub.generateNewName()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, !isRetail)
		info, err := e.MainEpub().ReadMetadata()
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.FullPath(), err)
		}
		e.EpubMetadata = info
		e.Metadata = info

		newName1, err := e.generateNewName("$a $y $t", !isRetail)
		if err != nil {
			t.Errorf("Error generating new name")
		}
		if newName1 != testEpub.expectedFormat1 {
			t.Errorf("Error getting new name, expected %s, got %s", testEpub.expectedFormat1, newName1)
		}
		newName2, err := e.generateNewName("$l/$a/$y. [$a] ($t)", !isRetail)
		if err != nil {
			t.Errorf("Error generating new name")
		}
		if newName2 != testEpub.expectedFormat2 {
			t.Errorf("Error getting new name, expected %s, got %s", testEpub.expectedFormat2, newName2)
		}

		e = NewBook(10+i, testEpub.filename, standardTestConfig, isRetail)
		info, err = e.MainEpub().ReadMetadata()
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.FullPath(), err)
		}
		e.EpubMetadata = info
		e.Metadata = info

		newName1, err = e.generateNewName("$a $y $t", isRetail)
		if err != nil {
			t.Errorf("Error generating new name")
		}
		if newName1 != testEpub.expectedFormat1Retail {
			t.Errorf("Error getting new name, expected %s, got %s", testEpub.expectedFormat1Retail, newName1)
		}
	}
}

func TestBookRefresh(t *testing.T) {
	fmt.Println("+ Testing Epub.Refresh()...")
	cfg := c.Config{EpubFilenameFormat: "$a $y $t", LibraryRoot: parentDir}
	for i, testEpub := range epubs {
		// copy testEpub.filename
		epubFilename := filepath.Base(testEpub.filename)
		epubDir := filepath.Dir(testEpub.filename)
		tempCopy := filepath.Join(parentDir, epubDir, "temp_"+epubFilename)

		err := helpers.CopyFile(filepath.Join(parentDir, testEpub.filename), tempCopy)
		if err != nil {
			t.Errorf("Error copying %s to %s", testEpub.filename, tempCopy)
		}

		// creating Epub object
		e := NewBook(i, tempCopy, cfg, isRetail)
		info, err := e.MainEpub().ReadMetadata()
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.FullPath(), err)
		}
		e.EpubMetadata = info
		e.Metadata = info

		// refresh
		wasRenamed, newName, err := e.Refresh()
		if err != nil {
			t.Errorf("Error generating new name: " + err.Error())
		}
		if !wasRenamed[0] {
			t.Errorf("Error renaming %s", tempCopy)
		}
		if wasRenamed[1] {
			t.Errorf("Error: should not have rename non-existant non-retail epub.")
		}
		if newName[0] != testEpub.expectedFormat1Retail {
			t.Errorf("Error renaming %s, got %s, expected %s", tempCopy, newName[0], testEpub.expectedFormat1Retail)
		}

		// getting epub path relative to parent dir (ie simulated library root) for comparison
		filename, err := filepath.Rel(parentDir, e.FullPath())
		if err != nil {
			t.Errorf("Error getting relative path: " + err.Error())
		}
		if newName[0] != filename {
			t.Errorf("Error setting new name %s, got %s, expected %s", tempCopy, newName[0], filename)
		}

		//  cleanup
		if err != nil || !wasRenamed[0] {
			err = os.Remove(tempCopy)
			if err != nil {
				t.Errorf("Error removing temp copy %s", tempCopy)
			}
		} else {
			err = os.Remove(filepath.Join(parentDir, newName[0]))
			if err != nil {
				t.Errorf("Error removing temp copy %s", newName[0])
			}
		}
	}
}

// TestBookSetReadDate tests for SetReadDate and SetReadDateToday
func TestBookSetReadDate(t *testing.T) {
	fmt.Println("+ Testing Epub.SetReadDate()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, isRetail)

		err := e.SetReadDateToday()
		if err != nil {
			t.Errorf("Error setting read date")
		}

		currentDate := time.Now().Local().Format("2006-01-02")
		if e.ReadDate != currentDate {
			t.Errorf("Error setting read date, expected %s, got %s", currentDate, e.ReadDate)
		}
	}
}

// TestBookProgress tests for SetProgress
func TestBookProgress(t *testing.T) {
	fmt.Println("+ Testing Epub.TestEpubProgress()...")
	e := NewBook(0, epubs[0].filename, standardTestConfig, isRetail)

	err := e.SetProgress("Shortlisted")
	if err != nil {
		t.Errorf("Error setting progress Shortlisted")
	}
	if e.Progress != "shortlisted" {
		t.Errorf("Error setting progress, expected %s, got %s", "shortlisted", e.Progress)
	}

	err = e.SetProgress("mhiuh")
	if err == nil {
		t.Errorf("Error setting progress should have failed")
	}
	if e.Progress != "shortlisted" {
		t.Errorf("Error setting progress, expected %s, got %s", "shortlisted", e.Progress)
	}
}
