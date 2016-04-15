package library

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	b "github.com/barsanuphe/endive/book"
	cfg "github.com/barsanuphe/endive/config"
	"launchpad.net/go-xdg"
)

const xdgIndexPath string = cfg.Endive + "/" + cfg.Endive + ".index"

// getIndexPath gets the default index path
func getIndexPath() (path string) {
	path, err := xdg.Cache.Find(xdgIndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			path = filepath.Join(xdg.Cache.Dirs()[0], xdgIndexPath)
		} else {
			panic(err)
		}
	}
	return
}

// LibraryDB manages the epub database and search
type LibraryDB struct {
	DatabaseFile string
	IndexFile    string // can be in XDG data path
	Books        []b.Book
}

// Load current DB
func (ldb *LibraryDB) Load() (err error) {
	fmt.Println("Loading database...")
	bytes, err := ioutil.ReadFile(ldb.DatabaseFile)
	if err != nil {
		if os.IsNotExist(err) {
			// first run
			return nil
		}
		return
	}
	err = json.Unmarshal(bytes, &ldb.Books)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// Save current DB
func (ldb *LibraryDB) Save() (hasSaved bool, err error) {
	jsonEpub, err := json.Marshal(ldb.Books)
	if err != nil {
		fmt.Println(err)
		return
	}
	// compare with input
	jsonEpubOld, err := ioutil.ReadFile(ldb.DatabaseFile)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
		return
	}

	if !bytes.Equal(jsonEpub, jsonEpubOld) {
		fmt.Println("Changes detected, saving database...")
		// writing db
		err = ioutil.WriteFile(ldb.DatabaseFile, jsonEpub, 0777)
		if err != nil {
			return
		}
		hasSaved = true
		// remove old index
		err = os.RemoveAll(getIndexPath())
		if err != nil {
			fmt.Println(err)
			return hasSaved, err
		}

		// indexing db
		numIndexed, err := ldb.Index()
		if err != nil {
			return hasSaved, err
		}
		fmt.Println("Saved and indexed " + strconv.FormatUint(numIndexed, 10) + " epubs.")
	}
	return
}

// generateID for a new Book
func (ldb *LibraryDB) generateID() (id int) {
	// id 0 for first Book
	if len(ldb.Books) == 0 {
		return
	}
	// find max ID of ldb.Books
	for _, book := range ldb.Books {
		if book.ID > id {
			id = book.ID
		}
	}
	id++
	return
}

// FindByID among known Books
func (ldb *LibraryDB) FindByID(id int) (result *b.Book, err error) {
	for i, bk := range ldb.Books {
		if bk.ID == id {
			return &ldb.Books[i], nil
		}
	}
	return &b.Book{}, errors.New("Could not find book with ID " + strconv.Itoa(id))
}

//FindByFilename among known Books
func (ldb *LibraryDB) FindByFilename(filename string) (result *b.Book, err error) {
	for i, bk := range ldb.Books {
		if bk.RetailEpub.Filename == filename || bk.NonRetailEpub.Filename == filename {
			return &ldb.Books[i], nil
		}
	}
	return &b.Book{}, errors.New("Could not find book with epub " + filename)
}

// SearchJSON current DB and output as JSON
func (ldb *LibraryDB) SearchJSON() (jsonOutput string, err error) {
	// TODO Search() then get JSON output from each result Epub
	// TODO OR --- the opposite. bleve can return JSON, Search has to parse it and locate the relevant Epub objects
	fmt.Println("Searching database with JSON output...")
	return
}

// ListNonRetailOnly among known epubs.
func (ldb *LibraryDB) ListNonRetailOnly() (nonretail []b.Book, err error) {
	// TODO return Search for querying non retail epubs, removing the epubs with same title/author but retail
	return
}

// ListRetailOnly among known epubs.
func (ldb *LibraryDB) ListRetailOnly() (retail []b.Book, err error) {
	return
}

// ListAuthors among known epubs.
func (ldb *LibraryDB) ListAuthors() (authors []string, err error) {
	return
}

// ListTags associated with known epubs.
func (ldb *LibraryDB) ListTags() (tags []string, err error) {
	// TODO search for tags in all epubs, remove duplicates
	return
}

// ListUntagged among known epubs.
func (ldb *LibraryDB) ListUntagged() (untagged []b.Book, err error) {
	return
}

// ListWithTag among known epubs.
func (ldb *LibraryDB) ListWithTag(tag string) (tagged []b.Book, err error) {
	return
}
