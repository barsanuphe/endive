package book

import (
	"fmt"
	"testing"

	"github.com/barsanuphe/endive/endive"
	"github.com/stretchr/testify/assert"
)

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		e := NewBook(ui, i, testEpub.filename, standardTestConfig, true)
		err := e.RetailEpub.GetHash()
		assert.Nil(err, "Error calculating hash for "+e.FullPath())
		assert.Equal(e.RetailEpub.Hash, testEpub.expectedSha256, "Error calculating sha256")
	}
}

func TestEpubFlagReplacement(t *testing.T) {
	fmt.Println("+ Testing Epub.FlagForReplacement()...")
	assert := assert.New(t)
	for i, testEpub := range epubs {
		e := NewBook(ui, i, testEpub.filename, standardTestConfig, true)
		assert.Equal(e.RetailEpub.NeedsReplacement, endive.False)

		err := e.RetailEpub.FlagForReplacement()
		assert.Nil(err, "Error flagging for replacement")
		assert.Equal(e.RetailEpub.NeedsReplacement, endive.True)
	}
}

// TestEpubCheck tests for Check
func TestEpubCheck(t *testing.T) {
	fmt.Println("+ Testing Epub.SetRetail()...")
	assert := assert.New(t)
	e := NewBook(ui, 0, epubs[0].filename, standardTestConfig, isRetail)
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
