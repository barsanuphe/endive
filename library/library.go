package library

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	b "github.com/barsanuphe/endive/book"
	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
	"github.com/bndr/gotabulate"
)

// Library manages Epubs
type Library struct {
	ConfigurationFile cfg.Config
	KnownHashesFile   cfg.KnownHashes
	DB                // anonymous, Library still has Epubs
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

	l = Library{ConfigurationFile: c, KnownHashesFile: h}
	l.DatabaseFile = c.DatabaseFile
	err = l.Load()
	if err != nil {
		return
	}
	// make each Book aware of current Config file
	for i := range l.Books {
		l.Books[i].Config = l.ConfigurationFile
		l.Books[i].NonRetailEpub.Config = l.ConfigurationFile
		l.Books[i].RetailEpub.Config = l.ConfigurationFile
	}
	return l, err
}

// ImportRetail imports epubs from the Retail source.
func (l *Library) ImportRetail() (err error) {
	fmt.Println("Library: Importing retail epubs...")
	defer h.TimeTrack(time.Now(), "Imported")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range l.ConfigurationFile.RetailSource {
		fmt.Println("Searching for retail epubs in " + source)
		epubs, hashes, err := h.ListEpubsInDirectory(source)
		if err != nil {
			return err
		}
		allEpubs = append(allEpubs, epubs...)
		allHashes = append(allHashes, hashes...)
	}
	return l.importEpubs(allEpubs, allHashes, true)
}

// ImportNonRetail imports epubs from the Non-Retail source.
func (l *Library) ImportNonRetail() (err error) {
	fmt.Println("Library: Importing non-retail epubs...")
	defer h.TimeTrack(time.Now(), "Imported")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range l.ConfigurationFile.RetailSource {
		fmt.Println("Searching for non-retail epubs in " + source)
		epubs, hashes, err := h.ListEpubsInDirectory(source)
		if err != nil {
			return err
		}
		allEpubs = append(allEpubs, epubs...)
		allHashes = append(allHashes, hashes...)
	}
	return l.importEpubs(allEpubs, allHashes, false)
}

// importEpubs, retail or not.
func (l *Library) importEpubs(allEpubs []string, allHashes []string, isRetail bool) (err error) {
	// force reload if it has changed
	err = l.KnownHashesFile.Load()
	if err != nil {
		return
	}
	defer l.KnownHashesFile.Save()

	newEpubs := 0
	// importing what is necessary
	for i, path := range allEpubs {
		hash := allHashes[i]
		// compare with known hashes
		if !l.KnownHashesFile.IsIn(hash) {
			// get Metadata from new epub
			m := b.NewMetadata()
			err = m.Read(path)
			if err != nil {
				return
			}

			// loop over Books to find similar Metadata
			var found, imported bool
			var knownBook *b.Book
			for i, book := range l.Books {
				if book.Metadata.IsSimilar(m) {
					found = true
					knownBook = &l.Books[i]
					break
				}
			}
			if !found {
				// new Book
				bk := b.NewBookWithMetadata(l.generateID(), path, l.ConfigurationFile, isRetail, m)
				imported, err = bk.Import(path, isRetail, hash)
				if err != nil {
					return
				}
				l.Books = append(l.Books, *bk)
			} else {
				// add to existing book
				fmt.Println("Adding epub to " + knownBook.ShortString())
				imported, err = knownBook.AddEpub(path, isRetail, hash)
				if err != nil {
					return
				}
			}

			if imported {
				// add hash to known hashes
				// TODO otherwise it'll pop up every other time
				added, err := l.KnownHashesFile.Add(hash)
				if !added || err != nil {
					return err
				}
				newEpubs++
			}
		} else {
			fmt.Println("Ignoring already imported epub " + path)
		}
	}
	if isRetail {
		fmt.Printf("Found %d retail epubs.\n", newEpubs)
	} else {
		fmt.Printf("Found %d non-retail epubs.\n", newEpubs)
	}
	return
}

// Refresh current DB
func (l *Library) Refresh() (renamed int, err error) {
	fmt.Println("Refreshing database...")

	// scan for new epubs
	allEpubs, allHashes, err := h.ListEpubsInDirectory(l.ConfigurationFile.LibraryRoot)
	if err != nil {
		return
	}

	// compare allEpubs with l.Epubs
	newEpubs := []string{}
	newHashes := []string{}
	for i, epub := range allEpubs {
		_, err = l.FindByFilename(epub)
		if err != nil { // no error == found Epub
			fmt.Println("NEW EPUB " + epub + " , will be imported as non-retail.")
			newEpubs = append(newEpubs, epub)
			newHashes = append(newHashes, allHashes[i])
		}
	}
	// import as non-retail
	err = l.importEpubs(allEpubs, allHashes, false)
	if err != nil {
		return
	}

	for _, book := range l.Books {
		wasRenamed, _, err := book.Refresh()
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

	// remove all empty dirs
	err = h.DeleteEmptyFolders(l.ConfigurationFile.LibraryRoot)
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(epubs []b.Book) (err error) {
	return
}

// DuplicateRetailEpub copies a retail epub to make a non-retail version.
func (l *Library) DuplicateRetailEpub(epub b.Book) (nonRetailEpub b.Book, err error) {
	// TODO find epub
	// TODO copy file
	return
}

// RunQuery and print the results
func (l *Library) RunQuery(query string) (results string, err error) {
	fmt.Println("Running query...")

	// remplace fields for simpler queries
	r := strings.NewReplacer(
		"author:", "metadata.fields.creator:",
		"title:", "metadata.fields.title:",
		"year:", "metadata.fields.year:",
		"language:", "metadata.fields.language:",
	)
	query = r.Replace(query)

	hits, err := l.Search(query)
	if err != nil {
		return
	}

	if len(hits) != 0 {
		var rows [][]string
		for _, res := range hits {
			rows = append(rows, []string{strconv.Itoa(res.ID), res.Metadata.Get("creator")[0], res.Metadata.Get("title")[0], res.Metadata.Get("year")[0], res.GetMainFilename()})
		}
		tabulate := gotabulate.Create(rows)
		tabulate.SetHeaders([]string{"ID", "Author", "Title", "Year", "Filename"})
		tabulate.SetEmptyString("N/A")
		//tabulate.SetMaxCellSize(64)
		//tabulate.SetWrapStrings(true)
		return tabulate.Render("simple"), err
	}
	return "Nothing.", err
}
