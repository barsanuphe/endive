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
	"os"
	"path/filepath"
	"strconv"
	"time"

	"strings"

	b "github.com/barsanuphe/endive/book"
	e "github.com/barsanuphe/endive/endive"
)

// Library manages Epubs
type Library struct {
	Config       e.Config
	KnownHashes  e.KnownHashes
	DatabaseFile string
	Books        b.Books
	Index        e.Indexer
	UI           e.UserInterface
}

// Close the library
func (l *Library) Close() error {
	hasSaved, err := l.Save()
	if err != nil {
		return err
	}
	if hasSaved {
		// db has been modified at some point, backup.
		if err := l.backup(); err != nil {
			l.UI.Error(err.Error())
		}
	}
	// remove lock
	return e.RemoveLock()
}

// importFromSource all detected epubs, tagging them as retail or non-retail as requested.
func (l *Library) importFromSource(sources []string, retail bool) (err error) {
	defer e.TimeTrack(l.UI, time.Now(), "Imported")
	sourceType := "retail"
	if !retail {
		sourceType = "non-retail"
	}
	l.UI.Title("Importing " + sourceType + " epubs...")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range sources {
		l.UI.SubTitle("Searching for " + sourceType + " epubs in " + source)
		epubs, hashes, err := e.ListEpubsInDirectory(source)
		if err != nil {
			return err
		}
		allEpubs = append(allEpubs, epubs...)
		allHashes = append(allHashes, hashes...)
	}
	return l.ImportEpubs(allEpubs, allHashes, retail)
}

// ImportRetail imports epubs from the Retail source.
func (l *Library) ImportRetail() (err error) {
	return l.importFromSource(l.Config.RetailSource, true)
}

// ImportNonRetail imports epubs from the Non-Retail source.
func (l *Library) ImportNonRetail() (err error) {
	return l.importFromSource(l.Config.NonRetailSource, false)
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
		ep := b.Epub{Filename: path, UI: l.UI}

		// compare with known hashes
		info := b.Metadata{}
		if !l.KnownHashes.IsIn(hash) {
			// get Metadata from new epub
			info, err = ep.ReadMetadata()
			if err != nil {
				if err.Error() == "ISBN not found in epub" {
					isbn, err := e.AskForISBN(l.UI)
					if err != nil {
						l.UI.Warning("Warning: ISBN still unknown.")
					} else {
						info.ISBN = isbn
					}
				} else {
					l.UI.Error("Could not analyze and import " + path)
					continue
				}
			}
			// ask if user really wants to import it
			importConfirmed = l.UI.YesOrNo(fmt.Sprintf("Found %s (%s).\nImport", filepath.Base(path), info.String()))
		} else {
			_, err := l.Books.FindByHash(hash)
			if err != nil {
				// get Metadata from new epub
				info, err = ep.ReadMetadata()
				if err != nil {
					if err.Error() == "ISBN not found in epub" {
						isbn, err := e.AskForISBN(l.UI)
						if err != nil {
							l.UI.Warning("Warning: ISBN still unknown.")
						} else {
							info.ISBN = isbn
						}
					} else {
						l.UI.Error("Could not analyze and import " + path)
						continue
					}
				}
				//confirm force import
				importConfirmed = l.UI.YesOrNo(fmt.Sprintf("File %s has already been imported but is not in the current library. Confirm importing again?", filepath.Base(path)))
			}
		}

		if importConfirmed {
			// loop over Books to find similar Metadata
			var imported bool
			knownBook, err := l.Books.FindByMetadata(info)
			if err != nil {
				// new Book
				l.UI.Debug("Creating new book.")
				bk := b.NewBookWithMetadata(l.UI, l.generateID(), path, l.Config, isRetail, info)
				imported, err = bk.Import(path, isRetail, hash)
				if err != nil {
					return err
				}
				l.Books = append(l.Books, *bk)
			} else {
				// add to existing book
				l.UI.Debug("Adding epub to " + knownBook.ShortString())
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
			l.UI.Debug("Ignoring already imported epub " + filepath.Base(path))
		}
	}
	if isRetail {
		l.UI.Debugf("Imported %d retail epubs.\n", newEpubs)
	} else {
		l.UI.Debugf("Imported %d non-retail epubs.\n", newEpubs)
	}
	return
}

// generateID for a new Book
func (l *Library) generateID() (id int) {
	// id 1 for first Book
	if len(l.Books) == 0 {
		return 1
	}
	// find max ID of ldb.Books and add 1
	for _, book := range l.Books {
		if book.ID > id {
			id = book.ID
		}
	}
	id++
	return
}

