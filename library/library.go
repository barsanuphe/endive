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
	"strings"
	"time"

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
func OpenLibrary() (l Library, err error) {
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

	l = Library{Config: c, KnownHashes: h}
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

// ImportRetail imports epubs from the Retail source.
func (l *Library) ImportRetail() (err error) {
	h.Logger.Debug("Importing retail epubs...")
	defer h.TimeTrack(time.Now(), "Imported")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range l.Config.RetailSource {
		fmt.Println("Searching for retail epubs in " + source)
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
	h.Logger.Debug("Importing non-retail epubs...")
	defer h.TimeTrack(time.Now(), "Imported")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range l.Config.NonRetailSource {
		fmt.Println("Searching for non-retail epubs in " + source)
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
		// compare with known hashes
		if !l.KnownHashes.IsIn(hash) {
			// get Metadata from new epub
			e := b.Epub{Filename: path}
			info, err := e.ReadMetadata()
			if err != nil {
				return err
			}

			// ask if user really wants to import it
			confirmImport := h.YesOrNo(fmt.Sprintf("Found %s (%s).\nImport", filepath.Base(path), info.String()))
			if !confirmImport {
				continue
			}

			// loop over Books to find similar Metadata
			var imported bool
			knownBook, err := l.FindByMetadata(info)
			if err != nil {
				// new Book
				h.Logger.Debugf("Creating new book.")
				bk := b.NewBookWithMetadata(l.generateID(), path, l.Config, isRetail, info)
				imported, err = bk.Import(path, isRetail, hash)
				if err != nil {
					return err
				}
				l.Books = append(l.Books, *bk)
			} else {
				// add to existing book
				h.Logger.Debugf("Adding epub to " + knownBook.ShortString())
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
				newEpubs++
			}
		} else {
			h.Logger.Infof("Ignoring already imported epub " + filepath.Base(path))
		}
	}
	if isRetail {
		h.Logger.Infof("Imported %d retail epubs.\n", newEpubs)
	} else {
		h.Logger.Infof("Imported %d non-retail epubs.\n", newEpubs)
	}
	return
}

// Refresh current DB
func (l *Library) Refresh() (renamed int, err error) {
	h.Logger.Info("Refreshing database...")

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
		if err != nil { // no error == found Epub
			h.Logger.Info("NEW EPUB " + epub + " , will be imported as non-retail.")
			newEpubs = append(newEpubs, epub)
			newHashes = append(newHashes, allHashes[i])
		}
	}
	// import as non-retail
	err = l.ImportEpubs(allEpubs, allHashes, false)
	if err != nil {
		return
	}

	for i := range l.Books {
		wasRenamed, _, err := l.Books[i].Refresh()
		if err != nil {
			return renamed, err
		}
		if wasRenamed[0] {
			renamed++
		}
		if wasRenamed[1] {
			renamed++
		}
	}

	// refreshing index
	_, err = l.Index()
	if err != nil {
		return
	}

	// remove all empty dirs
	err = h.DeleteEmptyFolders(l.Config.LibraryRoot)
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(epubs []b.Book) (err error) {
	// TODO
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

// RunQuery and print the results
func (l *Library) RunQuery(query string) (results string, err error) {
	// remplace fields for simpler queries
	r := strings.NewReplacer(
		"author:", "metadata.authors:",
		"title:", "metadata.title:",
		"year:", "metadata.year:",
		"language:", "metadata.language:",
		"series:", "metadata.series.seriesname:",
		"tags:", "metadata.tags.name:",
	)
	query = r.Replace(query)
	hits, err := l.Search(query)
	if err != nil {
		return
	}

	if len(hits) != 0 {
		return l.TabulateList(hits), err
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
