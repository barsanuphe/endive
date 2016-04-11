package main

import (
	"errors"
	"fmt"
	"testing"
)

var metadataEpubs = []struct {
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
		"N/A",
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

// TestEpubMetaData tests GetMetadata and HasMetadata
func TestMetaData(t *testing.T) {
	fmt.Println("+ Testing Epub.GetMetaData()...")
	for _, testEpub := range metadataEpubs {
		e := NewBook(testEpub.filename, standardTestConfig, true)

		hasMetadata := e.Metadata.HasAny()
		if hasMetadata {
			t.Errorf("Error: %s should not have metadata yet.", e.getMainFilename())
		}

		err := e.Metadata.Read(e.RetailEpub.Filename)
		if err != nil {
			if testEpub.expectedError == nil {
				t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.getMainFilename(), err)
			}
			if err.Error() != testEpub.expectedError.Error() {
				t.Errorf("Error getting Metadata for %s, got %s, expected %s", e.getMainFilename(), err, testEpub.expectedError)
			}
		}
		if e.Metadata.Get("title")[0] != testEpub.expectedTitle {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Get("title")[0], testEpub.expectedTitle)
		}
		if e.Metadata.Get("creator")[0] != testEpub.expectedAuthor {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Get("creator")[0], testEpub.expectedAuthor)
		}
		if e.Metadata.Get("year")[0] != testEpub.expectedPublicationYear {
			t.Errorf("GetMetadata(%s) returned %d, expected %d!", testEpub.filename, e.Metadata.Get("year")[0], testEpub.expectedPublicationYear)
		}
		if e.Metadata.Get("language")[0] != testEpub.expectedLanguage {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Get("language")[0], testEpub.expectedLanguage)
		}

		hasMetadata = e.Metadata.HasAny()
		if !hasMetadata {
			t.Errorf("Error: %s should have metadata by now.", e.getMainFilename())
		}

		fmt.Println(e.String())
	}
}