// Refresh current DB
func (l *Library) Refresh() (renamed int, err error) {
	l.UI.Info("Refreshing database...")

	// scan for new epubs
	allEpubs, allHashes, err := e.ListEpubsInDirectory(l.Config.LibraryRoot)
	if err != nil {
		return
	}
	// compare allEpubs with l.Epubs
	newEpubs := []string{}
	newHashes := []string{}
	for i, epub := range allEpubs {
		_, err = l.Books.FindByFullPath(epub)
		// no error == found Epub
		if err != nil {
			// check if hash is known
			book, err := l.Books.FindByHash(allHashes[i])
			if err != nil {
				// else, it's a new epub, import
				l.UI.Info("NEW EPUB " + epub + " , will be imported as non-retail.")
				newEpubs = append(newEpubs, epub)
				newHashes = append(newHashes, allHashes[i])
			} else {
				// if it is, rename found file to filename in DB
				destination := book.RetailEpub.FullPath()
				if book.NonRetailEpub.Hash == allHashes[i] {
					destination = book.NonRetailEpub.FullPath()
				}
				// check if retail epub already exists
				_, err := e.FileExists(destination)
				if err == nil {
					// file already exists
					l.UI.Errorf("Found epub %s with the same hash as %s, ignoring.", epub, destination)
				} else {
					l.UI.Warningf("Found epub %s which is called %s in the database, renaming.", epub, destination)
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
	err = e.DeleteEmptyFolders(l.Config.LibraryRoot, l.UI)
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(books []b.Book) (err error) {
	if !e.DirectoryExists(l.Config.EReaderMountPoint) {
		return errors.New("E-Reader mount point does not exist: " + l.Config.EReaderMountPoint)
	}
	if len(books) != 0 {
		l.UI.Title("Exporting books.")
		for _, book := range books {
			filename := book.CleanFilename()
			destination := filepath.Join(l.Config.EReaderMountPoint, filename)
			if !e.DirectoryExists(filepath.Dir(destination)) {
				err = os.MkdirAll(filepath.Dir(destination), 0777)
				if err != nil {
					return err
				}
			}
			if _, exists := e.FileExists(destination); exists != nil {
				l.UI.Info(" - Exporting " + book.ShortString())
				err = e.CopyFile(book.MainEpub().FullPath(), destination)
				if err != nil {
					return err
				}
			} else {
				l.UI.Info(" - Previously exported: " + book.ShortString())
			}
		}
	} else {
		l.UI.Title("Nothing to export.")
	}
	return
}

// DuplicateRetailEpub copies a retail epub to make a non-retail version.
func (l *Library) DuplicateRetailEpub(id int) (nonRetailEpub *b.Book, err error) {
	// TODO tests
	// find book from ID
	book, err := l.Books.FindByID(id)
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
	err = e.CopyFile(book.RetailEpub.FullPath(), targetFilename)
	if err != nil {
		return &b.Book{}, err
	}
	// create new Epub and refresh to get correct name
	book.NonRetailEpub = b.Epub{Filename: targetFilename, NeedsReplacement: "false", Config: l.Config, UI: l.UI, Hash: book.RetailEpub.Hash}
	_, _, err = book.RefreshEpub(book.NonRetailEpub, false)
	if err != nil {
		return book, err
	}
	return
}

// Search and print the results
func (l *Library) Search(query, sortBy string, limitFirst, limitLast bool, limitNumber int) (results b.Books, err error) {
	if err != nil {
		return
	}
	query = l.prepareQuery(query)

	booksPaths, err := l.Index.Query(query)
	if err != nil {
		// TODO const error in endive package
		if err.Error() == "Index is empty" {
			// rebuild index
			defer e.TimeTrack(l.UI, time.Now(), "Indexing")
			f := func() error {
				// convert Books to []GenericBook
				allBooks := bookSliceToGeneric(l.Books)
				return l.Index.Rebuild(allBooks)
			}
			if err := e.SpinWhileThingsHappen("Indexing", f); err != nil {
				return results, err
			}
			// trying again
			booksPaths, err = l.Index.Query(query)
			if err != nil {
				return
			}
		} else {
			return results, err
		}
	}
	if len(booksPaths) != 0 {
		// find the Book for each path
		books := b.Books{}
		for _, path := range booksPaths {
			book, err := l.Books.FindByFullPath(path)
			if err != nil {
				l.UI.Warning("Could not find Book: " + path)
			} else {
				books = append(books, *book)
			}
		}

		b.SortBooks(books, sortBy)
		if limitFirst && len(books) > limitNumber {
			books = books[:limitNumber]
		}
		if limitLast && len(books) > limitNumber {
			books = books[len(books)-limitNumber:]
		}
		return books, err
	}
	return b.Books{}, err
}

// SearchAndPrint results to a query
func (l *Library) SearchAndPrint(query, sortBy string, limitFirst, limitLast bool, limitNumber int) (results string, err error) {
	books, err := l.Search(query, sortBy, limitFirst, limitLast, limitNumber)
	return l.TabulateList(books), err
}

// prepareQuery before search
func (l *Library) prepareQuery(queryString string) string {
	// replace fields for simpler queries
	r := strings.NewReplacer(
		"author:", "metadata.authors:",
		"title:", "metadata.title:",
		"year:", "metadata.year:",
		"language:", "metadata.language:",
		"series:", "metadata.series.seriesname:",
		"tags:", "metadata.tags.name:",
		"tag:", "metadata.tags.name:",
		"publisher:", "metadata.publisher:",
		"category:", "metadata.category:",
		"genre:", "metadata.main_genre:",
		"description:", "metadata.description:",
	)
	return r.Replace(queryString)
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
		rows = append(rows, []string{strconv.Itoa(res.ID), res.Metadata.Author(), res.Metadata.Title(), res.Metadata.OriginalYear, relativePath})
	}
	return e.TabulateRows(rows, "ID", "Author", "Title", "Year", "Filename")
}

// ShowInfo returns a table with relevant information about a book.
func (l *Library) ShowInfo() string {
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
	return e.TabulateRows(rows, "Library", l.Config.LibraryRoot)
}
