package book

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/barsanuphe/endive/endive"
	"github.com/barsanuphe/endive/mock"
)

const (
	expectedAllBooks       = "bad input, expected complete Books"
	expectedFirstBook      = "expected first book"
	expectedSecondBook     = "expected second book"
	expectedUIPropagation  = "expected ui to have been propagated"
	expectedCfgPropagation = "expected config to have been propagated"
	badInput               = "Bad input, expected error"
)

func TestBooks(t *testing.T) {
	fmt.Println("+ Testing Books...")
	assert := assert.New(t)
	books := Books{}
	for i, testEpub := range epubs {
		b := NewBook(ui, i+1, testEpub.filename, standardTestConfig, !isRetail)
		err := b.MainEpub().GetHash()
		assert.Nil(err)
		b.Metadata, err = b.MainEpub().ReadMetadata()
		assert.NotNil(err, "Error should be found (no ISBN in test epubs) for "+b.FullPath())
		if err != nil {
			assert.Equal("ISBN not found in epub", err.Error(), "Error should only mention missing ISBN")
		}
		books.Add(b)
	}
	assert.Equal(2, len(books), "2 books have been added")
	assert.Equal(2, len(books.Books()), "2 books have been added")

	// test Propagate()
	cfg := endive.Config{}
	ui := &mock.UserInterface{}
	books.Propagate(ui, cfg)
	assert.Equal(books[0].Config, cfg, expectedCfgPropagation)
	assert.Equal(books[0].UI, ui, expectedUIPropagation)
	assert.Equal(books[1].NonRetailEpub.Config, cfg, expectedCfgPropagation)
	assert.Equal(books[1].RetailEpub.UI, ui, expectedUIPropagation)

	// test First()
	assert.Equal(1, len(books.First(1).Books()), expectedFirstBook)
	assert.Equal(1, books.First(1).Books()[0].ID(), expectedFirstBook)
	assert.Equal(2, len(books.First(10).Books()), expectedAllBooks)
	assert.Equal(2, len(books.First(-1).Books()), expectedAllBooks)
	// test Last()
	assert.Equal(1, len(books.Last(1).Books()), expectedSecondBook)
	assert.Equal(2, books.Last(1).Books()[0].ID(), expectedSecondBook)
	assert.Equal(2, len(books.Last(10).Books()), expectedAllBooks)
	assert.Equal(2, len(books.Last(-1).Books()), expectedAllBooks)

	// FindByHash()
	_, err := books.FindByHash("")
	assert.NotNil(err, badInput)
	_, err = books.FindByHash("dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc04")
	assert.NotNil(err, badInput)
	book, err := books.FindByHash("dc325b3aceb77d9f943425728c037fdcaf4af58e3abd771a8094f2424455cc03")
	assert.Nil(err, "Correct hash, expected hit")
	assert.Equal(1, book.ID(), expectedFirstBook)

	// FindByID()
	_, err = books.FindByID(-1)
	assert.NotNil(err, badInput)
	_, err = books.FindByID(10)
	assert.NotNil(err, badInput)
	book, err = books.FindByID(2)
	assert.Nil(err, "Existing ID, expected hit")
	assert.Equal(epubs[1].filename, book.CleanFilename(), expectedSecondBook)

	// FindByFullPath()
	_, err = books.FindByFullPath("")
	assert.NotNil(err, badInput)
	_, err = books.FindByFullPath("test.epub")
	assert.NotNil(err, badInput)
	book, err = books.FindByFullPath(epubs[0].filename)
	assert.Nil(err, "existing path, expected hit")
	assert.Equal(1, book.ID(), expectedFirstBook)

	// FindByMetadata()
	_, err = books.FindByMetadata("", "", "")
	assert.NotNil(err, badInput)
	_, err = books.FindByMetadata("47", "", "")
	assert.NotNil(err, badInput)
	_, err = books.FindByMetadata("9780340839935", "", "")
	assert.NotNil(err, badInput)
	_, err = books.FindByMetadata("", "human", "book")
	assert.NotNil(err, badInput)
	_, err = books.FindByMetadata("", "Alexandre Dumas", "book")
	assert.NotNil(err, badInput)
	book, err = books.FindByMetadata("", "Alexandre Dumas", "Le comte de Monte-Cristo, Tome I")
	assert.Nil(err, "valid input, expected hit")
	assert.Equal(2, book.ID(), expectedSecondBook)

	// Authors()
	a := books.Authors()
	assert.Equal(2, len(a), "Expected 2 authors")
	num, ok := a[epubs[0].expectedAuthor]
	assert.True(ok, "author should be in map")
	assert.Equal(1, num, "author has 1 book")

	// Tags()
	tags := books.Tags()
	assert.Equal(7, len(tags))
	num, ok = tags["dragons -- poetry"]
	assert.True(ok, "tag should be in map")
	assert.Equal(1, num, "tag belongs to 1 book")

	// RemoveByID()
	err = books.RemoveByID(-1)
	assert.NotNil(err, badInput)
	err = books.RemoveByID(10)
	assert.NotNil(err, badInput)
	err = books.RemoveByID(1)
	assert.Nil(err, "book should be removed")
	assert.Equal(1, len(books), "Only one book should remain")
	assert.Equal(2, books[0].ID(), expectedSecondBook)
}
