package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

var l = LibraryDB{DatabaseFile: "test/endive.json"}

func TestLdbLoad(t *testing.T) {
	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}
	if len(l.Books) != 2 {
		t.Errorf("Error loading epubs, expected 2 epubs, got %d: ", len(l.Books))
	}
	for _, epub := range l.Books {
		if hasMetadata := epub.Metadata.HasAny(); !hasMetadata {
			t.Errorf("Error loading epubs, epub %s does not have metadata in db", epub.getMainFilename())
		}
	}
}

func TestLdbSave(t *testing.T) {
	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}

	// save unchanged
	hasSaved, err := l.Save()
	if err != nil {
		t.Errorf("Error saving epubs to database: " + err.Error())
	}
	if hasSaved {
		t.Errorf("Error, db should not have been saved")
	}

	// changing DatabaseFile will make Save() compare current db with an
	// empty file, forcing save + new index
	l.DatabaseFile = "test/db2.json"
	hasSaved, err = l.Save()
	if err != nil {
		t.Errorf("Error saving epubs to database: " + err.Error())
	}
	if !hasSaved {
		t.Errorf("Error saving epubs to database")
	}

	// compare both jsons
	db1, err := ioutil.ReadFile("test/endive.json")
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
	l.DatabaseFile = "test/endive.json"
}

func TestLdbSearch(t *testing.T) {
	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}

	numIndexed, err := l.Index()
	if err != nil {
		t.Errorf("Error indexing epubs from database: %s", err.Error())
	}
	if numIndexed != 2 {
		t.Errorf("Error indexing epubs from database, expected 2, got %d.", numIndexed)
	}

	res, err := l.Search("fr")
	if err != nil {
		t.Errorf("Error searching fr")
	}
	if len(res) != 1 && res[0].getMainFilename() != "test/pg17989.epub" {
		t.Errorf("Error searching fr, unexpected results")
	}
	res, err = l.Search("author:dumas")
	if err != nil {
		t.Errorf("Error searching for rating: " + err.Error())
	}
	if len(res) != 1 {
		t.Errorf("Error searching rating:0, got %d hits, expected 1.", len(res))
	}
	res, err = l.Search("publicationyear:2005")
	if err != nil {
		t.Errorf("Error searching for publicationyear:2005")
	}
	if len(res) != 1 {
		t.Errorf("Error searching rating:0, got %d hits, expected 1.", len(res))
	}

	/*
		res, err = l.Search("publicationyear:2005")
		if err != nil {
			t.Errorf("Error searching fr")
		}
		fmt.Println(res)
		res, err = l.Search("2005")
		if err != nil {
			t.Errorf("Error searching fr")
		}
		fmt.Println(res)
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
	*/
}
