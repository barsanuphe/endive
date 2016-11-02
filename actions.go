package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	b "github.com/barsanuphe/endive/book"
	e "github.com/barsanuphe/endive/endive"
	l "github.com/barsanuphe/endive/library"

	"github.com/urfave/cli"
)

const (
	parsingError      = "Error parsing arguments: %s"
	exportAllBooks    = "Exporting everything to E-Reader..."
	exportSelection   = "Exporting selection to E-Reader..."
	exportBookError   = "Error exporting books to e-reader: %s"
	exportFilterError = "Error filtering books for export to e-reader"
)


func editMetadata(endive *Endive, id int, args ...string) {
	// if ID, list tags of ID
	bk, err := e.Collection.FindByID(id)
	if err != nil {
		return err
	}
	book := bk.(*b.Book)
	endive.UI.Title("Editing metadata for " + book.String() + "\n")
	if err := book.EditField(args...); err != nil {
		endive.UI.Errorf("Error editing metadata for book ID#%d\n", book.ID())
	}
	_, _, err = book.Refresh()
	if err != nil {
		endive.UI.Errorf("Error refreshing book ID#%d\n", book.ID())
		return
	}
	showInfo(endive, id)
}

func refreshMetadata(endive *Endive, id int, args ...string) {
	bk, err := e.Collection.FindByID(id)
	if err != nil {
		return err
	}
	book := bk.(*b.Book)

	switch len(args) {
	case 0:
		// refresh all metadata
		if endive.UI.Accept("Confirm refreshing metadata for " + book.String()) {
			err := book.ForceMetadataRefresh()
			if err != nil {
				endive.UI.Errorf("Error reinitializing metadata for book ID#%d\n", book.ID())
			}
		}
	case 1:
		// field is specified
		field := args[0]
		// check if valid field name
		_, isIn := e.StringInSlice(strings.ToLower(field), b.MetadataFieldNames)
		if !isIn {
			endive.UI.Error("Invalid metadata field " + field)
			return
		}
		// ask for confirmation
		if endive.UI.Accept("Confirm refreshing metadata field " + field + " for " + book.String()) {
			err := book.ForceMetadataFieldRefresh(field)
			if err != nil {
				endive.UI.Errorf("Error reinitializing metadata field "+field+" for book ID#%d", book.ID)
				endive.UI.Error(err.Error())
			}
		}
	}
	_, _, err = book.Refresh()
	if err != nil {
		endive.UI.Errorf("Error refreshing book ID#%d\n", book.ID())
		return
	}
	showInfo(endive, id)
}

func setProgress(endive *Endive, id int, progress, rating, review string) error {
	// if ID, list tags of ID
	bk, err := e.Collection.FindByID(id)
	if err != nil {
		return err
	}
	book := bk.(*b.Book)

	// setting progress
	if err := book.Set("progress", progress); err != nil {
		endive.UI.Error("Progress must be among: unread/shortlisted/reading/read")
		return
	}
	// if first time set as read, set date too.
	if book.Progress == "read" && book.ReadDate == "" {
		book.SetReadDateToday()
	}
	if rating != "" {
		if err := book.Set("rating", rating); err != nil {
			endive.UI.Error("Rating must be a number between 0 and 5.")
			return err
		}
	}
	if review != "" {
		if err := book.Set("review", review); err != nil {
			endive.UI.Error("Could not add review.")
			return err
		}
	}
	showInfo(endive, id)
	return nil
}

func showInfo(endive *Endive, id int) {
	if id != InvalidID {
		// if ID, list tags of ID
		bk, err := e.Collection.FindByID(id)
		if err != nil {
			return err
		}
		book := bk.(*b.Book)
		fmt.Println(book.ShowInfo())
	} else {
		fmt.Println(endive.Library.ShowInfo())
	}
}

func listTags(endive *Endive, id int) error {
	if id != InvalidID {
		// if ID, list tags of ID
		bk, err := e.Collection.FindByID(id)
		if err != nil {
			return err
		}
		book := bk.(*b.Book)
		var rows [][]string
		rows = append(rows, []string{book.String(), book.Metadata.Tags.String()})
		endive.UI.Display(e.TabulateRows(rows, "Book", "Tags"))
	} else {
		// list all tags
		tags := endive.Library.Collection.Tags()
		endive.UI.Display(e.TabulateMap(tags, "Tag", "# of Books"))
	}
	return nil
}

func listSeries(endive *Endive, id int) error {
	if id != InvalidID {
		// if ID, list series of ID
		bk, err := e.Collection.FindByID(id)
		if err != nil {
			return err
		}
		book := bk.(*b.Book)
		var rows [][]string
		rows = append(rows, []string{book.String(), book.Metadata.Series.String()})
		endive.UI.Display(e.TabulateRows(rows, "Book", "Series"))
	} else {
		// list all tags
		tags := endive.Library.Collection.Series()
		endive.UI.Display(e.TabulateMap(tags, "Series", "# of Books"))
	}
	return nil
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
	endive.UI.Title(exportSelection)
	query := strings.Join(parts, " ")
	var err error
	var books e.Collection
	books = &b.Books{}
	books, err = endive.Library.Search(query, "default", -1, -1, books)
	if err != nil {
		endive.UI.Error(exportFilterError)
		return
	}
	if err := endive.Library.ExportToEReader(books); err != nil {
		endive.UI.Errorf(exportBookError, err.Error())
	}
}

func exportAll(endive *Endive) {
	endive.UI.Title(exportAllBooks)
	err := endive.Library.ExportToEReader(endive.Library.Collection)
	if err != nil {
		endive.UI.Errorf(exportBookError, err.Error())
	}
}

func displayBooks(ui e.UserInterface, books e.Collection, firstNBooks, lastNBooks int, sortBy string) {
	if sortBy != "" {
		books.Sort(sortBy)
	}
	if firstNBooks != -1 {
		books = books.First(firstNBooks)
	}
	if lastNBooks != -1 {
		books = books.Last(lastNBooks)
	}
	ui.Display(books.Table())
}
