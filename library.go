package main

import "fmt"

// Library manages Epubs
type Library struct {
	ConfigurationFile Config
	LibraryDB         // anonymous, Library still has Epubs
}

// ImportRetail imports epubs from the Retail source.
func (l *Library) ImportRetail() (err error) {
	fmt.Println("Library: Importing retail epubs...")
	// TODO walk l.ConfigurationFile.RetailSource

	return
}

// ImportNonRetail imports epubs from the Non-Retail source.
func (l *Library) ImportNonRetail() (err error) {
	fmt.Println("Library: Importing non-retail epubs...")
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(epubs []Epub) (err error) {
	return
}

// DuplicateRetailEpub copies a retail epub to make a non-retail version.
func (l *Library) DuplicateRetailEpub(epub Epub) (nonRetailEpub Epub, err error) {
	return
}

// RunQuery and print the results
func (l *Library) RunQuery(query string) (results string, err error) {
	fmt.Println("Running query...")
	// TODO l.Search(query) then loop over results to generate text
	return
}
