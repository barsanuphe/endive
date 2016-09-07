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
	"time"

	b "github.com/barsanuphe/endive/book"
	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"

	"github.com/jhoonb/archivex"
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

// DB manages the epub database and search
type DB struct {
	DatabaseFile         string
	IndexFile            string // can be in XDG data path
	Books                b.Books
	indexNeedsRebuilding bool
}

// load JSON contents into Books
func (ldb *DB) loadBooks() (bks b.Books, jsonContent []byte, err error) {
	jsonContent, err = ioutil.ReadFile(ldb.DatabaseFile)
	if err != nil {
		if os.IsNotExist(err) {
			// first run
			return
		}
		return
	}
	err = json.Unmarshal(jsonContent, &bks)
	if err != nil {
		fmt.Println(err)
	}
	return
}

// Load current DB
func (ldb *DB) Load() (err error) {
	h.Debug("Loading database...")
	ldb.Books, _, err = ldb.loadBooks()
	return err
}

// Save current DB
func (ldb *DB) Save() (hasSaved bool, err error) {
	h.Debug("Determining if database should be saved...")
	jsonEpub, err := json.Marshal(ldb.Books)
	if err != nil {
		fmt.Println(err)
		return
	}

	// compare with input
	oldBooks, jsonEpubOld, err := ldb.loadBooks()
	if err != nil && !os.IsNotExist(err) {
		h.Error("Error loading database")
	}

	if !bytes.Equal(jsonEpub, jsonEpubOld) {
		h.Debug("Changes detected, saving database...")
		// writing db
		err = ioutil.WriteFile(ldb.DatabaseFile, jsonEpub, 0777)
		if err != nil {
			return
		}
		hasSaved = true

		// index what is needed.
		// diff to check the changes
		n, m, d := ldb.Books.Diff(oldBooks)
		// update the index
		err = ldb.IndexDiff(n, m, d)
		if err != nil {
			h.Error("Error updating index, it may be necessary to build it anew")
			ldb.indexNeedsRebuilding = true
			return
		}
	}
	return
}

// Backup current database.
func (ldb *DB) Backup() (err error) {
	h.Debug("Backup up database...")
	// generate archive filename with date.
	archiveName, err := cfg.GetArchiveUniqueName(ldb.DatabaseFile)
	if err != nil {
		return
	}
	// creating tarball
	tar := new(archivex.TarFile)
	err = tar.Create(archiveName)
	if err != nil {
		return
	}
	err = tar.AddFile(ldb.DatabaseFile)
	if err != nil {
		return
	}
	tar.Close()
	return
}

func (ldb *DB) rebuildIndex() (err error) {
	// remove old index
	err = os.RemoveAll(getIndexPath())
	if err != nil {
		return err
	}
	// indexing db
	numIndexed, err := ldb.IndexAll()
	if err != nil {
		return err
	}
	h.Debug("Indexed " + strconv.FormatUint(numIndexed, 10) + " epubs.")
	return
}

// generateID for a new Book
func (ldb *DB) generateID() (id int) {
	// id 0 for first Book
	if len(ldb.Books) == 0 {
		return
	}
	// find max ID of ldb.Books and add 1
	for _, book := range ldb.Books {
		if book.ID > id {
			id = book.ID
		}
	}
	id++
	return
}

// Check all Books
func (ldb *DB) Check() (err error) {
	defer h.TimeTrack(time.Now(), "Checking")
	for i := range ldb.Books {
		h.Debug("Checking " + ldb.Books[i].ShortString())
		retailChanged, nonRetailChanged, err := ldb.Books[i].Check()
		if err != nil {
			return err
		}
		if retailChanged {
			err = errors.New("Retail epub has changed for book: " + ldb.Books[i].ShortString())
			return err
		}
		if nonRetailChanged {
			h.Warning("Non-retail epub for book " + ldb.Books[i].ShortString() + " has changed, check if this is normal.")
		}
	}
	return
}

// FindByMetadata among known Books
func (ldb *DB) FindByMetadata(i b.Metadata) (result *b.Book, err error) {
	// TODO tests
	for j, book := range ldb.Books {
		if book.Metadata.IsSimilar(i) || book.EpubMetadata.IsSimilar(i) {
			return &ldb.Books[j], nil
		}
	}
	return &b.Book{}, errors.New("Could not find book with info " + i.String())
}

//FindByHash among known Books
func (ldb *DB) FindByHash(hash string) (result *b.Book, err error) {
	// TODO check valid hash?
	for i, bk := range ldb.Books {
		if bk.RetailEpub.Hash == hash || bk.NonRetailEpub.Hash == hash {
			return &ldb.Books[i], nil
		}
	}
	return &b.Book{}, errors.New("Could not find book with hash " + hash)
}

// RemoveByID a book from the db
func (ldb *DB) RemoveByID(id int) (err error) {
	var found bool
	removeIndex := -1
	for i := range ldb.Books {
		if ldb.Books[i].ID == id {
			found = true
			removeIndex = i
			break
		}
	}
	if found {
		h.Info("REMOVING from db " + ldb.Books[removeIndex].ShortString())
		ldb.Books = append((ldb.Books)[:removeIndex], (ldb.Books)[removeIndex+1:]...)
	} else {
		err = errors.New("Did not find book with ID " + strconv.Itoa(id))
	}
	return
}

// ListNonRetailOnly among known epubs.
func (ldb *DB) ListNonRetailOnly() b.Books {
	return ldb.Books.FilterNonRetailOnly()
}

// ListRetail among known epubs.
func (ldb *DB) ListRetail() b.Books {
	return ldb.Books.FilterRetail()
}

// ListAuthors among known epubs.
func (ldb *DB) ListAuthors() (authors map[string]int) {
	authors = make(map[string]int)
	for _, book := range ldb.Books {
		author := book.Metadata.Author()
		authors[author]++
	}
	return
}

// ListPublishers among known epubs.
func (ldb *DB) ListPublishers() (publishers map[string]int) {
	publishers = make(map[string]int)
	for _, book := range ldb.Books {
		if book.Metadata.Publisher != "" {
			publisher := book.Metadata.Publisher
			publishers[publisher]++
		} else {
			publishers["Unknown"]++
		}

	}
	return
}

// ListTags associated with known epubs.
func (ldb *DB) ListTags() (tags map[string]int) {
	tags = make(map[string]int)
	for _, book := range ldb.Books {
		for _, tag := range book.Metadata.Tags {
			tags[tag.Name]++
		}
	}
	return
}

// ListSeries associated with known epubs.
func (ldb *DB) ListSeries() (series map[string]int) {
	series = make(map[string]int)
	for _, book := range ldb.Books {
		for _, s := range book.Metadata.Series {
			series[s.Name]++
		}
	}
	return
}

// ListUntagged among known epubs.
func (ldb *DB) ListUntagged() b.Books {
	return ldb.Books.FilterUntagged()
}

// ListByProgress returns a slice of Books with the given reading progress.
func (ldb *DB) ListByProgress(progress string) b.Books {
	return ldb.Books.FilterByProgress(progress)
}

// ListIncomplete among known epubs.
func (ldb *DB) ListIncomplete() b.Books {
	return ldb.Books.FilterIncomplete()
}
