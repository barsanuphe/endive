package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"encoding/json"

	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/language/en"
)

// LibraryDB manages the epub database and search
type LibraryDB struct {
	DatabaseFile string
	IndexFile    string // can be in XDG data path
	Epubs        []Epub
}

// Load current DB
func (ldb *LibraryDB) Load() (err error) {
	fmt.Println("Loading database...")
	bytes, err := ioutil.ReadFile(ldb.DatabaseFile)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &ldb.Epubs)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// Save current DB
func (ldb *LibraryDB) Save() (err error) {
	fmt.Println("Saving database...")
	jsonEpub, err := json.Marshal(ldb.Epubs)
	if err != nil {
		fmt.Println(err)
		return
	}
	// writing db
	err = ioutil.WriteFile(ldb.DatabaseFile, jsonEpub, 0777)
	if err != nil {
		return
	}
	// indexing db
	numIndexed, err := ldb.Index()
	if err != nil {
		return
	}
	// TODO see how to remove files no longer present from index
	fmt.Println("Saved and indexed " + strconv.FormatUint(numIndexed, 10) + " epubs.")
	return
}

func buildIndexMapping() (*bleve.IndexMapping, error) {
	// a generic reusable mapping for english text
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	epubMapping := bleve.NewDocumentMapping()

	epubMapping.AddFieldMappingsAt("filename", englishTextFieldMapping)
	epubMapping.AddFieldMappingsAt("description", englishTextFieldMapping)
	epubMapping.AddFieldMappingsAt("language", englishTextFieldMapping)
	epubMapping.AddFieldMappingsAt("test", englishTextFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("epub", epubMapping)

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = "en"

	return indexMapping, nil
}

func openIndex(path string) bleve.Index {
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new index...")
		// create a mapping
		indexMapping, err := buildIndexMapping()
		index, err = bleve.New(path, indexMapping)
		if err != nil {
			log.Fatal(err)
		}
	} else if err == nil {
		//log.Printf("Opening existing index...")
	} else {
		log.Fatal(err)
	}
	return index
}

// Index current DB
func (ldb *LibraryDB) Index() (numIndexed uint64, err error) {
	// open index
	index := openIndex("endive.index")
	defer index.Close()

	// read the bytes
	jsonBytes, err := ioutil.ReadFile(ldb.DatabaseFile)
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonBytes, &ldb.Epubs)
	if err != nil {
		fmt.Print("Error:", err)
	}

	// index by filename
	for _, epub := range ldb.Epubs {
		index.Index(epub.Filename, epub) // jsonBytes
	}

	// check number of indexed documents
	numIndexed, err = index.DocCount()
	if err != nil {
		return
	}

	fmt.Println("Indexed: " + strconv.FormatUint(numIndexed, 10) + " epubs.")

	return
}

// Search current DB
func (lbd *LibraryDB) Search(queryString string) (results []Epub, err error) {
	// TODO make sure the index is up to date

	fmt.Println("Searching database for " + queryString + " ...")

	// search for some text
	query := bleve.NewQueryStringQuery(queryString)
	search := bleve.NewSearchRequest(query)
	// open index
	index := openIndex("endive.index")
	defer index.Close()
	searchResults, err := index.Search(search)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(searchResults.Total)
	if searchResults.Total != 0 {
		for _, hit := range searchResults.Hits {
			fmt.Println("Found " + hit.ID)
		}
	}
	// TODO run bleve query, return Epub results
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
