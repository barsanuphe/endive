package library

import "testing"

func TestSearch(t *testing.T) {
	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}

	// search before indexing to check if index is built then.
	res, err := l.Search("fr")
	if err != nil {
		t.Errorf("Error searching fr")
	}
	if len(res) != 1 && res[0].GetMainFilename() != "test/pg17989.epub" {
		t.Errorf("Error searching fr, unexpected results")
	}

	numIndexed, err := l.Index()
	if err != nil {
		t.Errorf("Error indexing epubs from database: %s", err.Error())
	}
	if numIndexed != 2 {
		t.Errorf("Error indexing epubs from database, expected 2, got %d.", numIndexed)
	}

	res, err = l.Search("fr")
	if err != nil {
		t.Errorf("Error searching fr")
	}
	if len(res) != 1 && res[0].GetMainFilename() != "test/pg17989.epub" {
		t.Errorf("Error searching fr, unexpected results")
	}
	res, err = l.Search("metadata.fields.creator:dumas")
	if err != nil {
		t.Errorf("Error searching for metadata.fields.creator:dumas " + err.Error())
	}
	if len(res) != 1 {
		t.Errorf("Error searching metadata.fields.creator:dumas, got %d hits, expected 1.", len(res))
	}
	res, err = l.Search("metadata.fields.year:2005")
	if err != nil {
		t.Errorf("Error searching for metadata.fields.year:2005")
	}
	if len(res) != 1 {
		t.Errorf("Error searching metadata.fields.year:2005, got %d hits, expected 1.", len(res))
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
