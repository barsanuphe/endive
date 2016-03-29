package main

import "fmt"

// Library manages Epubs
type Library struct {
	Root              string
	ConfigurationFile string
	LibraryDB         // anonymous, Library still has Epubs
}

// ImportRetail imports epubs from the Retail source.
func (l *Library) ImportRetail() (err error) {
	return
}

// ImportNonRetail imports epubs from the Non-Retail source.
func (l *Library) ImportNonRetail() (err error) {
	return
}

// ExportToEReader selected epubs.
func (l *Library) ExportToEReader(epubs Epubs) (err error) {
	return
}

// DuplicateRetailEpub copies a retail epub to make a non-retail version.
func (l *Library) DuplicateRetailEpub(epub EpubRetail) (nonRetailEpub EpubNonRetail, err error) {
	return
}

// RunQuery and print the results
func (l *Library) RunQuery(query string) (results string, err error) {
	fmt.Println("Running query...")
	// TODO l.Search(query) then loop over results to generate text
	return
}
