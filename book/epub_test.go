package book

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)
		err := e.RetailEpub.GetHash()
		assert.Nil(err, "Error calculating hash for "+e.FullPath())
		assert.Equal(e.RetailEpub.Hash, testEpub.expectedSha256, "Error calculating sha256")
	}
}

func TestEpubFlagReplacement(t *testing.T) {
	fmt.Println("+ Testing Epub.FlagForReplacement()...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)
		assert.Equal(e.RetailEpub.NeedsReplacement, "false")

		err := e.RetailEpub.FlagForReplacement()
		assert.Nil(err, "Error flagging for replacement")
		assert.Equal(e.RetailEpub.NeedsReplacement, "true")
	}
}

// TestEpubCheck tests for Check
func TestEpubCheck(t *testing.T) {
	fmt.Println("+ Testing Epub.SetRetail()...")
	assert := assert.New(t)
	e := NewBook(0, epubs[0].filename, standardTestConfig, isRetail)
	e.RetailEpub.GetHash()

	// checking retail
	hasChanged, err := e.RetailEpub.Check()
	assert.Nil(err, "Error checking hash for "+e.FullPath())
	assert.False(hasChanged, "Error: ebook should be not have changed")

	oldHash := e.RetailEpub.Hash
	e.RetailEpub.Hash = ""
	hasChanged, err = e.RetailEpub.Check()
	assert.Nil(err, "Error checking retail hash")
	assert.True(hasChanged, "Error: ebook has changed")

	// testing non-retail
	e.RetailEpub.Hash = oldHash
	// checking non retail
	hasChanged, err = e.RetailEpub.Check()
	assert.Nil(err, "Error checking retail hash")
	assert.False(hasChanged, "Error: ebook should be not have changed")

	e.RetailEpub.Hash = ""
	hasChanged, err = e.RetailEpub.Check()
	assert.Nil(err, "Error checking non retail hash, should have been ok")
	assert.True(hasChanged, "Error: ebook has changed")
}

var isbns = []struct {
	candidate     string
	expectedISBN  string
	expectedError error
}{
	{
		"urn:ISBN: 12-2323-4-333-432  ",
		"1223234333432",
		nil,
	},
	{
		"1223234333432",
		"1223234333432",
		nil,
	},
	{
		"A223234333432",
		"",
		errors.New("ISBN-13 not found"),
	},
	{
		"urn:isbn: 12-23-4-333-432  ",
		"",
		errors.New("ISBN-13 not found"),
	},
}

func TestEpubCleanISBN(t *testing.T) {
	fmt.Println("+ Testing Info/CleanISBN()...")
	for _, c := range isbns {
		isbn, err := cleanISBN(c.candidate)
		if err == nil && c.expectedError != nil {
			t.Errorf("Unexpected error cleaning isbn %s", c.candidate)
		} else if err != nil && c.expectedError == nil {
			t.Errorf("Unexpected error cleaning isbn %s", c.candidate)
		} else if err != nil && c.expectedError != nil && c.expectedError.Error() != err.Error() {
			t.Errorf("Unexpected error cleaning isbn %s: got %s, expected %s", c.candidate, c.expectedError.Error(), err.Error())
		}
		assert.Equal(t, isbn, c.expectedISBN, "Error cleaning isbn")
	}
}
