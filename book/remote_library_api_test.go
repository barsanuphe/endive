package book

import (
	"os"
	"testing"
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
	if len(key) == 0 {
		t.Error("Cannot get Goodreads API key")
		t.FailNow()
	}

	g := GoodReads{}
	for _, book := range grBooks {
		// getting book_id
		bookID := g.GetBookIDByQuery(book.author, book.title, key)
		if bookID != book.expectedID {
			t.Errorf("Bad book id, got %s, expected %s.", bookID, book.expectedID)
		}
		// getting book information from book_id
		b := g.GetBook(bookID, key)
		if b.Author() != book.author {
			t.Errorf("Bad author, got %s, expected %s.", b.Author(), book.author)
		}
		if b.MainTitle != book.title && b.OriginalTitle != book.title {
			t.Errorf("Bad title, got %s / %s, expected %s.", b.MainTitle, b.OriginalTitle, book.title)
		}
		if b.Year != book.expectedYear {
			t.Errorf("Bad year, got %s, expected %s.", b.Year, book.expectedYear)
		}
		if b.String() != book.expectedFullTitle {
			t.Errorf("Bad title, got %s, expected %s.", b.String(), book.expectedFullTitle)
		}

		// getting book_id by isbn
		bookID = g.GetBookIDByISBN(book.isbn, key)
		if bookID != book.expectedID {
			t.Errorf("Bad book id, got %s, expected %s.", bookID, book.expectedID)
		}
	}

}
