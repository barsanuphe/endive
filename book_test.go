package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var epubs = []struct {
	filename                string
	expectedError           error
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
		errors.New("Metadata field creator does not exist"),
		"Beowulf / An Anglo-Saxon Epic Poem",
		"Unknown",
		"2005",
		"en",
		"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03",
		`{"filename":"test/pg16328.epub","hash":"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03","isretail":"false","progress":"unread","series":null,"author":"Unknown","title":"Beowulf / An Anglo-Saxon Epic Poem","language":"en","publicationyear":"2005","readdate":"","tags":null,"rating":"","review":"","description":"","replace":"false","isbn":"http://www.gutenberg.org/files/16328/16328-h/16328-h.htm"}`,
		"Unknown 2005 Beowulf - An Anglo-Saxon Epic Poem.epub",
		"Unknown 2005 Beowulf - An Anglo-Saxon Epic Poem [retail].epub",
		"en/Unknown/2005. [Unknown] (Beowulf - An Anglo-Saxon Epic Poem).epub",
	},
	{
		"test/pg17989.epub",
		nil,
		"Le comte de Monte-Cristo, Tome I",
		"Alexandre Dumas",
		"2006",
		"fr",
		"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a",
		`{"filename":"test/pg17989.epub","hash":"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a","isretail":"false","progress":"unread","series":null,"author":"Alexandre Dumas","title":"Le comte de Monte-Cristo, Tome I","language":"fr","publicationyear":"2006","readdate":"","tags":null,"rating":"","review":"","description":"","replace":"false","isbn":"http://www.gutenberg.org/files/17989/17989-h/17989-h.htm"}`,
		"Alexandre Dumas 2006 Le comte de Monte-Cristo, Tome I.epub",
		"Alexandre Dumas 2006 Le comte de Monte-Cristo, Tome I [retail].epub",
		"fr/Alexandre Dumas/2006. [Alexandre Dumas] (Le comte de Monte-Cristo, Tome I).epub",
	},
}

var standardTestConfig = Config{LibraryRoot: "."}
var isRetail = true


// TestJSON tests both JSON() and FromJSON().
func TestEpubJSON(t *testing.T) {
	fmt.Println("+ Testing Epub.JSON()...")
	for _, testEpub := range epubs {
		e := NewBook(testEpub.filename, standardTestConfig, isRetail)
		err := e.Metadata.Read(e.RetailEpub.Filename)
		if err != nil {
			t.Errorf("Error getting Metadata for epub %s", e.getMainFilename())
		}
		err = e.RetailEpub.GetHash()
		if err != nil {
			t.Errorf("Error getting Hash for epub %s", e.getMainFilename())
		}
		jsonString, err := e.JSON()
		if err != nil {
			t.Errorf("Error exporting epub %s to JSON string", e.getMainFilename())
		}
		if jsonString != testEpub.expectedJSONString {
			t.Errorf("JSON(%s) returned:\n%s\nexpected:\n%s!", testEpub.filename, jsonString, testEpub.expectedJSONString)
		}
		// recreating new Epub object from Json string
		f := Book{}
		f.FromJSON([]byte(jsonString))
		// comparing a few fields
		if e.Metadata.Get("title")[0] != e.Metadata.Get("title")[0] {
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

// TestTag tests AddTag, RemoveTag and HasTag
func TestEpubTag(t *testing.T) {
	fmt.Println("+ Testing Epub.AddTag()...")
	for _, testEpub := range epubs {
		e := NewBook(testEpub.filename, standardTestConfig, isRetail)
		tagName := "test_é!/?*èç1"

		err := e.AddTag(tagName)
		if err != nil {
			t.Errorf("Error adding Tag %s for epub %s", tagName, e.getMainFilename())
		}
		hasTag := e.HasTag(tagName)
		if !hasTag {
			t.Errorf("Error:  expected epub %s to have tag %s", e.getMainFilename(), tagName)
		}
		hasTag = e.HasTag(tagName + "A")
		if hasTag {
			t.Errorf("Error: did not expect epub %s to have tag %s", e.getMainFilename(), tagName+"A")
		}
		err = e.RemoveTag(tagName)
		if err != nil {
			t.Errorf("Error removing Tag %s for epub %s", tagName, e.getMainFilename())
		}
		hasTag = e.HasTag(tagName)
		if hasTag {
			t.Errorf("Error: did not expect epub %s to have tag %s", e.getMainFilename(), tagName)
		}
	}
}

func TestEpubNewName(t *testing.T) {
	fmt.Println("+ Testing Epub.generateNewName()...")
	for _, testEpub := range epubs {
		e := NewBook(testEpub.filename, standardTestConfig, isRetail)
		err := e.Metadata.Read(e.RetailEpub.Filename)
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.getMainFilename(), err)
		}

		newName1, err := e.generateNewName("$a $y $t", isRetail)
		if err != nil {
			t.Errorf("Error generating new name")
		}
		if newName1 != testEpub.expectedFormat1 {
			t.Errorf("Error getting new name, expected %s, got %s", testEpub.expectedFormat1, newName1)
		}
		newName2, err := e.generateNewName("$l/$a/$y. [$a] ($t)", isRetail)
		if err != nil {
			t.Errorf("Error generating new name")
		}
		if newName2 != testEpub.expectedFormat2 {
			t.Errorf("Error getting new name, expected %s, got %s", testEpub.expectedFormat2, newName2)
		}

		err = e.RetailEpub.SetRetail()
		if err != nil {
			t.Errorf("Error setting retail")
		}
		newName1, err = e.generateNewName("$a $y $t", isRetail)
		if err != nil {
			t.Errorf("Error generating new name")
		}
		if newName1 != testEpub.expectedFormat1Retail {
			t.Errorf("Error getting new name, expected %s, got %s", testEpub.expectedFormat1Retail, newName1)
		}
	}
}

func TestEpubRefresh(t *testing.T) {
	fmt.Println("+ Testing Epub.Refresh()...")
	c := Config{EpubFilenameFormat: "$a $y $t", LibraryRoot: "."}
	for _, testEpub := range epubs {

		// copy testEpub.filename
		epubFilename := filepath.Base(testEpub.filename)
		epubDir := filepath.Dir(testEpub.filename)
		tempCopy := filepath.Join(epubDir, "temp_"+epubFilename)

		err := CopyFile(testEpub.filename, tempCopy)
		if err != nil {
			t.Errorf("Error copying %s to %s", testEpub.filename, tempCopy)
		}

		// creating Epub object
		e := NewBook(tempCopy, c, isRetail)
		err = e.Metadata.Read(e.RetailEpub.Filename)
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.getMainFilename(), err)
		}
		fmt.Println(e.String())

		// refresh
		wasRenamed, newName, err := e.Refresh()
		fmt.Println()
		if err != nil {
			t.Errorf("Error generating new name: " + err.Error())
		}
		if !wasRenamed {
			t.Errorf("Error renaming %s", tempCopy)
		}
		if newName != testEpub.expectedFormat1 {
			t.Errorf("Error renaming %s, got %s, expected %s", tempCopy, newName, testEpub.expectedFormat1)
		}
		if newName != e.getMainFilename() {
			t.Errorf("Error setting new name %s, got %s, expected %s", tempCopy, newName, e.getMainFilename())
		}

		//  cleanup
		if err != nil || !wasRenamed {
			err = os.Remove(tempCopy)
			if err != nil {
				t.Errorf("Error removing temp copy %s", tempCopy)
			}
		} else {
			err = os.Remove(newName)
			if err != nil {
				t.Errorf("Error removing temp copy %s", newName)
			}
		}
	}
}

