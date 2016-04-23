package library

import (
	"testing"

	cfg "github.com/barsanuphe/endive/config"
)

func TestSearch(t *testing.T) {
	c := cfg.Config{}
	k := cfg.KnownHashes{}
	ldb := DB{DatabaseFile: "../test/endive.json"}
	l := Library{c, k, ldb}

	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}

	// search before indexing to check if index is built then.
	res, err := l.Search("fr")
	if err != nil {
		t.Errorf("Error searching fr")
	}
	if len(res) != 1 && res[0].FullPath() != "test/pg17989.epub" {
		t.Errorf("Error searching fr, unexpected results")
	}

	numIndexed, err := l.Index()
	if err != nil {
		t.Errorf("Error indexing epubs from database: %s", err.Error())
	}
	if numIndexed != 2 {
		t.Errorf("Error indexing epubs from database, expected 2, got %d.", numIndexed)
	}

	res, err = l.Search("metadata.language:fr")
	if err != nil {
		t.Errorf("Error searching fr")
	}
	if len(res) != 1 && res[0].FullPath() != "test/pg17989.epub" {
		t.Errorf("Error searching fr, unexpected results")
	}
	res, err = l.Search("metadata.authors:dumas")
	if err != nil {
		t.Errorf("Error searching for author:dumas " + err.Error())
	}
	if len(res) != 1 {
		t.Errorf("Error searching author:dumas, got %d hits, expected 1.", len(res))
	}
	res, err = l.Search("metadata.year:2005")
	if err != nil {
		t.Errorf("Error searching for year:2005")
	}
	if len(res) != 1 {
		t.Errorf("Error searching metadata.fields.year:2005, got %d hits, expected 1.", len(res))
	}
	// TODO search all fields

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
