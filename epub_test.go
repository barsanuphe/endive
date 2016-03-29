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
}{
	{
		"test/pg17989.epub",
		nil,
		"Le comte de Monte-Cristo, Tome I",
		"Alexandre Dumas",
		2006,
	},
	{
		"test/pg16328.epub",
		errors.New("Metadata field creator does not exist"),
		"Beowulf / An Anglo-Saxon Epic Poem",
		"Unknown",
		2005,
	},
}

func TestGetMetaData(t *testing.T) {
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
		fmt.Println(e.String())
	}
}
