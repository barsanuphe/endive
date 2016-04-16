package book

import (
	"fmt"
	"testing"
)

// TestEpubMetaData tests GetMetadata and HasMetadata
func TestMetaData(t *testing.T) {
	fmt.Println("+ Testing Epub.GetMetaData()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)

		// testing HasAny
		hasMetadata := e.Metadata.HasAny()
		if hasMetadata {
			t.Errorf("Error: %s should not have metadata yet.", e.GetMainFilename())
		}

		// testing Read
		err := e.Metadata.Read(e.GetMainFilename())
		if err != nil {
			if testEpub.expectedError == nil {
				t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.GetMainFilename(), err)
			}
			if err.Error() != testEpub.expectedError.Error() {
				t.Errorf("Error getting Metadata for %s, got %s, expected %s", e.GetMainFilename(), err, testEpub.expectedError)
			}
		}
		// testing Get, GetFirstValue
		if e.Metadata.Get("title")[0] != testEpub.expectedTitle {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Get("title")[0], testEpub.expectedTitle)
		}
		if e.Metadata.GetFirstValue("title") != testEpub.expectedTitle {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.GetFirstValue("title"), testEpub.expectedTitle)
		}
		if e.Metadata.Get("creator")[0] != testEpub.expectedAuthor {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Get("creator")[0], testEpub.expectedAuthor)
		}
		if e.Metadata.Get("year")[0] != testEpub.expectedPublicationYear {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Get("year")[0], testEpub.expectedPublicationYear)
		}
		if e.Metadata.Get("language")[0] != testEpub.expectedLanguage {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Get("language")[0], testEpub.expectedLanguage)
		}

		// testing HasAny
		hasMetadata = e.Metadata.HasAny()
		if !hasMetadata {
			t.Errorf("Error: %s should have metadata by now.", e.GetMainFilename())
		}

		// testing IsSimilar
		o := NewMetadata()
		if e.Metadata.IsSimilar(o) {
			t.Errorf("Error: metadata should not be similar.")
		}
		// copying manually
		o.Fields["creator"] = []string{}
		o.Fields["creator"] = append(o.Fields["creator"], e.Metadata.Get("creator")...)
		o.Fields["title"] = []string{}
		o.Fields["title"] = append(o.Fields["title"], e.Metadata.Get("title")...)
		// checking again
		if !e.Metadata.IsSimilar(o) {
			t.Errorf("Error: metadata should be similar.")
		}

	}
}
