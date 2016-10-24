package book

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	validValue   = "Valid value, expected no error."
	invalidValue = "Invalid value, should have triggered an error."
)

func TestMetadata(t *testing.T) {
	fmt.Println("+ Testing MetaData...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		e := NewBook(ui, i, testEpub.filename, standardTestConfig, true)

		// testing HasAny
		hasMetadata := e.Metadata.HasAny()
		assert.False(hasMetadata, "Error, should not have metadata yet.")

		// reading info
		info, err := e.MainEpub().ReadMetadata()
		assert.NotNil(err, "Error should be found (no ISBN in test epubs) for "+e.FullPath())
		if err != nil {
			assert.Equal("ISBN not found in epub", err.Error(), "Error should only mention missing ISBN")
		}
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
		o.BookTitle = e.Metadata.BookTitle
		// checking again
		assert.True(e.Metadata.IsSimilar(o), "Error: metadata should be similar.")
	}
}

func TestMetadataSet(t *testing.T) {
	fmt.Println("+ Testing MetaData.Set()...")
	assert := assert.New(t)
	e := NewBook(ui, 0, epubs[0].filename, standardTestConfig, isRetail)

	// set unknown field
	err := e.Metadata.Set("rating", "a")
	assert.NotNil(err, invalidValue)

	// Set ISBN
	err = e.Metadata.Set("ISBN", "hihi")
	assert.NotNil(err, invalidValue)
	err = e.Metadata.Set("isbn", "9780340839935")
	assert.Nil(err, validValue)
	assert.Equal("9780340839935", e.Metadata.ISBN)
	err = e.Metadata.Set("isbn", "0340839937")
	assert.Nil(err, validValue)
	assert.Equal("9780340839935", e.Metadata.ISBN)
	err = e.Metadata.Set("ISBN", "9-78-0-4--410-13-593")
	assert.Nil(err, validValue)
	assert.Equal("9780441013593", e.Metadata.ISBN)

	// Set Tags
	err = e.Metadata.Set("tags", "hihi, HOHO")
	assert.Nil(err)
	assert.Equal(2, len(e.Metadata.Tags), "Metadata should have 2 tags")
	assert.Equal("hihi, hoho", e.Metadata.Tags.String())
	err = e.Metadata.Set("tags", "")
	assert.Nil(err)
	assert.Equal("", e.Metadata.Tags.String())

	// Set Series
	err = e.Metadata.Set("series", "hihi, HOHO")
	assert.Nil(err)
	assert.Equal(2, len(e.Metadata.Series), "Metadata should have 2 series")
	assert.Equal("hihi #0, HOHO #0", e.Metadata.Series.String())
	err = e.Metadata.Set("series", "hihi:0.5, HOHO:7-9")
	assert.Nil(err)
	assert.Equal(2, len(e.Metadata.Series), "Metadata should have 2 series")
	assert.Equal("hihi #0.5, HOHO #7,8,9", e.Metadata.Series.String())
	err = e.Metadata.Set("series", "")
	assert.Nil(err)
	assert.Equal("", e.Metadata.Series.String())

	// Set authors
	err = e.Metadata.Set("author", "  hihi , HOHO  ")
	assert.Nil(err)
	assert.Equal(2, len(e.Metadata.Authors), "Metadata should have 2 authors")
	assert.Equal("hihi, HOHO", e.Metadata.Author())

	// Set years
	err = e.Metadata.Set("year", "hihi")
	assert.NotNil(err, invalidValue)
	err = e.Metadata.Set("edition_year", "2013")
	assert.Nil(err, validValue)
	assert.Equal("2013", e.Metadata.EditionYear)

	// Set category
	err = e.Metadata.Set("category", "hihi")
	assert.NotNil(err, invalidValue)
	err = e.Metadata.Set("category", "NonFiction")
	assert.Nil(err, validValue)
	for _, vc := range validCategories {
		err = e.Metadata.Set("category", vc)
		assert.Nil(err, validValue)
		assert.Equal(vc, e.Metadata.Category)
	}

	// Set type
	err = e.Metadata.Set("type", "hihi")
	assert.NotNil(err, invalidValue)
	for _, vt := range validTypes {
		err = e.Metadata.Set("type", vt)
		assert.Nil(err, validValue)
		assert.Equal(vt, e.Metadata.Type)
	}

	// Set description
	err = e.Metadata.Set("description", "simple description")
	assert.Nil(err, validValue)
	assert.Equal("simple description", e.Metadata.Description)
	err = e.Metadata.Set("description", `simple <a href="link">description</a>`)
	assert.Nil(err, validValue)
	assert.Equal("simple description", e.Metadata.Description)

	// Set language
	err = e.Metadata.Set("language", "eng")
	assert.Nil(err, validValue)
	assert.Equal("en", e.Metadata.Language)

	// set simple field
	err = e.Metadata.Set("publisher", "m. publisher")
	assert.Nil(err, validValue)

}
