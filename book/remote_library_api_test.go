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
	require.NotEqual(t, len(key), 0, "Cannot get Goodreads API key")
	g := GoodReads{}
	assert := assert.New(t)
	for _, book := range grBooks {
		// getting book_id
		bookID := g.GetBookIDByQuery(book.author, book.title, key)
		assert.Equal(bookID, book.expectedID, "Bad book id")
		// getting book information from book_id
		b := g.GetBook(bookID, key)
		assert.Equal(b.Author(), book.author, "Bad author")
		if b.MainTitle != book.title && b.OriginalTitle != book.title {
			t.Errorf("Bad title, got %s / %s, expected %s.", b.MainTitle, b.OriginalTitle, book.title)
		}
		assert.Equal(b.Year, book.expectedYear, "Bad year")
		assert.Equal(b.String(), book.expectedFullTitle, "Bad title")

		// getting book_id by isbn
		bookID = g.GetBookIDByISBN(book.isbn, key)
		assert.Equal(bookID, book.expectedID, "Bad book id")
	}
}
