package main

import (
	"errors"
	"fmt"
	"testing"
)

var epubs = []struct {
	filename                string
	expectedError           error
	expectedTitle           string
	expectedAuthor          string
	expectedPublicationYear int
	expectedLanguage        string
	expectedSha256          string
	expectedJSONString	string
}{
	{
		"test/pg17989.epub",
		nil,
		"Le comte de Monte-Cristo, Tome I",
		"Alexandre Dumas",
		2006,
		"fr",
		"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a",
		`{"filename":"test/pg17989.epub","relativepath":"","hash":"acd2b8eba1b11456bacf11e690edf56bc57774053668644ef34f669138ebdd9a","isretail":false,"progress":0,"series":null,"author":"Alexandre Dumas","title":"Le comte de Monte-Cristo, Tome I","language":"fr","publicationyear":2006,"readdate":"","tags":null,"rating":0,"review":""}`,
	},
	{
		"test/pg16328.epub",
		errors.New("Metadata field creator does not exist"),
		"Beowulf / An Anglo-Saxon Epic Poem",
		"Unknown",
		2005,
		"en",
		"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03",
		`{"filename":"test/pg16328.epub","relativepath":"","hash":"dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03","isretail":false,"progress":0,"series":null,"author":"Unknown","title":"Beowulf / An Anglo-Saxon Epic Poem","language":"en","publicationyear":2005,"readdate":"","tags":null,"rating":0,"review":""}`,
	},
}

func TestEpubGetMetaData(t *testing.T) {
	fmt.Println("+ Testing Epub.GetMetaData()...")
	for _, test_epub := range epubs {
		e := Epub{Filename: test_epub.filename}
		err := e.GetMetadata()
		if err != nil {
			if test_epub.expectedError == nil {
				t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.Filename, err)
			}
			if err.Error() != test_epub.expectedError.Error() {
				t.Errorf("Error getting Metadata for %s, got %s, expected %s", e.Filename, err, test_epub.expectedError)
			}
		}
		if e.Title != test_epub.expectedTitle {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", test_epub.filename, e.Title, test_epub.expectedTitle)
		}
		if e.Author != test_epub.expectedAuthor {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", test_epub.filename, e.Author, test_epub.expectedAuthor)
		}
		if e.PublicationYear != test_epub.expectedPublicationYear {
			t.Errorf("GetMetadata(%s) returned %d, expected %d!", test_epub.filename, e.PublicationYear, test_epub.expectedPublicationYear)
		}
		if e.Language != test_epub.expectedLanguage {
			t.Errorf("GetMetadata(%s) returned %d, expected %d!", test_epub.filename, e.Language, test_epub.expectedLanguage)
		}
		fmt.Println(e.String())
	}
}

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	for _, test_epub := range epubs {
		e := Epub{Filename: test_epub.filename}
		err := e.GetHash()
		if err != nil {
			t.Errorf("Error calculating hash for %s", e.Filename)
		}
		if e.Hash != test_epub.expectedSha256 {
			t.Errorf("GetHash(%s) returned %s, expected %s!", test_epub.filename, e.Hash, test_epub.expectedSha256)
		}
	}
}

// TestJSON tests both JSON() and FromJSON().
func TestEpubJSON(t *testing.T) {
	fmt.Println("+ Testing Epub.JSON()...")
	for _, test_epub := range epubs {
		e := Epub{Filename: test_epub.filename}
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
		if jsonString != test_epub.expectedJSONString {
			t.Errorf("JSON(%s) returned:\n%s\nexpected:\n%s!", test_epub.filename, jsonString, test_epub.expectedJSONString)
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
	for _, test_epub := range epubs {
		e := Epub{Filename: test_epub.filename}
		tagName := "test_é!/?*èç1"

		err := e.AddTag(tagName)
		if err != nil {
			t.Errorf("Error adding Tag %s for epub %s", tagName, e.Filename)
		}
		hasTag := e.HasTag(tagName)
		if ! hasTag {
			t.Errorf("Error:  expected epub %s to have tag %s", e.Filename, tagName)
		}
		hasTag = e.HasTag(tagName+"A")
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



