package library

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	b "github.com/barsanuphe/endive/book"
	"github.com/barsanuphe/endive/db"
	e "github.com/barsanuphe/endive/endive"
	"github.com/barsanuphe/endive/mock"
)

/*
func TestLibrarySearch(t *testing.T) {
	c := e.Config{}
	k := e.KnownHashes{}
	l := Library{Config: c, KnownHashes: k, DatabaseFile: "../test/endive.json", Index: &mock.IndexService{}, UI: &mock.UserInterface{}}
	assert := assert.New(t)

	err := l.Load()
	assert.Nil(err, "Error loading epubs from database")
	results, err := l.SearchAndPrint("language:fr", "default", false, false, 0)
	assert.Nil(err, "Error running query")
	fmt.Println(results)
	// TODO search all fields to check replacements
}
*/

const (
	root       = ".."
	mountPoint = "../test/ereader"
	b1Filename = "test/pg16328.epub"
	b2Filename = "test/pg17989.epub"
	dbFilename = "../test/endive.json"

	errExpectedExport   = "File should be exported to ereader mount point"
	errUnexpectedExport = "File should not have been exported to ereader mount point"
	errExportOK         = "Export should be ok."
	errBookInLibrary    = "Book with ID %d exists."
	errExpectedMarked   = "Book with ID %d should be marked as exported"
	errUnexpectedMarked = "Book with ID %d should *not* be marked as exported"
)

func TestExport(t *testing.T) {
	assert := assert.New(t)

	// create Config with LibraryRoot + EReaderMountPoint
	c := e.Config{}
	c.LibraryRoot = root
	c.EReaderMountPoint = mountPoint
	// create mount point
	if err := os.MkdirAll(c.EReaderMountPoint, 0777); err != nil {
		panic(err)
	}
	defer os.RemoveAll(c.EReaderMountPoint)

	ui := &mock.UserInterface{}
	db := &db.JSONDB{}
	db.SetPath(dbFilename)

	// create collection for Library, with 2 epubs
	libB1Filename := filepath.Join(root, b1Filename)
	exportedB1Filename := filepath.Join(c.EReaderMountPoint, b1Filename)
	libB2Filename := filepath.Join(root, b2Filename)
	exportedB2Filename := filepath.Join(c.EReaderMountPoint, b2Filename)
	b1 := b.NewBook(ui, 1, libB1Filename, c, true)
	b2 := b.NewBook(ui, 2, libB2Filename, c, true)
	var collection e.Collection
	collection = &b.Books{}
	collection.Add(b1, b2)

	// create Library with Config + Collection
	l := Library{Collection: collection, Index: &mock.IndexService{}, UI: ui, Config: c, DB: db}
	err := l.Load()
	assert.Nil(err, "Error loading epubs from database")

	// Export first epub
	err = l.ExportToEReader(l.Collection.First(1), c.EReaderMountPoint)
	assert.Nil(err, errExportOK)
	// check right epub was copied
	_, exists := e.FileExists(exportedB1Filename)
	assert.Nil(exists, errExpectedExport)
	_, exists = e.FileExists(exportedB2Filename)
	assert.NotNil(exists, errUnexpectedExport)
	// check right epub was marked as exported
	lb1, err := l.Collection.FindByID(1)
	assert.Nil(err, fmt.Sprintf(errBookInLibrary, 1))
	assert.Equal(e.True, lb1.(*b.Book).IsExported, fmt.Sprintf(errExpectedMarked, 1))
	lb2, err := l.Collection.FindByID(2)
	assert.Nil(err, fmt.Sprintf(errBookInLibrary, 2))
	assert.NotEqual(e.True, lb2.(*b.Book).IsExported, fmt.Sprintf(errUnexpectedMarked, 2))

	// copy the 2nd epub manually
	err = e.CopyFile(libB2Filename, exportedB2Filename)
	assert.Nil(err, "Book with ID2 should be copied without any problem.")

	// export again
	err = l.ExportToEReader(l.Collection.First(1), "")
	assert.Nil(err, errExportOK)
	// check both epubs were copied
	_, exists = e.FileExists(exportedB1Filename)
	assert.Nil(exists, errExpectedExport)
	_, exists = e.FileExists(exportedB2Filename)
	assert.Nil(exists, errExpectedExport)
	// check both epubs were marked as exported
	lb1, err = l.Collection.FindByID(1)
	assert.Nil(err, fmt.Sprintf(errBookInLibrary, 1))
	assert.Equal(e.True, lb1.(*b.Book).IsExported, fmt.Sprintf(errExpectedMarked, 1))
	lb2, err = l.Collection.FindByID(2)
	assert.Nil(err, fmt.Sprintf(errBookInLibrary, 2))
	assert.Equal(e.True, lb2.(*b.Book).IsExported, fmt.Sprintf(errExpectedMarked, 2))
}

func TestGenerateID(t *testing.T) {
	assert := assert.New(t)

	c := e.Config{}
	ui := &mock.UserInterface{}
	db := &db.JSONDB{}
	l := Library{Collection: &b.Books{}, Index: &mock.IndexService{}, UI: ui, Config: c, DB: db}

	// first book
	id := l.GenerateID()
	assert.Equal(1, id, "First book should have ID 1")
	// second book
	l.Collection.Add(b.NewBook(ui, 1, "", c, true))
	id = l.GenerateID()
	assert.Equal(2, id, "Second book should have ID 2")
	// after ID 1789
	l.Collection.Add(b.NewBook(ui, 1789, "", c, true))
	id = l.GenerateID()
	assert.Equal(1790, id, "ID shoudl be 1789+1")
}
