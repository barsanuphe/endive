package main

import (
	"testing"
	"io/ioutil"
	"bytes"
	"os"
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

func TestSave(t *testing.T) {
	l := LibraryDB{DatabaseFile: "test/db.json"}

	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}

	l.DatabaseFile = "test/db2.json"
	err = l.Save()
	if err != nil {
		t.Errorf("Error saving epubs to database: " + err.Error())
	}

	// compare both jsons
	db1, err := ioutil.ReadFile("test/db.json")
	db2, err2 := ioutil.ReadFile("test/db2.json")
	if err != nil || err2 != nil {
		t.Errorf("Error reading db file")
	}
	if !bytes.Equal(db1, db2) {
            t.Errorf("Error: original db != saved db")
        }
	// remove db2
	err = os.Remove("test/db2.json")
	if err != nil {
		t.Errorf("Error removing temp copy test/db2.json")
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
