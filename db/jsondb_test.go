package db

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/libgit2/git2go"
	"github.com/stretchr/testify/assert"

	"github.com/barsanuphe/endive/book"
	"github.com/barsanuphe/endive/endive"
)

var testDbName = "../test/endive.json"

func TestDBLoad(t *testing.T) {
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

func TestDBSave(t *testing.T) {
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

func TestDBBackup(t *testing.T) {
	assert := assert.New(t)
	libraryRoot := "../test/library"
	databaseFile := "../test/library/endive_test.json"

	db := JSONDB{}
	db.SetPath(filepath.Base(databaseFile))

	// makedirs c.LibraryRoot + defer removing all test files
	assert.Nil(os.MkdirAll(libraryRoot, 0777))
	defer os.RemoveAll(libraryRoot)
	// copy test/endive.json inside
	assert.Nil(endive.CopyFile(testDbName, databaseFile))

	// db.Backup(newdir)
	assert.Nil(db.Backup(libraryRoot))
	// assert .git exists and git log has 1 commit
	assert.True(endive.DirectoryExists(filepath.Join(libraryRoot, ".git")))
	repo, err := git.OpenRepository(libraryRoot)
	assert.Nil(err)
	head, err := repo.Head()
	assert.Nil(err)
	headCommit, err := repo.LookupCommit(head.Target())
	assert.Nil(err)
	assert.Equal(uint(0), headCommit.ParentCount())

	// backup again
	assert.Nil(db.Backup(libraryRoot))
	// assert repo has 2 commits (ie head commit has 1 parent)
	assert.True(endive.DirectoryExists(filepath.Join(libraryRoot, ".git")))
	head, err = repo.Head()
	assert.Nil(err)
	headCommit, err = repo.LookupCommit(head.Target())
	assert.Nil(err)
	assert.Equal(uint(1), headCommit.ParentCount())

}
