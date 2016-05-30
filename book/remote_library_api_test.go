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
		"The Thorn of Emberlain",
		"8074907",
		"2016",
		"Scott Lynch (2016) The Thorn of Emberlain [Gentleman Bastard (#4)]",
		"9780575079588",
	},
	{
		"George Orwell",
		"Animal Farm: A Fairy Story",
		"7613",
		"1945",
		"George Orwell (1945) Animal Farm: A Fairy Story",
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
		bookID := g.GetBookIDByQuery(book.author, book.title, key)
		assert.Equal(book.expectedID, bookID, "Bad book id")
		// getting book information from book_id
		b := g.GetBook(bookID, key)
		b.Clean(standardTestConfig)
		assert.Equal(book.author, b.Author(), "Bad author")
		if b.MainTitle != book.title && b.OriginalTitle != book.title {
			t.Errorf("Bad title, got %s / %s, expected %s.", b.MainTitle, b.OriginalTitle, book.title)
		}
		assert.Equal(book.expectedYear, b.Year, "Bad year")
		assert.Equal(book.expectedFullTitle, b.String(), "Bad title")

		// getting book_id by isbn
		bookID = g.GetBookIDByISBN(book.isbn, key)
		assert.Equal(book.expectedID, bookID, "Bad book id")
	}
}
