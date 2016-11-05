package main

import (
	"errors"
	"fmt"
	"strings"

	b "github.com/barsanuphe/endive/book"
	e "github.com/barsanuphe/endive/endive"
)

const (
	exportSelection   = "Exporting selection to E-Reader..."
	exportBookError   = "Error exporting books to e-reader: %s"
	exportFilterError = "Error filtering books for export to e-reader"
)

func editMetadata(endive *Endive, books []*b.Book, args ...string) error {
	var rows [][]string
	for _, book := range books {
		endive.UI.Title("Editing metadata for " + book.String() + "\n")
		beforeEditBook := *book
		if err := book.EditField(args...); err != nil {
			endive.UI.Errorf("Error editing metadata for book ID#%d\n", book.ID())
		}
		if _, _, err := book.Refresh(); err != nil {
			endive.UI.Errorf("Error refreshing book ID#%d\n", book.ID())
			return err
		}
		diffs := beforeEditBook.OutputDiffTable(book, true)
		diffs = beforeEditBook.AddIDToDiff(diffs)
		rows = append(rows, diffs...)
	}
	fmt.Println(e.TabulateRows(rows, "ID", "Previous value", "Current value"))
	return nil
}

func refreshMetadata(endive *Endive, books []*b.Book, args ...string) error {
	for _, book := range books {
		if len(args) == 0 {
			// refresh all metadata
			if endive.UI.Accept("Confirm refreshing metadata for " + book.String()) {
				if err := book.ForceMetadataRefresh(); err != nil {
					endive.UI.Errorf("Error reinitializing metadata for book ID#%d\n", book.ID())
					return err
				}
			}
		} else if len(args) == 1 {
			// field is specified
			field := args[0]
			// check if valid field name
			_, isIn := e.StringInSlice(strings.ToLower(field), b.MetadataFieldNames)
			if !isIn {
				endive.UI.Error("Invalid metadata field " + field)
				return errors.New("Invalid metadata field")
			}
			// ask for confirmation
			if endive.UI.Accept("Confirm refreshing metadata field " + field + " for " + book.String()) {
				err := book.ForceMetadataFieldRefresh(field)
				if err != nil {
					endive.UI.Errorf("Error reinitializing metadata field "+field+" for book ID#%d", book.ID)
					endive.UI.Error(err.Error())
					return err
				}
			}
		}
		if _, _, err := book.Refresh(); err != nil {
			endive.UI.Errorf("Error refreshing book ID#%d\n", book.ID())
			return err
		}
		showInfo(endive, book)
	}
	return nil
}

func reviewBook(endive *Endive, book *b.Book, rating, review string) error {
	if err := book.Set("rating", rating); err != nil {
		endive.UI.Error("Rating must be a number between 0 and 5.")
		return err
	}
	if review != "" {
		if err := book.Set("review", review); err != nil {
			endive.UI.Error("Could not add review.")
			return err
		}
	}
	showInfo(endive, book)
	return nil
}

func setProgress(endive *Endive, books []*b.Book, progress string) error {
	for _, book := range books {
		// setting progress
		if err := book.Set("progress", progress); err != nil {
			return err
		}
		endive.UI.Title("%s set as %s.\n", book.String(), progress)
		// if first time set as read, set date too.
		if book.Progress == "read" && book.ReadDate == "" {
			book.SetReadDateToday()
			endive.UI.Title("Set read date to today.")
		}
	}
	return nil
}

func showInfo(endive *Endive, book *b.Book) {
	if book != nil {
		fmt.Println(book.ShowInfo())
	} else {
		fmt.Println(endive.Library.ShowInfo())
	}
}

func search(endive *Endive, parts []string, firstNBooks, lastNBooks int, sortBy string) {
	query := strings.Join(parts, " ")
	endive.UI.Debug("Searching for '" + query + "'...")
	var results e.Collection
	results = &b.Books{}
	hits, err := endive.Library.SearchAndPrint(query, sortBy, firstNBooks, lastNBooks, results)
	if err != nil {
		endive.UI.Error(err.Error())
		return
	}
	endive.UI.Display(hits)
}

func listImportableEpubs(endive *Endive, isRetail bool) {
	var candidates e.EpubCandidates
	var err error
	var txt string

	if isRetail {
		candidates, err = endive.analyzeSources(endive.Config.RetailSource, isRetail)
		txt = fmt.Sprintf("Found %d retail epubs to import: ", len(candidates))
	} else {
		candidates, err = endive.analyzeSources(endive.Config.NonRetailSource, isRetail)
		txt = fmt.Sprintf("Found %d non-retail epubs to import: ", len(candidates))
	}
	if err != nil {
		endive.UI.Error(err.Error())
		return
	}
	if len(candidates) != 0 {
		endive.UI.SubPart(txt)
		for _, cd := range candidates {
			fmt.Println(" - " + cd.Filename)
		}
	} else {
		endive.UI.SubPart("Nothing to import.")
	}
}

func importEpubs(endive *Endive, epubs []string, isRetail bool) {
	if len(epubs) >= 1 {
		// import valid paths
		if err := endive.ImportSpecific(isRetail, epubs...); err != nil {
			endive.UI.Error(err.Error())
			return
		}
	} else {
		var err error
		if isRetail {
			if len(endive.Config.RetailSource) == 0 {
				endive.UI.Error("No retail source found in configuration file!")
				return
			}
			err = endive.ImportRetail()
		} else {
			if len(endive.Config.NonRetailSource) == 0 {
				endive.UI.Error("No non-retail source found in configuration file!")
				return
			}
			err = endive.ImportNonRetail()
		}
		if err != nil {
			endive.UI.Error(err.Error())
		}
	}
}

func exportFilter(endive *Endive, parts []string) {
	query := strings.Join(parts, " ")
	var err error
	var books e.Collection
	books = &b.Books{}
	books, err = endive.Library.Search(query, "default", -1, -1, books)
	if err != nil {
		endive.UI.Error(exportFilterError)
		return
	}
	exportCollection(endive, books)
}

func exportCollection(endive *Endive, collection e.Collection) {
	endive.UI.Title(exportSelection)
	if err := endive.Library.ExportToEReader(collection); err != nil {
		endive.UI.Errorf(exportBookError, err.Error())
	}
}

func displayBooks(ui e.UserInterface, books e.Collection, firstNBooks, lastNBooks int, sortBy string) {
	if sortBy != "" {
		books.Sort(sortBy)
	}
	if firstNBooks != invalidLimit {
		books = books.First(firstNBooks)
	}
	if lastNBooks != invalidLimit {
		books = books.Last(lastNBooks)
	}
	ui.Display(books.Table())
}
