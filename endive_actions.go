package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	b "github.com/barsanuphe/endive/book"
	h "github.com/barsanuphe/endive/helpers"
	l "github.com/barsanuphe/endive/library"

	"github.com/codegangsta/cli"
)

func checkPort(c *cli.Context) (port int, err error) {
	if len(c.Args()) < 1 {
		err = errors.New("Not enough arguments")
		return
	}
	port, err = strconv.Atoi(c.Args()[0])
	if err != nil {
		err = errors.New("Argument must be a valid port number")
		return
	}
	return
}

func checkArgsWithID(l *l.Library, args []string) (book *b.Book, other []string, err error) {
	if len(args) < 1 {
		err = errors.New("Not enough arguments")
		return
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return
	}
	// get book from ID
	book, err = l.FindByID(id)
	if err != nil {
		return
	}
	other = args[1:]
	return
}

func editMetadata(lb *l.Library, c *cli.Context) {
	book, args, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	if err := book.EditField(args...); err != nil {
		h.Errorf("Error editing metadata for book ID#%d", book.ID)
	}
	_, _, err = book.Refresh()
	if err != nil {
		h.Errorf("Error refreshing book ID#%d", book.ID)
		return
	}
	showInfo(lb, c)
}

func refreshMetadata(lb *l.Library, c *cli.Context) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	// ask confirmation
	if h.YesOrNo("Confirm refreshing metadata for " + book.ShortString()) {
		err := book.ForceMetadataRefresh()
		if err != nil {
			h.Errorf("Error reinitializing metadata for book ID#%d", book.ID)
		}
	}
	_, _, err = book.Refresh()
	if err != nil {
		h.Errorf("Error refreshing book ID#%d", book.ID)
		return
	}
	showInfo(lb, c)
}

func markRead(lb *l.Library, c *cli.Context) {
	book, args, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	if len(args) >= 1 {
		// TODO check rating format
		book.Rating = args[0]
	}
	if len(args) == 2 {
		book.Review = args[1]
	}
	book.SetReadDateToday()
	book.SetProgress("read")
}

func showInfo(lb *l.Library, c *cli.Context) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		fmt.Println("Error parsing arguments: " + err.Error())
		return
	}
	fmt.Println(book.ShowInfo())
}

func listTags(lb *l.Library, c *cli.Context) (err error) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		// list all tags
		tags := lb.ListTags()
		fmt.Println(h.TabulateMap(tags, "Tag", "# of Books"))

	} else {
		// if ID, list tags of ID
		var rows [][]string
		rows = append(rows, []string{book.ShortString(), book.Metadata.Tags.String()})
		fmt.Println(h.TabulateRows(rows, "Book", "Tags"))
	}
	return
}

func listSeries(lb *l.Library, c *cli.Context) {
	book, _, err := checkArgsWithID(lb, c.Args())
	if err != nil {
		// list all series
		series := lb.ListSeries()
		fmt.Println(h.TabulateMap(series, "Series", "# of Books"))
	} else {
		// if ID, list series of ID
		var rows [][]string
		rows = append(rows, []string{book.ShortString(), book.Metadata.Series.String()})
		fmt.Println(h.TabulateRows(rows, "Book", "Series"))
	}
	return
}

func search(lb *l.Library, c *cli.Context) {
	if c.NArg() == 0 {
		fmt.Println("No query found!")
	} else {
		query := strings.Join(c.Args(), " ")
		fmt.Println("Searching for '" + query + "'...")
		results, err := lb.Search(query)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		fmt.Println(results)
	}
}

func importEpubs(lb *l.Library, c *cli.Context, isRetail bool) {
	if len(c.Args()) >= 1 {
		// check valid path
		validPaths := []string{}
		validHashes := []string{}
		for _, path := range c.Args() {
			validPath, err := h.FileExists(path)
			if err == nil && filepath.Ext(validPath) == ".epub" {
				validPaths = append(validPaths, validPath)
				validHash, err := h.CalculateSHA256(path)
				if err != nil {
					return
				}
				validHashes = append(validHashes, validHash)
			}
		}
		// import valid paths
		err := lb.ImportEpubs(validPaths, validHashes, isRetail)
		if err != nil {
			return
		}
	} else {
		var err error
		if isRetail {
			if len(lb.Config.RetailSource) == 0 {
				h.Error("No retail source found in configuration file!")
				return
			}
			err = lb.ImportRetail()
		} else {
			if len(lb.Config.NonRetailSource) == 0 {
				h.Error("No non-retail source found in configuration file!")
				return
			}
			err = lb.ImportNonRetail()
		}
		if err != nil {
			panic(err)
		}
	}
}

func exportFilter(lb *l.Library, c *cli.Context) {
	fmt.Println("Exporting selection to E-Reader...")
	query := strings.Join(c.Args(), " ")
	books, err := lb.RunQuery(query)
	if err != nil {
		h.Error("Error filtering books for export to e-reader")
		return
	}
	err = lb.ExportToEReader(books)
	if err != nil {
		h.Errorf("Error exporting books to e-reader: %s", err.Error())
	}
}

func exportAll(lb *l.Library, c *cli.Context) {
	fmt.Println("Exporting everything to E-Reader...")
	err := lb.ExportToEReader(lb.Books)
	if err != nil {
		h.Errorf("Error exporting books to e-reader: %s", err.Error())
	}
}
