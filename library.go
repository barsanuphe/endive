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
	defer timeTrack(time.Now(), "Scanned")

	err = l.KnownHashesFile.Load()
	if err != nil {
		return
	}

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

	newEpubs := 0
	isRetail := true
	// importing what is necessary
	for i, path := range allEpubs {
		hash := allHashes[i]
		// compare with known hashes
		if !l.KnownHashesFile.IsIn(hash) {

			// TODO: check if duplicate!!!!
			// make Epub, get metadata
			e := Epub{Filename: path}
			e.Hash = hash
			err = e.GetMetadata()
			if err != nil {
				return
			}

			// see if we have duplicates
			if l.hasCopy(e, isRetail) {
				fmt.Println("Skipping duplicate " + e.String())
				return
			}

			// if not duplicate
			fmt.Println("Importing " + path)
			err = e.Import(true)
			if err != nil {
				return
			}

			// add hash to known hashes
			added, err := l.KnownHashesFile.Add(hash)
			if !added || err != nil {
				return err
			}

			// set to retail, get metadata, hash, refresh
			// TODO: make e.New(retail bool)
			// TODO: append to l.Epubs
			newEpubs++
		} else {
			fmt.Println("Ignoring already imported epub " + path)
		}

	}
	fmt.Printf("Found %d retail epubs.\n", newEpubs)

	_, err = l.KnownHashesFile.Save()
	return
}

// ImportNonRetail imports epubs from the Non-Retail source.
func (l *Library) ImportNonRetail() (err error) {
	fmt.Println("Library: Importing non-retail epubs...")
	defer timeTrack(time.Now(), "Scanned")

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

	newEpubs := 0
	// importing what is necessary
	for i, path := range allEpubs {
		hash := allHashes[i]
		// compare with known hashes
		if !l.KnownHashesFile.IsIn(hash) {
			/*
				// TODO: check if duplicate!!!!
				// make Epub, get metadata
				e := Epub{Filename: path}
				e.Hash = hash
				err = e.GetMetadata()
				if err != nil {
					return
				}

				// see if we have duplicates
				if l.hasCopy(e, isRetail) {
					fmt.Println("Skipping duplicate " + e.String())
					return
				}

				// if not duplicate
				fmt.Println("Importing " + path)
				err = e.Import(true)
				if err != nil {
					return
				}

				// add hash to known hashes
				added, err := l.KnownHashesFile.Add(hash)
				if !added || err != nil {
					return
				}

				// set to retail, get metadata, hash, refresh
				// TODO: make e.New(retail bool)
				// TODO: append to l.Epubs
			*/
			newEpubs++
		} else {
			fmt.Println("Ignoring already imported epub " + path)
		}

	}
	fmt.Printf("Found %d non-retail epubs.\n", newEpubs)
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(epubs []Epub) (err error) {
	return
}

// DuplicateRetailEpub copies a retail epub to make a non-retail version.
func (l *Library) DuplicateRetailEpub(epub Epub) (nonRetailEpub Epub, err error) {
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
			rows = append(rows, []string{res.Author, res.Title, res.PublicationYear, res.Filename})
		}
		tabulate := gotabulate.Create(rows)
		tabulate.SetHeaders([]string{"Author", "Title", "Year", "Filename"})
		tabulate.SetEmptyString("N/A")
		return tabulate.Render("simple"), err
	}
	return "Nothing.", err
}

// Refresh current DB
func (l *Library) Refresh() (renamed int, err error) {
	fmt.Println("Refreshing database...")

	// scan for new epubs
	allEpubs, _, err := listEpubsInDirectory(l.ConfigurationFile.LibraryRoot)
	if err != nil {
		return
	}
	// compare allEpubs with l.Epubs
	for _, epub := range allEpubs {
		_, err = l.FindByFilename(epub)
		if err != nil {  // no error == found Epub
			fmt.Println("NEW EPUB " + epub)
			// TODO: import
			// TODO: as non-retail? look for retail suffix?
		}
	}

	for _, epub := range l.Epubs {
		// TODO
		oldName := epub.Filename
		wasRenamed, _, err := epub.Refresh()
		if err != nil {
			return renamed, err
		}
		if wasRenamed {
			fmt.Println("Moved " + oldName + " to " + epub.Filename)
			renamed++
		}
	}
	return
}
