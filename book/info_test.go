package book

import (
	"fmt"
	"testing"
)

func TestInfo(t *testing.T) {
	fmt.Println("+ Testing Epub.GetMetaData()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)

		// testing HasAny
		hasMetadata := e.Metadata.HasAny()
		if hasMetadata {
			t.Errorf("Error: %s should not have metadata yet.", e.FullPath())
		}
		// reading info
		info, err := e.MainEpub().ReadMetadata()
		if err != nil {
			t.Errorf("Error getting Metadata for %s, got %s, expected nil", e.FullPath(), err)
		}
		e.EpubMetadata = info
		e.Metadata = info

		// testing Get, GetFirstValue
		if e.Metadata.Title() != testEpub.expectedTitle {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Title(), testEpub.expectedTitle)
		}
		if e.Metadata.Author() != testEpub.expectedAuthor {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Author(), testEpub.expectedAuthor)
		}
		if e.Metadata.Year != testEpub.expectedPublicationYear {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Year, testEpub.expectedPublicationYear)
		}
		if e.Metadata.Language != testEpub.expectedLanguage {
			t.Errorf("GetMetadata(%s) returned %s, expected %s!", testEpub.filename, e.Metadata.Language, testEpub.expectedLanguage)
		}

		// testing HasAny
		hasMetadata = e.Metadata.HasAny()
		if !hasMetadata {
			t.Errorf("Error: %s should have metadata by now.", e.FullPath())
		}

		// testing IsSimilar
		o := Info{}
		if e.Metadata.IsSimilar(o) {
			t.Errorf("Error: metadata should not be similar.")
		}
		// copying manually
		o.Authors = []string{}
		o.Authors = append(o.Authors, e.Metadata.Authors...)
		o.MainTitle = e.Metadata.MainTitle
		// checking again
		if !e.Metadata.IsSimilar(o) {
			t.Errorf("Error: metadata should be similar.")
		}

	}
}
