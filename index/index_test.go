package index

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	cfg "github.com/barsanuphe/endive/config"
	"github.com/barsanuphe/endive/endive"
	"github.com/barsanuphe/endive/library"
	"github.com/barsanuphe/endive/mock"
)

func TestSearch(t *testing.T) {
	indexPath := "../test/test_index"
	c := cfg.Config{}
	ui := &mock.UserInterface{}
	k := cfg.KnownHashes{}
	assert := assert.New(t)

	l := library.Library{Index: &Index{}, UI: ui, Config: c, KnownHashes: k, DatabaseFile: "../test/endive.json"}
	err := l.Load()
	assert.Nil(err, "Error loading epubs from database")
	l.Index.SetPath(indexPath)

	// search before indexing to check if index is built then.
	_, err = l.Index.Query("fr")
	assert.NotNil(err, "Index not built yet")

	// index
	// convert Books to []GenericBook
	allBooks := make([]endive.GenericBook, len(l.Books))
	for i := range l.Books {
		allBooks[i] = &l.Books[i]
	}
	err = l.Index.Rebuild(allBooks)
	assert.Nil(err, "Error indexing epubs from database")

	numIndexed := l.Index.Count()
	assert.EqualValues(2, numIndexed, "Error indexing epubs from database, expected 2")

	results, err := l.Index.Query("fr")
	assert.Nil(err, "Error opening index")
	assert.EqualValues(1, len(results), "Error searching fr, unexpected results")
	if len(results) >= 1 {
		assert.Equal("test/pg17989.epub", results[0], "Error searching fr, unexpected results")
	}

	// metadata.language:fr
	results, err = l.Index.Query("metadata.language:fr")
	assert.Nil(err, "Error searching language:fr")
	assert.Equal(1, len(results), "Error searching language:fr, unexpected results")
	if len(results) >= 1 {
		assert.Equal("test/pg17989.epub", results[0], "Error searching language:fr, unexpected results")
	}
	// metadata.authors:dumas
	results, err = l.Index.Query("metadata.authors:dumas")
	assert.Nil(err, "Error searching author:dumas")
	assert.EqualValues(1, len(results), "Error searching author:dumas, unexpected results")
	if len(results) >= 1 {
		assert.Equal("test/pg17989.epub", results[0], "Error searching author:dumas, unexpected results")
	}
	// metadata.year:2005
	results, err = l.Index.Query("metadata.year:2005")
	assert.Nil(err, "Error searching year:2005")
	assert.EqualValues(1, len(results), "Error searching year:2005, unexpected results")
	// metadata.year:2205
	results, err = l.Index.Query("metadata.year:2205")
	assert.Nil(err, "Error searching year:2205")
	assert.EqualValues(0, len(results), "Error searching year:2205, did not expect results")

	// remove index
	err = os.RemoveAll(indexPath)
	if err != nil {
		assert.Nil(err, "Error removing index")
	}

	// TODO search all fields
	/*
		  		l.Search("en")
				l.Search("language:en")
				l.Search("Dumas")
				l.Search("author:Dumas")
				l.Search("Author:Dumas")
				l.Search("title:Beowulf")
				l.Search("author:Beowulf")
				l.Search("tags:littérature")
				l.Search("tags:sf")
	*/
}
