/*
Package library is a subpackage of Endive.

Library tracks and manipulates all the Books known to Endive.
It can:
	- import books (retail and non-retail)
	- build a database of said books and their metadata
	- search this database
	- organize the books according to the configuration file

*/
package library

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"os"

	b "github.com/barsanuphe/endive/book"
	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
)

// Library manages Epubs
type Library struct {
	Config      cfg.Config
	KnownHashes cfg.KnownHashes
	DB
}

// OpenLibrary constucts a valid new Library
func OpenLibrary() (l *Library, err error) {
	// config
	configPath, err := cfg.GetConfigPath()
	if err != nil {
		return
	}
	c := cfg.Config{Filename: configPath}
	// config load
	err = c.Load()
	if err != nil {
		return
	}
	// check config
	err = c.Check()
	if err != nil {
		return
	}

	// check lock
	err = cfg.SetLock()
	if err != nil {
		return
	}

	// known hashes
	hashesPath, err := cfg.GetKnownHashesPath()
	if err != nil {
		return
	}
	// load known hashes file
	h := cfg.KnownHashes{Filename: hashesPath}
	err = h.Load()
	if err != nil {
		return
	}

	l = &Library{Config: c, KnownHashes: h}
	l.DatabaseFile = c.DatabaseFile
	err = l.Load()
	if err != nil {
		return
	}
	// make each Book aware of current Config file
	for i := range l.Books {
		l.Books[i].Config = l.Config
		l.Books[i].NonRetailEpub.Config = l.Config
		l.Books[i].RetailEpub.Config = l.Config
	}
	return l, err
}

// Close the library
func (l *Library) Close() (err error) {
	_, err = l.Save()
	if err != nil {
		return
	}
	// write dirty index marker
	if l.indexNeedsRebuilding {
		err = cfg.SetDirtyIndexMarker()
		if err != nil {
			return
		}
	}
	// remove lock
	return cfg.RemoveLock()
}

// ImportRetail imports epubs from the Retail source.
func (l *Library) ImportRetail() (err error) {
	h.Title("Importing retail epubs...")
	defer h.TimeTrack(time.Now(), "Imported")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range l.Config.RetailSource {
		h.Subtitle("Searching for retail epubs in " + source)
		epubs, hashes, err := h.ListEpubsInDirectory(source)
		if err != nil {
			return err
		}
		allEpubs = append(allEpubs, epubs...)
		allHashes = append(allHashes, hashes...)
	}
	return l.ImportEpubs(allEpubs, allHashes, true)
}

// ImportNonRetail imports epubs from the Non-Retail source.
func (l *Library) ImportNonRetail() (err error) {
	h.Title("Importing non-retail epubs...")
	defer h.TimeTrack(time.Now(), "Imported")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range l.Config.NonRetailSource {
		h.Subtitle("Searching for non-retail epubs in " + source)
		epubs, hashes, err := h.ListEpubsInDirectory(source)
		if err != nil {
			return err
		}
		allEpubs = append(allEpubs, epubs...)
		allHashes = append(allHashes, hashes...)
	}
	return l.ImportEpubs(allEpubs, allHashes, false)
}

// ImportEpubs files that are retail, or not.
func (l *Library) ImportEpubs(allEpubs []string, allHashes []string, isRetail bool) (err error) {
	// force reload if it has changed
	err = l.KnownHashes.Load()
	if err != nil {
		return
	}
	defer l.KnownHashes.Save()

	newEpubs := 0
	// importing what is necessary
	for i, path := range allEpubs {
		hash := allHashes[i]
		importConfirmed := false
		e := b.Epub{Filename: path}

		// compare with known hashes
		info := b.Metadata{}
		if !l.KnownHashes.IsIn(hash) {
			// get Metadata from new epub
			info, err = e.ReadMetadata()
			if err != nil {
				h.Error("Could not analyze and import " + path)
				continue
			}
			// ask if user really wants to import it
			importConfirmed = h.YesOrNo(fmt.Sprintf("Found %s (%s).\nImport", filepath.Base(path), info.String()))
		} else {
			_, err := l.FindByHash(hash)
			if err != nil {
				// get Metadata from new epub
				info, err = e.ReadMetadata()
				if err != nil {
					h.Error("Could not analyze and import " + path)
					continue
				}
				//confirm force import
				importConfirmed = h.YesOrNo(fmt.Sprintf("File %s has already been imported but is not in the current library. Confirm importing again?", filepath.Base(path)))
			}
		}

		if importConfirmed {
			// loop over Books to find similar Metadata
			var imported bool
			knownBook, err := l.FindByMetadata(info)
			if err != nil {
				// new Book
				h.Debug("Creating new book.")
				bk := b.NewBookWithMetadata(l.generateID(), path, l.Config, isRetail, info)
				imported, err = bk.Import(path, isRetail, hash)
				if err != nil {
					return err
				}
				l.Books = append(l.Books, *bk)
			} else {
				// add to existing book
				h.Debug("Adding epub to " + knownBook.ShortString())
				imported, err = knownBook.AddEpub(path, isRetail, hash)
				if err != nil {
					return err
				}
			}

			if imported {
				// add hash to known hashes
				// NOTE: otherwise it'll pop up every other time
				added, err := l.KnownHashes.Add(hash)
				if !added || err != nil {
					return err
				}
				// saving now == saving import progress, in case of interruption
				_, err = l.KnownHashes.Save()
				if err != nil {
					return err
				}
				// saving database also
				_, err = l.Save()
				if err != nil {
					return err
				}
				newEpubs++
			}
		} else {
			h.Debug("Ignoring already imported epub " + filepath.Base(path))
		}
	}
	if isRetail {
		h.Debugf("Imported %d retail epubs.\n", newEpubs)
	} else {
		h.Debugf("Imported %d non-retail epubs.\n", newEpubs)
	}
	return
}

