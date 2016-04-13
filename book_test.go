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
		`{"id":0,"retail":{"filename":"test/pg16328.epub","hash":"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03","replace":"false"},"nonretail":{"filename":"","hash":"","replace":""},"metadata":{"fields":{"contributor":["J. Lesslie Hall"],"coverage":["Unknown"],"creator":["Unknown"],"description":["Unknown"],"format":["Unknown"],"identifier":["http://www.gutenberg.org/ebooks/16328"],"language":["en"],"meta":["Unknown"],"publisher":["Unknown"],"relation":["Unknown"],"rights":["Public domain in the USA."],"source":["http://www.gutenberg.org/files/16328/16328-h/16328-h.htm"],"subject":["Epic poetry, English (Old)","Monsters -- Poetry","Dragons -- Poetry"],"title":["Beowulf / An Anglo-Saxon Epic Poem"],"type":["Unknown"],"year":["2005"]}},"series":null,"tags":null,"progress":"unread","readdate":"","rating":"","review":"","description":""}`,
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
		`{"id":1,"retail":{"filename":"test/pg17989.epub","hash":"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a","replace":"false"},"nonretail":{"filename":"","hash":"","replace":""},"metadata":{"fields":{"contributor":["Unknown"],"coverage":["Unknown"],"creator":["Alexandre Dumas"],"description":["Unknown"],"format":["Unknown"],"identifier":["http://www.gutenberg.org/ebooks/17989"],"language":["fr"],"meta":["Unknown"],"publisher":["Unknown"],"relation":["Unknown"],"rights":["Public domain in the USA."],"source":["http://www.gutenberg.org/files/17989/17989-h/17989-h.htm"],"subject":["Historical fiction","Revenge -- Fiction","Adventure stories","Prisoners -- Fiction","France -- History -- 19th century -- Fiction","Pirates -- Fiction","Dantès, Edmond (Fictitious character) -- Fiction"],"title":["Le comte de Monte-Cristo, Tome I"],"type":["Unknown"],"year":["2006"]}},"series":null,"tags":null,"progress":"unread","readdate":"","rating":"","review":"","description":""}`,
		"Alexandre Dumas 2006 Le comte de Monte-Cristo, Tome I.epub",
		"Alexandre Dumas 2006 Le comte de Monte-Cristo, Tome I [retail].epub",
		"fr/Alexandre Dumas/2006. [Alexandre Dumas] (Le comte de Monte-Cristo, Tome I).epub",
	},
}
var standardTestConfig = Config{LibraryRoot: "."}
var isRetail = true

// TestBookJSON tests both JSON() and FromJSON().
func TestBookJSON(t *testing.T) {
	fmt.Println("+ Testing Epub.JSON()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, isRetail)
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

// TestBookTag tests AddTag, RemoveTag and HasTag
func TestBookTag(t *testing.T) {
	fmt.Println("+ Testing Epub.AddTag()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, isRetail)
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

func TestBookNewName(t *testing.T) {
	fmt.Println("+ Testing Epub.generateNewName()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, !isRetail)
		err := e.Metadata.Read(e.NonRetailEpub.Filename)
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.getMainFilename(), err)
		}

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
		err = e.Metadata.Read(e.RetailEpub.Filename)
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.getMainFilename(), err)
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

func TestBookRefresh(t *testing.T) {
	fmt.Println("+ Testing Epub.Refresh()...")
	c := Config{EpubFilenameFormat: "$a $y $t", LibraryRoot: "."}
	for i, testEpub := range epubs {
		// copy testEpub.filename
		epubFilename := filepath.Base(testEpub.filename)
		epubDir := filepath.Dir(testEpub.filename)
		tempCopy := filepath.Join(epubDir, "temp_"+epubFilename)

		err := CopyFile(testEpub.filename, tempCopy)
		if err != nil {
			t.Errorf("Error copying %s to %s", testEpub.filename, tempCopy)
		}

		// creating Epub object
		e := NewBook(i, tempCopy, c, isRetail)
		err = e.Metadata.Read(e.RetailEpub.Filename)
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.getMainFilename(), err)
		}

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
		if newName[0] != e.getMainFilename() {
			t.Errorf("Error setting new name %s, got %s, expected %s", tempCopy, newName[0], e.getMainFilename())
		}

		//  cleanup
		if err != nil || !wasRenamed[0] {
			err = os.Remove(tempCopy)
			if err != nil {
				t.Errorf("Error removing temp copy %s", tempCopy)
			}
		} else {
			err = os.Remove(newName[0])
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