// TestEpubReadDate tests for SetReadDate and SetReadDateToday
func TestEpubSetReadDate(t *testing.T) {
	fmt.Println("+ Testing Epub.SetReadDate()...")
	for _, testEpub := range epubs {
		e := NewBook(testEpub.filename, standardTestConfig, isRetail)

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

// TestEpubProgress tests for SetProgress
func TestEpubProgress(t *testing.T) {
	fmt.Println("+ Testing Epub.TestEpubProgress()...")
	e := NewBook(epubs[0].filename, standardTestConfig, isRetail)

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

// TestEpubRetail tests for SetRetail, SetNonRetail and Check
func TestEpubRetail(t *testing.T) {
	fmt.Println("+ Testing Epub.SetRetail()...")
	e := NewBook(epubs[0].filename, standardTestConfig, isRetail)
	e.RetailEpub.GetHash()

	// testing retail
	err := e.RetailEpub.SetRetail()
	if err != nil {
		t.Errorf("Error setting retail")
	}
	if e.RetailEpub.Retail == "false" {
		t.Errorf("Error: ebook should be retail")
	}
	mode, err := os.Stat(e.getMainFilename())
	if mode.Mode() != 0444 {
		t.Errorf("Error: ebook should be read-only")
	}
	// checking retail
	hasChanged, err := e.RetailEpub.Check()
	if err != nil {
		t.Errorf("Error checking hash" + err.Error())
	}
	if hasChanged {
		t.Errorf("Error: ebook should be not have changed")
	}
	oldHash := e.RetailEpub.Hash
	e.RetailEpub.Hash = ""
	hasChanged, err = e.RetailEpub.Check()
	if err == nil {
		t.Errorf("Error checking retail hash, should have raised error")
	}
	if !hasChanged {
		t.Errorf("Error: ebook has changed")
	}

	// testing non-retail
	e.RetailEpub.Hash = oldHash
	err = e.RetailEpub.SetNonRetail()
	if err != nil {
		t.Errorf("Error setting non-retail")
	}
	if e.RetailEpub.Retail == "true" {
		t.Errorf("Error: ebook should not be retail")
	}
	mode, err = os.Stat(e.getMainFilename())
	if mode.Mode() != 0777 {
		t.Errorf("Error: ebook should be read-write")
	}

	// checking non retail
	hasChanged, err = e.RetailEpub.Check()
	if err != nil {
		t.Errorf("Error checking hash")
	}
	if hasChanged {
		t.Errorf("Error: ebook should be not have changed")
	}
	e.RetailEpub.Hash = ""
	hasChanged, err = e.RetailEpub.Check()
	if err != nil {
		t.Errorf("Error checking non retail hash, should have been ok")
	}
	if !hasChanged {
		t.Errorf("Error: ebook has changed")
	}
}
