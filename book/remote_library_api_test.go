package book

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var grBooks = []struct {
	author            string
	title             string
	expectedID        string
	expectedYear      string
	expectedFullTitle string
	isbn              string
}{
	{
		"Scott Lynch",
		"The Republic of Thieves (Gentleman Bastard, #3)",
		"2890090",
		"2013",
		"Scott Lynch (2013) The Republic of Thieves (Gentleman Bastard, #3) [Gentleman Bastard #3]",
		"9780553804690",
	},
	{
		"George Orwell",
		"Animal Farm",
		"7613",
		"1945",
		"George Orwell (1945) Animal Farm",
		"9780452284241",
	},
}

// TestGoodReads tests goodreads search
func TestGoodReads(t *testing.T) {
	// make sure it is set
	key := os.Getenv("GR_API_KEY")
	require.NotEqual(t, 0, len(key), "Cannot get Goodreads API key")
	g := GoodReads{}
	assert := assert.New(t)
	for _, book := range grBooks {
		// getting book_id
		bookID, err := g.GetBookIDByQuery(book.author, book.title, key)
		assert.Nil(err, "Unexpected error")
		assert.Equal(book.expectedID, bookID, "Bad book id")

		// getting book information from book_id
		b, err := g.GetBook(bookID, key)
		assert.Nil(err, "Unexpected error")
		b.Clean(standardTestConfig)
		assert.Equal(book.author, b.Author(), "Bad author")
		if b.Title() != book.title {
			t.Errorf("Bad title, got %s, expected %s.", b.Title(), book.title)
		}
		assert.Equal(book.expectedYear, b.OriginalYear, "Bad year")
		assert.Equal(book.expectedFullTitle, b.String(), "Bad title")

		// getting book_id by isbn
		bookID, err = g.GetBookIDByISBN(book.isbn, key)
		assert.Nil(err, "Unexpected error")
		assert.Equal(book.expectedID, bookID, "Bad book id")
	}
}
