package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	b "github.com/barsanuphe/endive/book"
	"github.com/barsanuphe/endive/db"
	en "github.com/barsanuphe/endive/endive"
	l "github.com/barsanuphe/endive/library"
	"github.com/barsanuphe/endive/mock"
)

func TestImportPaths(t *testing.T) {
	var book en.GenericBook
	assert := assert.New(t)

	// config
	c := en.Config{}
	c.LibraryRoot = "test/library"
	c.DatabaseFile = "test/library/endive_test.json"
	c.EpubFilenameFormat = "$a - $t"
	// makedirs c.LibraryRoot + defer removing all test files
	if err := os.MkdirAll(c.LibraryRoot, 0777); err != nil {
		panic(err)
	}
	defer os.RemoveAll(c.LibraryRoot)

	// building endive struct
	db := &db.JSONDB{}
	db.SetPath(c.DatabaseFile)
	ui := &mock.UserInterface{}
	lib := l.Library{Collection: &b.Books{}, Config: c, Index: &mock.IndexService{}, UI: ui, DB: db}
	err := lib.Load()
	assert.Nil(err, "Error loading epubs from database")
	k := en.KnownHashes{Filename: "test/library/test_hashes.json"}
	endive := Endive{hashes: k, Config: c, UI: ui, Library: lib}

	// the actual testing begins.

	fmt.Println("\n\t+ 1. import first nonretail")
	// modifying mock UI output
	importedFilename := filepath.Join(c.LibraryRoot, "unknown - Beowulf - An Anglo-Saxon Epic Poem.epub")
	// importing
	err = endive.ImportSpecific(false, "test/pg16328.epub")
	assert.Nil(err, "import should be successful")
	// testing file has been imported and renamed
	_, err = en.FileExists(importedFilename)
	assert.Nil(err, "Imported file should exist")
	// testing Library contains imported epub
	book, err = endive.Library.Collection.FindByFullPath(importedFilename)
	assert.Nil(err, "Imported epub should be in collection")
	assert.Equal(1, book.ID(), "First book should have ID 1.")

	fmt.Println("\n\t+ 2. import retail when nonretail exists")
	importedFilename = filepath.Join(c.LibraryRoot, "unknown - Beowulf - An Anglo-Saxon Epic Poem [retail].epub")
	// importing
	err = endive.ImportSpecific(true, "test/pg16328_empty.epub")
	assert.Nil(err, "import should be successful")
	// testing file has been imported and renamed
	_, err = en.FileExists(importedFilename)
	assert.Nil(err, "Imported file should exist")
	// testing Library contains imported epub
	book, err = endive.Library.Collection.FindByFullPath(importedFilename)
	assert.Nil(err, "Imported epub should be in collection")
	assert.Equal(1, book.ID(), "Trumped nonretail epub for book with ID 1.")

	fmt.Println("\n\t+ 3. import first retail again")
	err = endive.ImportSpecific(true, "test/pg16328_empty.epub")
	assert.NotNil(err, "import should not be successful")

	fmt.Println("\n\t+ 4. import first retail with imported file missing")
	// removing file
	err = os.Remove(importedFilename)
	assert.Nil(err, "Error removing retail epub")
	_, err = endive.Refresh()
	assert.Nil(err, "Error refreshing after removing retail epub")
	// importing (mock UI will agree to force import if missing)
	err = endive.ImportSpecific(true, "test/pg16328_empty.epub")
	assert.Nil(err, "import should be successful")
	// testing file has been imported and renamed
	_, err = en.FileExists(importedFilename)
	assert.Nil(err, "Imported file should exist")
	// testing Library contains imported epub
	book, err = endive.Library.Collection.FindByFullPath(importedFilename)
	assert.Nil(err, "Imported epub should be in collection")
	assert.Equal(1, book.ID(), "Second book should still have ID 1.")

	fmt.Println("\n\t+ 5. different epubs with different hashes but Config.EpubFileFormat returns the same filename")
	// change ebook format & propagate
	endive.Config.EpubFilenameFormat = "$y"
	endive.Library.Collection.Propagate(endive.UI, endive.Config)
	_, err = endive.Refresh()
	assert.Nil(err, "Error refreshing after removing retail epub")
	// modifying mock UI output
	ui.UpdateValuesResult = []string{"irrelevant"}
	importedFilename = filepath.Join(c.LibraryRoot, "2005 [retail].epub")
	// testing file has been imported and renamed
	_, err = en.FileExists(importedFilename)
	assert.Nil(err, "File "+importedFilename+" should exist")
	// importing
	err = endive.ImportSpecific(true, "test/pg16328_empty2.epub")
	assert.Nil(err, "import should be successful")
	// there should be 2 books now
	book, err = endive.Library.Collection.FindByID(2)
	assert.Nil(err, "Book ID2 exists")
	// testing the new imported epub exists (with random name)
	_, err = en.FileExists(book.FullPath())
	assert.Nil(err, "File "+book.FullPath()+" should exist")
	// testing the original lives on
	_, err = en.FileExists(importedFilename)
	assert.Nil(err, "File "+importedFilename+" should exist")

	// TODO test import a non retail and a second non retail for same book...
	// TODO test import a retail and a second retail for same book...
}

func TestImportSource(t *testing.T) {
	assert := assert.New(t)

	// config
	c := en.Config{}
	c.LibraryRoot = "test/library"
	c.DatabaseFile = "test/library/endive_test.json"
	c.RetailSource = []string{"test"}
	c.NonRetailSource = []string{"test"}
	c.EpubFilenameFormat = "$a - $t"

	// building endive struct
	db := &db.JSONDB{}
	db.SetPath(c.DatabaseFile)
	ui := &mock.UserInterface{}
	lib := l.Library{Collection: &b.Books{}, Config: c, Index: &mock.IndexService{}, UI: ui, DB: db}
	err := lib.Load()
	assert.Nil(err, "Error loading epubs from database")
	k := en.KnownHashes{Filename: "test/library/test_hashes.json"}
	endive := Endive{hashes: k, Config: c, UI: ui, Library: lib}

	// the actual testing begins.

	// analyzing
	candidates, err := endive.analyzeSources(endive.Config.RetailSource, true)
	assert.Nil(err, "import should be successful")
	assert.Equal(4, len(candidates), "Expected to find 4 epubs.")
	assert.Equal(4, len(candidates.Importable()), "Expected to find 4 epubs.")
}
