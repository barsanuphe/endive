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
	"strings"

	e "github.com/barsanuphe/endive/endive"
)

// Library manages Epubs
type Library struct {
	Config     e.Config
	Collection e.Collection
	Index      e.Indexer
	UI         e.UserInterface
	DB         e.Database
}

// Close the library
func (l *Library) Close() error {
	hasSaved, err := l.Save()
	if err != nil {
		l.UI.Error(err.Error())
		return err
	}
	if hasSaved {
		// db has been modified at some point, backup.
		if err := l.backup(); err != nil {
			l.UI.Error(err.Error())
		}
	}
	return nil
}

// GenerateID for a new Book
func (l *Library) GenerateID() (id int) {
	// id 1 for first Book
	if len(l.Collection.Books()) == 0 {
		return 1
	}
	// find max ID of ldb.Books and add 1
	for _, book := range l.Collection.Books() {
		if book.ID() > id {
			id = book.ID()
		}
	}
	id++
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(books e.Collection) (err error) {
	if !e.DirectoryExists(l.Config.EReaderMountPoint) {
		return errors.New("E-Reader mount point does not exist: " + l.Config.EReaderMountPoint)
	}
	if len(books.Books()) != 0 {
		l.UI.Title("Exporting books.")
		for _, book := range books.Books() {
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
				err = e.CopyFile(book.FullPath(), destination)
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

// Search and print the results
func (l *Library) Search(query, sortBy string, limitFirst, limitLast bool, limitNumber int) (results e.Collection, err error) {
	if err != nil {
		return
	}
	query = l.prepareQuery(query)

	booksPaths, err := l.Index.Query(query)
	if err != nil {
		// TODO const error in endive package
		if err.Error() == "Index is empty" {
			// rebuild index
			if err := l.RebuildIndex(); err != nil {
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
		var books e.Collection
		for _, path := range booksPaths {
			book, err := l.Collection.FindByFullPath(path)
			if err != nil {
				l.UI.Warning("Could not find Book: " + path)
			} else {
				books.Add(book)
			}
		}

		books.Sort(sortBy)
		if limitFirst {
			books = books.First(limitNumber)
		}
		if limitLast {
			books = books.Last(limitNumber)
		}
		return books, err
	}
	return nil, err
}

// SearchAndPrint results to a query
func (l *Library) SearchAndPrint(query, sortBy string, limitFirst, limitLast bool, limitNumber int) (results string, err error) {
	books, err := l.Search(query, sortBy, limitFirst, limitLast, limitNumber)
	return books.Table(), err
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

// ShowInfo returns a table with relevant information about a book.
func (l *Library) ShowInfo() string {
	var rows [][]string
	rows = append(rows, []string{"Number of books", fmt.Sprintf("%d", len(l.Collection.Books()))})
	bks := l.Collection.Retail().Books()
	rows = append(rows, []string{"Number of books with a retail version", fmt.Sprintf("%d", len(bks))})
	infoMap := l.Collection.Authors()
	rows = append(rows, []string{"Number of authors", fmt.Sprintf("%d", len(infoMap))})
	infoMap = l.Collection.Tags()
	rows = append(rows, []string{"Number of tags", fmt.Sprintf("%d", len(infoMap))})
	infoMap = l.Collection.Series()
	rows = append(rows, []string{"Number of series", fmt.Sprintf("%d", len(infoMap))})
	bks = l.Collection.Untagged().Books()
	rows = append(rows, []string{"Number of untagged books", fmt.Sprintf("%d", len(bks))})
	bks = l.Collection.Progress("read").Books()
	rows = append(rows, []string{"Number of read books", fmt.Sprintf("%d", len(bks))})
	bks = l.Collection.Progress("reading").Books()
	rows = append(rows, []string{"Number of books currently being read", fmt.Sprintf("%d", len(bks))})
	bks = l.Collection.Progress("shortlisted").Books()
	rows = append(rows, []string{"Number of books shortlisted for imminent reading", fmt.Sprintf("%d", len(bks))})
	bks = l.Collection.Progress("unread").Books()
	rows = append(rows, []string{"Number of unread books", fmt.Sprintf("%d", len(bks))})
	return e.TabulateRows(rows, "Library", l.Config.LibraryRoot)
}
