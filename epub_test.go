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

var standardConfig = Config{LibraryRoot: "."}

// TestEpubMetaData tests GetMetadata and HasMetadata
func TestEpubMetaData(t *testing.T) {
	fmt.Println("+ Testing Epub.GetMetaData()...")
	for _, testEpub := range epubs {
		e := NewEpub(testEpub.filename, standardConfig)

		hasMetadata := e.HasMetadata()
		if hasMetadata {
			t.Errorf("Error: %s should not have metadata yet.", e.Filename)
		}

		err := e.GetMetadata()
		if err != nil {
			if testEpub.expectedError == nil {
				t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.Filename, err)
			}
			if err.Error() != testEpub.expectedError.Error() {
				t.Errorf("Error getting Metadata for %s, got %s, expected %s", e.Filename, err, testEpub.expectedError)
			}
		}
		if e.Title != testEpub.expectedTitle {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Title, testEpub.expectedTitle)
		}
		if e.Author != testEpub.expectedAuthor {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Author, testEpub.expectedAuthor)
		}
		if e.PublicationYear != testEpub.expectedPublicationYear {
			t.Errorf("GetMetadata(%s) returned %d, expected %d!", testEpub.filename, e.PublicationYear, testEpub.expectedPublicationYear)
		}
		if e.Language != testEpub.expectedLanguage {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Language, testEpub.expectedLanguage)
		}

		hasMetadata = e.HasMetadata()
		if !hasMetadata {
			t.Errorf("Error: %s should have metadata by now.", e.Filename)
		}

		fmt.Println(e.String())
	}
}

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	for _, testEpub := range epubs {
		e := NewEpub(testEpub.filename, standardConfig)
		err := e.GetHash()
		if err != nil {
			t.Errorf("Error calculating hash for %s", e.Filename)
		}
		if e.Hash != testEpub.expectedSha256 {
			t.Errorf("GetHash(%s) returned %s, expected %s!", testEpub.filename, e.Hash, testEpub.expectedSha256)
		}
	}
}

