package book

import (
	"os"
	"testing"
)

var grBooks = []struct {
	gr_id             string
	expectedYear      string
	expectedFullTitle string
}{
	{
		"8074907",
		"2016",
		"Scott Lynch (2016) The Thorn of Emberlain [Gentleman Bastard #4]",
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

	for _, book := range grBooks {
		b := GetBook(book.gr_id, key)

		if b.Year != book.expectedYear {
			t.Errorf("Bad year, got %s, expected %s.", b.Year, book.expectedYear)
		}
		if b.FullTitle() != book.expectedFullTitle {
			t.Errorf("Bad title, got %s, expected %s.", b.FullTitle(), book.expectedFullTitle)
			t.Errorf("Bad title, got %s, expected %s.", b.FullTitle(), book.expectedFullTitle)
		}

	}
}
