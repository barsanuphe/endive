package book

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	fmt.Println("+ Testing Epub.GetMetaData()...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)

		// testing HasAny
		hasMetadata := e.Metadata.HasAny()
		assert.False(hasMetadata, "Error, should not have metadata yet.")

		// reading info
		info, err := e.MainEpub().ReadMetadata()
		assert.Nil(err, "Error getting Metadata for  + ", e.FullPath())
		e.EpubMetadata = info
		e.Metadata = info

		// testing Get, GetFirstValue
		assert.Equal(e.Metadata.Title(), testEpub.expectedTitle, "Error getting title")
		assert.Equal(e.Metadata.Author(), testEpub.expectedAuthor, "Error getting author")
		assert.Equal(e.Metadata.OriginalYear, testEpub.expectedPublicationYear, "Error getting year")
		assert.Equal(e.Metadata.Language, testEpub.expectedLanguage, "Error getting language")

		// testing HasAny
		hasMetadata = e.Metadata.HasAny()
		assert.True(hasMetadata, "Error, should have metadata")

		// testing IsSimilar
		o := Metadata{}
		assert.False(e.Metadata.IsSimilar(o), "Error: metadata should not be similar.")

		// copying manually
		o.Authors = []string{}
		o.Authors = append(o.Authors, e.Metadata.Authors...)
		o.MainTitle = e.Metadata.MainTitle
		// checking again
		assert.True(e.Metadata.IsSimilar(o), "Error: metadata should be similar.")
	}
}
