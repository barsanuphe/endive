package db

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/barsanuphe/endive/book"
	"github.com/barsanuphe/endive/endive"
)

var testDbName = "../test/endive.json"

func TestLdbLoad(t *testing.T) {
	assert := assert.New(t)

	db := JSONDB{}
	db.SetPath(testDbName)

	var collection endive.Collection
	collection = &book.Books{}
	err := db.Load(collection)
	assert.Nil(err, "Error loading epubs from database")
	assert.Equal(2, len(collection.Books()), "Error loading epubs, expected 2 epubs")
	for _, epub := range collection.Books() {
		// convert to Book
		var b book.Book
		b = *epub.(*book.Book)
		hasMetadata := b.Metadata.HasAny()
		assert.True(hasMetadata, "Error loading epubs, epub %s does not have metadata in db", epub.FullPath())
	}
}

func TestLdbSave(t *testing.T) {
	assert := assert.New(t)
	tempTestDbName := "../test/db2.json"

	db := JSONDB{}
	db.SetPath(testDbName)

	var collection endive.Collection
	collection = &book.Books{}
	err := db.Load(collection)
	assert.Nil(err, "Error loading epubs from database")

	// save unchanged
	hasSaved, err := db.Save(collection)
	assert.Nil(err, "Error saving epubs to database")
	assert.False(hasSaved, "Error, db should not have been saved")

	// changing DatabaseFile will make Save() compare current db with an
	// empty file, forcing save + new index
	db.SetPath(tempTestDbName)
	hasSaved, err = db.Save(collection)
	assert.Nil(err, "Error saving epubs to database")
	assert.True(hasSaved, "Error saving to database")

	// compare both jsons
	db1, err := ioutil.ReadFile(testDbName)
	db2, err2 := ioutil.ReadFile(tempTestDbName)
	assert.Nil(err, "Error reading db file")
	assert.Nil(err2, "Error reading db file")
	assert.True(bytes.Equal(db1, db2), "Error: original db != saved db")

	// remove db2
	err = os.Remove(tempTestDbName)
	assert.Nil(err, "Error removing temp copy test/db2.json")
}
