package main

import "testing"

func TestSearch(t *testing.T) {
	l := LibraryDB{DatabaseFile: "test/db.json"}
	numIndexed, err := l.Index()
	if err != nil {
		t.Errorf("Error indexing epubs from database.")
	}
	if numIndexed != 3 {
		t.Errorf("Error indexing epubs from database, expected 2, got %d.", numIndexed)
	}

	l.Search("fr")
	l.Search("Description:fr")
	l.Search("Language:en")
	l.Search("Language:fr")
	l.Search("very old")
	l.Search("Description:very old")
	l.Search("poem")
	l.Search("Tags:poem")
	l.Search("Tags:not relevant")
	l.Search("Tags:littérature")
	l.Search("Tags:relevant")
	l.Search("relevant")
	l.Search("Tags:sf")
}
