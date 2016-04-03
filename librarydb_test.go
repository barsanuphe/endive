package main

import (
	"testing"
)

func TestLoad(t *testing.T) {
	l := LibraryDB{DatabaseFile: "test/db.json"}

	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}
	if len(l.Epubs) != 2 {
		t.Errorf("Error loading epubs, expected 2 epubs, got %d: ", len(l.Epubs))
	}
	for _, epub := range l.Epubs {
		if hasMetadata := epub.HasMetadata(); !hasMetadata {
			t.Errorf("Error loading epubs, epub %s does not have metadata in db", epub.Filename)
		}
	}
}

func TestSearch(t *testing.T) {
	l := LibraryDB{DatabaseFile: "test/db.json"}

	numIndexed, err := l.Index()
	if err != nil {
		t.Errorf("Error indexing epubs from database.")
	}
	if numIndexed != 2 {
		t.Errorf("Error indexing epubs from database, expected 2, got %d.", numIndexed)
	}

	l.Search("fr")
	l.Search("en")
	l.Search("language:en")
	l.Search("language:fr")
	l.Search("Dumas")
	l.Search("author:Dumas")
	l.Search("author:dumas")
	l.Search("Author:Dumas")

	l.Search("title:Beowulf")
	l.Search("author:Beowulf")

	l.Search("tags:littérature")
	l.Search("tags:sf")
}
