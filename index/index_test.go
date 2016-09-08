package index

// TODO mock GenericBook to test this.
/*
import (
	"testing"

	cfg "github.com/barsanuphe/endive/config"

	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	c := cfg.Config{}
	k := cfg.KnownHashes{}
	ldb := DB{DatabaseFile: "../test/endive.json"}
	assert := assert.New(t)

	l := Library{c, k, ldb}
	err := l.Load()
	assert.Nil(err, "Error loading epubs from database")

	// search before indexing to check if index is built then.
	res, err := l.RunQuery("fr")
	assert.Nil(err, "Error searching fr")
	assert.EqualValues(len(res), 1, "Error searching fr, unexpected results")
	assert.Equal(res[0].FullPath(), "test/pg17989.epub", "Error searching fr, unexpected results")
	// index
	numIndexed, err := l.IndexAll()
	assert.Nil(err, "Error indexing epubs from database")
	assert.EqualValues(numIndexed, 2, "Error indexing epubs from database, expected 2")
	// metadata.language:fr
	res, err = l.RunQuery("language:fr")
	assert.Nil(err, "Error searching language:fr")
	assert.Equal(len(res), 1, "Error searching language:fr, unexpected results")
	assert.Equal(res[0].FullPath(), "test/pg17989.epub", "Error searching language:fr, unexpected results")
	// metadata.authors:dumas
	res, err = l.RunQuery("author:dumas")
	assert.Nil(err, "Error searching author:dumas")
	assert.EqualValues(len(res), 1, "Error searching author:dumas, unexpected results")
	assert.Equal(res[0].FullPath(), "test/pg17989.epub", "Error searching author:dumas, unexpected results")
	// metadata.year:2005
	res, err = l.RunQuery("year:2005")
	assert.Nil(err, "Error searching year:2005")
	assert.EqualValues(len(res), 1, "Error searching year:2005, unexpected results")

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

}
*/
