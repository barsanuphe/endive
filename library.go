package main

import (
	"fmt"
	"time"

	"github.com/bndr/gotabulate"
)

// Library manages Epubs
type Library struct {
	ConfigurationFile Config
	KnownHashesFile   KnownHashes
	LibraryDB         // anonymous, Library still has Epubs
}

// OpenLibrary constucts a valid new Library
func OpenLibrary() (l Library, err error) {
	// config
	configPath, err := getConfigPath()
	if err != nil {
		return
	}
	c := Config{Filename: configPath}
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
	hashesPath, err := getKnownHashesPath()
	if err != nil {
		return
	}
	// load known hashes file
	h := KnownHashes{Filename: hashesPath}
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
	return l, err
}

// ImportRetail imports epubs from the Retail source.
func (l *Library) ImportRetail() (err error) {
	fmt.Println("Library: Importing retail epubs...")
	defer timeTrack(time.Now(), "Imported")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range l.ConfigurationFile.RetailSource {
		epubs, hashes, err := listEpubsInDirectory(source)
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
	defer timeTrack(time.Now(), "Imported")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range l.ConfigurationFile.RetailSource {
		epubs, hashes, err := listEpubsInDirectory(source)
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
			m := NewMetadata()
			err = m.Read(path)
			if err != nil {
				return
			}
			// loop over Books to find similar Metadata
			var found, imported bool
			var knownBook *Book
			for _, book := range l.Books {
				if book.Metadata.IsSimilar(m) {
					found = true
					knownBook = &book
					break
				}
			}
			if !found {
				// new Book
				b := NewBookWithMetadata(path, l.ConfigurationFile, isRetail, m)
				imported, err = b.Import(path, isRetail, hash)
				if err != nil {
					return
				}
				l.Books = append(l.Books, *b)
			} else {
				// add to existing book
				imported, err = knownBook.Import(path, isRetail, hash)
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
	allEpubs, allHashes, err := listEpubsInDirectory(l.ConfigurationFile.LibraryRoot)
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
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(epubs []Book) (err error) {
	return
}

// DuplicateRetailEpub copies a retail epub to make a non-retail version.
func (l *Library) DuplicateRetailEpub(epub Book) (nonRetailEpub Book, err error) {
	// TODO find epub
	// TODO copy file
	return
}

// RunQuery and print the results
func (l *Library) RunQuery(query string) (results string, err error) {
	fmt.Println("Running query...")
	// TODO check query?

	hits, err := l.Search(query)
	if err != nil {
		return
	}

	if len(hits) != 0 {
		var rows [][]string
		for _, res := range hits {
			rows = append(rows, []string{res.Metadata.Get("author")[0], res.Metadata.Get("title")[0], res.Metadata.Get("year")[0], res.getMainFilename()})
		}
		tabulate := gotabulate.Create(rows)
		tabulate.SetHeaders([]string{"Author", "Title", "Year", "Filename"})
		tabulate.SetEmptyString("N/A")
		return tabulate.Render("simple"), err
	}
	return "Nothing.", err
}