// Refresh current DB
func (l *Library) Refresh() (renamed int, err error) {
	h.Info("Refreshing database...")

	// scan for new epubs
	allEpubs, allHashes, err := h.ListEpubsInDirectory(l.Config.LibraryRoot)
	if err != nil {
		return
	}
	// compare allEpubs with l.Epubs
	newEpubs := []string{}
	newHashes := []string{}
	for i, epub := range allEpubs {
		_, err = l.FindByFilename(epub)
		// no error == found Epub
		if err != nil {
			// check if hash is known
			book, err := l.FindByHash(allHashes[i])
			if err != nil {
				// else, it's a new epub, import
				h.Info("NEW EPUB " + epub + " , will be imported as non-retail.")
				newEpubs = append(newEpubs, epub)
				newHashes = append(newHashes, allHashes[i])
			} else {
				// if it is, rename found file to filename in DB
				destination := book.RetailEpub.FullPath()
				if book.NonRetailEpub.Hash == allHashes[i] {
					destination = book.NonRetailEpub.FullPath()
				}
				// check if retail epub already exists
				_, err := h.FileExists(destination)
				if err == nil {
					// file already exists
					h.Errorf("Found epub %s with the same hash as %s, ignoring.", epub, destination)
				} else {
					h.Warning("Found epub %s which is called %s in the database, renaming.", epub, destination)
					// rename found file to retail name in db
					err = os.Rename(epub, destination)
					if err != nil {
						return 0, err
					}
				}
			}
		}
	}
	// import new books as non-retail
	err = l.ImportEpubs(newEpubs, newHashes, false)
	if err != nil {
		return
	}
	// refresh all books
	deletedBooks := []int{}
	for i := range l.Books {
		wasRenamed, _, err := l.Books[i].Refresh()
		if err != nil {
			return renamed, err
		}
		if !l.Books[i].HasEpub() {
			// mark for deletion
			deletedBooks = append(deletedBooks, l.Books[i].ID)
		}
		if wasRenamed[0] {
			renamed++
		}
		if wasRenamed[1] {
			renamed++
		}
	}

	// remove empty books
	for _, id := range deletedBooks {
		err := l.RemoveByID(id)
		if err != nil {
			return renamed, err
		}
	}
	// remove all empty dirs
	err = h.DeleteEmptyFolders(l.Config.LibraryRoot)
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(books []b.Book) (err error) {
	if !h.DirectoryExists(l.Config.EReaderMountPoint) {
		return errors.New("E-Reader mount point does not exist: " + l.Config.EReaderMountPoint)
	}
	h.Title("Exporting books.")
	if len(books) != 0 {
		for _, book := range books {
			destination := filepath.Join(l.Config.EReaderMountPoint, book.MainEpub().Filename)
			if !h.DirectoryExists(filepath.Dir(destination)) {
				err = os.MkdirAll(filepath.Dir(destination), 0777)
				if err != nil {
					return err
				}
			}
			if _, exists := h.FileExists(destination); exists != nil {
				h.Info(" - Exporting " + book.ShortString())
				err = h.CopyFile(book.MainEpub().FullPath(), destination)
				if err != nil {
					return err
				}
			}
		}
	}
	return
}

// DuplicateRetailEpub copies a retail epub to make a non-retail version.
func (l *Library) DuplicateRetailEpub(id int) (nonRetailEpub *b.Book, err error) {
	// TODO tests
	// find book from ID
	book, err := l.FindByID(id)
	if err != nil {
		return &b.Book{}, err
	}
	if !book.HasRetail() {
		return &b.Book{}, errors.New("Book has no retail epub")
	}
	if book.HasNonRetail() {
		return &b.Book{}, errors.New("Book already has non-retail epub")
	}
	// copy file
	targetFilename := filepath.Join(filepath.Dir(book.RetailEpub.FullPath()), "copy.epub")
	err = h.CopyFile(book.RetailEpub.FullPath(), targetFilename)
	if err != nil {
		return &b.Book{}, err
	}
	// create new Epub and refresh to get correct name
	book.NonRetailEpub = b.Epub{Filename: targetFilename, NeedsReplacement: "false", Config: l.Config, Hash: book.RetailEpub.Hash}
	_, _, err = book.RefreshEpub(book.NonRetailEpub, false)
	if err != nil {
		return book, err
	}
	return
}

// RebuildIndexBeforeSearch if index is dirty.
func (l *Library) RebuildIndexBeforeSearch() (err error) {
	return h.SpinWhileThingsHappen("Indexing", l.rebuildIndex)
}

// Search and print the results
func (l *Library) Search(query, sortBy string, limitFirst, limitLast bool, limitNumber int) (results string, err error) {
	if cfg.IsIndexDirty() {
		if err := l.RebuildIndexBeforeSearch(); err != nil {
			return "", err
		}
		// remove marker after successful re-indexing
		if err := cfg.RemoveDirtyIndexMarker(); err != nil {
			return "", err
		}
	}

	books, err := l.RunQuery(query)
	if err != nil {
		return
	}

	if len(books) != 0 {
		b.SortBooks(books, sortBy)
		if limitFirst && len(books) > limitNumber {
			books = books[:limitNumber]
		}
		if limitLast && len(books) > limitNumber {
			books = books[len(books)-limitNumber:]
		}
		return l.TabulateList(books), err
	}
	return "Nothing.", err
}

// TabulateList of books
func (l *Library) TabulateList(books []b.Book) (table string) {
	if len(books) == 0 {
		return
	}
	var rows [][]string
	for _, res := range books {
		relativePath, err := filepath.Rel(l.Config.LibraryRoot, res.FullPath())
		if err != nil {
			panic(errors.New("File " + res.FullPath() + " not in library?"))
		}
		rows = append(rows, []string{strconv.Itoa(res.ID), res.Metadata.Author(), res.Metadata.Title(), res.Metadata.Year, relativePath})
	}
	return h.TabulateRows(rows, "ID", "Author", "Title", "Year", "Filename")
}

// ShowInfo returns a table with relevant information about a book.
func (l *Library) ShowInfo() (desc string) {
	var rows [][]string
	rows = append(rows, []string{"Number of books", fmt.Sprintf("%d", len(l.Books))})
	bks := l.ListRetail()
	rows = append(rows, []string{"Number of books with a retail version", fmt.Sprintf("%d", len(bks))})
	infoMap := l.ListAuthors()
	rows = append(rows, []string{"Number of authors", fmt.Sprintf("%d", len(infoMap))})
	infoMap = l.ListTags()
	rows = append(rows, []string{"Number of tags", fmt.Sprintf("%d", len(infoMap))})
	infoMap = l.ListSeries()
	rows = append(rows, []string{"Number of series", fmt.Sprintf("%d", len(infoMap))})
	bks = l.ListUntagged()
	rows = append(rows, []string{"Number of untagged books", fmt.Sprintf("%d", len(bks))})
	bks = l.ListByProgress("read")
	rows = append(rows, []string{"Number of read books", fmt.Sprintf("%d", len(bks))})
	bks = l.ListByProgress("reading")
	rows = append(rows, []string{"Number of books currently being read", fmt.Sprintf("%d", len(bks))})
	bks = l.ListByProgress("shortlisted")
	rows = append(rows, []string{"Number of books shortlisted for imminent reading", fmt.Sprintf("%d", len(bks))})
	bks = l.ListByProgress("unread")
	rows = append(rows, []string{"Number of unread books", fmt.Sprintf("%d", len(bks))})
	return h.TabulateRows(rows, "Library", l.Config.LibraryRoot)
}
