package main

import (
	"fmt"
	"testing"
)

// TestEpubMetaData tests GetMetadata and HasMetadata
func TestMetaData(t *testing.T) {
	fmt.Println("+ Testing Epub.GetMetaData()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)

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
