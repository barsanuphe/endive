package library

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	b "github.com/barsanuphe/endive/book"
	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"

	e "github.com/barsanuphe/endive/endive"
	"github.com/jhoonb/archivex"
)

// load JSON contents into Books
func (l *Library) loadBooks() (bks b.Books, jsonContent []byte, err error) {
	jsonContent, err = ioutil.ReadFile(l.DatabaseFile)
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
func (l *Library) Load() (err error) {
	h.Debug("Loading database...")
	l.Books, _, err = l.loadBooks()
	return err
}

// Save current DB
func (l *Library) Save() (hasSaved bool, err error) {
	h.Debug("Determining if database should be saved...")
	jsonEpub, err := json.Marshal(l.Books)
	if err != nil {
		fmt.Println(err)
		return
	}

	// compare with input
	oldBooks, jsonEpubOld, err := l.loadBooks()
	if err != nil && !os.IsNotExist(err) {
		h.Error("Error loading database")
	}

	if !bytes.Equal(jsonEpub, jsonEpubOld) {
		h.Debug("Changes detected, saving database...")
		// writing db
		err = ioutil.WriteFile(l.DatabaseFile, jsonEpub, 0777)
		if err != nil {
			return
		}
		hasSaved = true

		// index what is needed.
		// diff to check the changes
		n, m, d := l.Books.Diff(oldBooks)

		// update the index
		err = l.Index.Update(bookToGeneric(n), bookToGeneric(m), bookToGeneric(d))
		if err != nil {
			h.Error("Error updating index, it may be necessary to build it anew")
			defer h.TimeTrack(time.Now(), "Indexing")
			f := func() error {
				// convert Books to []GenericBook
				allBooks := []e.GenericBook{}
				for _, b := range l.Books {
					allBooks = append(allBooks, &b)
				}
				return l.Index.Rebuild(allBooks)
			}
			if err := h.SpinWhileThingsHappen("Indexing", f); err != nil {
				return hasSaved, err
			}
			// index is now correct
			return hasSaved, nil
		}
	}
	return hasSaved, nil
}

func bookToGeneric(x map[string]b.Book) (y map[string]e.GenericBook) {
	y = make(map[string]e.GenericBook)
	for k, v := range x {
		y[k] = &v
	}
	return
}

// Backup current database.
func (l *Library) backup() (err error) {
	h.Debug("Backup up database...")
	// generate archive filename with date.
	archiveName, err := cfg.GetArchiveUniqueName(l.DatabaseFile)
	if err != nil {
		return
	}
	// creating tarball
	tar := new(archivex.TarFile)
	err = tar.Create(archiveName)
	if err != nil {
		return
	}
	err = tar.AddFile(l.DatabaseFile)
	if err != nil {
		return
	}
	tar.Close()
	return
}

// Check all Books
func (l *Library) Check() error {
	defer h.TimeTrack(time.Now(), "Checking")
	for i := range l.Books {
		h.Debug("Checking " + l.Books[i].ShortString())
		retailChanged, nonRetailChanged, err := l.Books[i].Check()
		if err != nil {
			return err
		}
		if retailChanged {
			err = errors.New("Retail epub has changed for book: " + l.Books[i].ShortString())
			return err
		}
		if nonRetailChanged {
			h.Warning("Non-retail epub for book " + l.Books[i].ShortString() + " has changed, check if this is normal.")
		}
	}
	return nil
}

// RemoveByID a book from the db
func (l *Library) RemoveByID(id int) (err error) {
	var found bool
	removeIndex := -1
	for i := range l.Books {
		if l.Books[i].ID == id {
			found = true
			removeIndex = i
			break
		}
	}
	if found {
		h.Info("REMOVING from db " + l.Books[removeIndex].ShortString())
		l.Books = append((l.Books)[:removeIndex], (l.Books)[removeIndex+1:]...)
	} else {
		err = errors.New("Did not find book with ID " + strconv.Itoa(id))
	}
	return
}

// ListNonRetailOnly among known epubs.
func (l *Library) ListNonRetailOnly() b.Books {
	return l.Books.FilterNonRetailOnly()
}

// ListRetail among known epubs.
func (l *Library) ListRetail() b.Books {
	return l.Books.FilterRetail()
}

// ListAuthors among known epubs.
func (l *Library) ListAuthors() (authors map[string]int) {
	authors = make(map[string]int)
	for _, book := range l.Books {
		author := book.Metadata.Author()
		authors[author]++
	}
	return
}

// ListPublishers among known epubs.
func (l *Library) ListPublishers() (publishers map[string]int) {
	publishers = make(map[string]int)
	for _, book := range l.Books {
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
func (l *Library) ListTags() (tags map[string]int) {
	tags = make(map[string]int)
	for _, book := range l.Books {
		for _, tag := range book.Metadata.Tags {
			tags[tag.Name]++
		}
	}
	return
}

// ListSeries associated with known epubs.
func (l *Library) ListSeries() (series map[string]int) {
	series = make(map[string]int)
	for _, book := range l.Books {
		for _, s := range book.Metadata.Series {
			series[s.Name]++
		}
	}
	return
}

// ListUntagged among known epubs.
func (l *Library) ListUntagged() b.Books {
	return l.Books.FilterUntagged()
}

// ListByProgress returns a slice of Books with the given reading progress.
func (l *Library) ListByProgress(progress string) b.Books {
	return l.Books.FilterByProgress(progress)
}

// ListIncomplete among known epubs.
func (l *Library) ListIncomplete() b.Books {
	return l.Books.FilterIncomplete()
}