// TestJSON tests both JSON() and FromJSON().
func TestEpubJSON(t *testing.T) {
	fmt.Println("+ Testing Epub.JSON()...")
	for _, testEpub := range epubs {
		e := NewEpub(testEpub.filename, standardConfig)
		err := e.GetMetadata()
		if err != nil {
			t.Errorf("Error getting Metadata for epub %s", e.Filename)
		}
		err = e.GetHash()
		if err != nil {
			t.Errorf("Error getting Hash for epub %s", e.Filename)
		}
		jsonString, err := e.JSON()
		if err != nil {
			t.Errorf("Error exporting epub %s to JSON string", e.Filename)
		}
		if jsonString != testEpub.expectedJSONString {
			t.Errorf("JSON(%s) returned:\n%s\nexpected:\n%s!", testEpub.filename, jsonString, testEpub.expectedJSONString)
		}
		// recreating new Epub object from Json string
		f := Epub{}
		f.FromJSON([]byte(jsonString))
		// comparing a few fields
		if e.Author != f.Author && e.Title != f.Title {
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
		e := NewEpub(testEpub.filename, standardConfig)
		tagName := "test_é!/?*èç1"

		err := e.AddTag(tagName)
		if err != nil {
			t.Errorf("Error adding Tag %s for epub %s", tagName, e.Filename)
		}
		hasTag := e.HasTag(tagName)
		if !hasTag {
			t.Errorf("Error:  expected epub %s to have tag %s", e.Filename, tagName)
		}
		hasTag = e.HasTag(tagName + "A")
		if hasTag {
			t.Errorf("Error: did not expect epub %s to have tag %s", e.Filename, tagName+"A")
		}
		err = e.RemoveTag(tagName)
		if err != nil {
			t.Errorf("Error removing Tag %s for epub %s", tagName, e.Filename)
		}
		hasTag = e.HasTag(tagName)
		if hasTag {
			t.Errorf("Error: did not expect epub %s to have tag %s", e.Filename, tagName)
		}
	}
}

// TestEpubSeries tests AddSeries, RemoveSeries and HasSeries
func TestEpubSeries(t *testing.T) {
	fmt.Println("+ Testing Epub.AddSeries()...")
	for i, testEpub := range epubs {
		e := NewEpub(testEpub.filename, standardConfig)
		seriesName := "test_é!/?*èç1"
		seriesName2 := "test2"

		// testing adding series
		seriesModified := e.AddSeries(seriesName, float32(i))
		if !seriesModified {
			t.Errorf("Error adding Series %s - %f for epub %s", seriesName, float32(i), e.Filename)
		}
		// testing adding second series
		seriesModified = e.AddSeries(seriesName2, float32(i))
		if !seriesModified {
			t.Errorf("Error adding Series %s - %f for epub %s", seriesName2, float32(i), e.Filename)
		}

		// testing having series
		hasSeries, index, seriesIndex := e.HasSeries(seriesName)
		if !hasSeries {
			t.Errorf("Error:  expected epub %s to have series %s", e.Filename, seriesName)
		}
		if index != 0 {
			t.Errorf("Error:  expected epub %s to have series %s at index 0", e.Filename, seriesName)
		}
		if seriesIndex != float32(i) {
			t.Errorf("Error:  expected epub %s to have series %s, book %f and not %f", e.Filename, seriesName, float32(i), seriesIndex)
		}
		hasSeries, index, seriesIndex = e.HasSeries(seriesName2)
		if !hasSeries {
			t.Errorf("Error:  expected epub %s to have series %s", e.Filename, seriesName2)
		}
		if index != 1 {
			t.Errorf("Error:  expected epub %s to have series %s at index 1", e.Filename, seriesName2)
		}
		if seriesIndex != float32(i) {
			t.Errorf("Error:  expected epub %s to have series %s, book %f and not %f", e.Filename, seriesName2, float32(i), seriesIndex)
		}

		hasSeries, _, _ = e.HasSeries(seriesName + "ç")
		if hasSeries {
			t.Errorf("Error:  did not expect epub %s to have series %s", e.Filename, seriesName+"ç")
		}

		// testing updating series index
		seriesModified = e.AddSeries(seriesName, float32(i)+0.5)
		if !seriesModified {
			t.Errorf("Error adding Series %s - %f for epub %s", seriesName, float32(i)+0.5, e.Filename)
		}
		// testing having modified series
		hasSeries, index, seriesIndex = e.HasSeries(seriesName)
		if !hasSeries {
			t.Errorf("Error:  expected epub %s to have series %s", e.Filename, seriesName)
		}
		if index != 0 {
			t.Errorf("Error:  expected epub %s to have series %s at index 0", e.Filename, seriesName)
		}
		if seriesIndex != float32(i)+0.5 {
			t.Errorf("Error:  expected epub %s to have series %s, book %f and not %f", e.Filename, seriesName, float32(i)+0.5, seriesIndex)
		}

		// testing removing series
		seriesRemoved := e.RemoveSeries(seriesName)
		if !seriesRemoved {
			t.Errorf("Error removing Series %s for epub %s", seriesName, e.Filename)
		}
		hasSeries, _, _ = e.HasSeries(seriesName)
		if hasSeries {
			t.Errorf("Error: did not expect epub %s to have series %s", e.Filename, seriesName)
		}
	}
}

func TestEpubNewName(t *testing.T) {
	fmt.Println("+ Testing Epub.generateNewName()...")
	for _, testEpub := range epubs {
		e := NewEpub(testEpub.filename, standardConfig)
		err := e.GetMetadata()
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.Filename, err)
		}

		newName1, err := e.generateNewName("$a $y $t")
		if err != nil {
			t.Errorf("Error generating new name")
		}
		if newName1 != testEpub.expectedFormat1 {
			t.Errorf("Error getting new name, expected %s, got %s", testEpub.expectedFormat1, newName1)
		}
		newName2, err := e.generateNewName("$l/$a/$y. [$a] ($t)")
		if err != nil {
			t.Errorf("Error generating new name")
		}
		if newName2 != testEpub.expectedFormat2 {
			t.Errorf("Error getting new name, expected %s, got %s", testEpub.expectedFormat2, newName2)
		}

		err = e.SetRetail()
		if err != nil {
			t.Errorf("Error setting retail")
		}
		newName1, err = e.generateNewName("$a $y $t")
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
		e := NewEpub(tempCopy, c)
		err = e.GetMetadata()
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.Filename, err)
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
		if newName != e.Filename {
			t.Errorf("Error setting new name %s, got %s, expected %s", tempCopy, newName, e.Filename)
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
		e := NewEpub(testEpub.filename, standardConfig)

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
	e := NewEpub(epubs[0].filename, standardConfig)

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
	e := NewEpub(epubs[0].filename, standardConfig)
	e.GetHash()

	// testing retail
	err := e.SetRetail()
	if err != nil {
		t.Errorf("Error setting retail")
	}
	if e.IsRetail == "false" {
		t.Errorf("Error: ebook should be retail")
	}
	mode, err := os.Stat(e.Filename)
	if mode.Mode() != 0444 {
		t.Errorf("Error: ebook should be read-only")
	}
	// checking retail
	hasChanged, err := e.Check()
	if err != nil {
		t.Errorf("Error checking hash" + err.Error())
	}
	if hasChanged {
		t.Errorf("Error: ebook should be not have changed")
	}
	oldHash := e.Hash
	e.Hash = ""
	hasChanged, err = e.Check()
	if err == nil {
		t.Errorf("Error checking retail hash, should have raised error")
	}
	if !hasChanged {
		t.Errorf("Error: ebook has changed")
	}

	// testing non-retail
	e.Hash = oldHash
	err = e.SetNonRetail()
	if err != nil {
		t.Errorf("Error setting non-retail")
	}
	if e.IsRetail == "true" {
		t.Errorf("Error: ebook should not be retail")
	}
	mode, err = os.Stat(e.Filename)
	if mode.Mode() != 0777 {
		t.Errorf("Error: ebook should be read-write")
	}

	// checking non retail
	hasChanged, err = e.Check()
	if err != nil {
		t.Errorf("Error checking hash")
	}
	if hasChanged {
		t.Errorf("Error: ebook should be not have changed")
	}
	e.Hash = ""
	hasChanged, err = e.Check()
	if err != nil {
		t.Errorf("Error checking non retail hash, should have been ok")
	}
	if !hasChanged {
		t.Errorf("Error: ebook has changed")
	}
}
