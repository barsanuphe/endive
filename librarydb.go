package main

import "fmt"

// LibraryDB manages the epub database and search
type LibraryDB struct {
	DatabaseFile string
	IndexFile    string // can be in XDG data path
	Epubs        []Epub
}

// Refresh current DB
func (lbd *LibraryDB) Refresh() (err error) {
	fmt.Println("Refreshing database...")
	// TODO loop over known Epubs, make them refresh (rename from md) and generate their json
	// TODO aggregate in a single json file, the DB
	// TODO automatically index it with bleve
	return
}

// Search current DB
func (lbd *LibraryDB) Search() (results []Epub, err error) {
	// TODO make sure the index is up to date
	// TODO run bleve query, return results
	fmt.Println("Searching database...")
	return
}

// SearchJSON current DB and output as JSON
func (lbd *LibraryDB) SearchJSON() (jsonOutput string, err error) {
	// TODO Search() then get JSON output from each result Epub
	// TODO OR --- the opposite. bleve can return JSON, Search has to parse it and locate the relevant Epub objects
	fmt.Println("Searching database with JSON output...")
	return
}

// ListNonRetailOnly among known epubs.
func (lbd *LibraryDB) ListNonRetailOnly() (nonretail []Epub, err error) {
	// TODO return Search for querying non retail epubs, removing the epubs with same title/author but retail
	return
}

// ListRetailOnly among known epubs.
func (lbd *LibraryDB) ListRetailOnly() (retail []Epub, err error) {
	return
}

// ListAuthors among known epubs.
func (lbd *LibraryDB) ListAuthors() (authors []string, err error) {
	return
}

// ListTags associated with known epubs.
func (lbd *LibraryDB) ListTags() (tags []string, err error) {
	// TODO search for tags in all epubs, remove duplicates
	return
}

// ListUntagged among known epubs.
func (lbd *LibraryDB) ListUntagged() (untagged []Epub, err error) {
	return
}

// ListWithTag among known epubs.
func (lbd *LibraryDB) ListWithTag(tag string) (tagged []Epub, err error) {
	return
}
