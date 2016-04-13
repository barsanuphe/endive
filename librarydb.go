package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzers/keyword_analyzer"
	"github.com/blevesearch/bleve/analysis/language/en"
	"launchpad.net/go-xdg"
)

const xdgIndexPath string = endive + "/" + endive + ".index"

// getIndexPath gets the default index path
func getIndexPath() (path string) {
	path, err := xdg.Cache.Find(xdgIndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			path = filepath.Join(xdg.Cache.Dirs()[0], xdgIndexPath)
		} else {
			panic(err)
		}
	}
	return
}

// LibraryDB manages the epub database and search
type LibraryDB struct {
	DatabaseFile string
	IndexFile    string // can be in XDG data path
	Books        []Book
}

// Load current DB
func (ldb *LibraryDB) Load() (err error) {
	fmt.Println("Loading database...")
	bytes, err := ioutil.ReadFile(ldb.DatabaseFile)
	if err != nil {
		if os.IsNotExist(err) {
			// first run
			return nil
		}
		return
	}
	err = json.Unmarshal(bytes, &ldb.Books)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// Save current DB
func (ldb *LibraryDB) Save() (hasSaved bool, err error) {
	jsonEpub, err := json.Marshal(ldb.Books)
	if err != nil {
		fmt.Println(err)
		return
	}
	// compare with input
	jsonEpubOld, err := ioutil.ReadFile(ldb.DatabaseFile)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println(err)
		return
	}

	if !bytes.Equal(jsonEpub, jsonEpubOld) {
		fmt.Println("Changes detected, saving database...")
		// writing db
		err = ioutil.WriteFile(ldb.DatabaseFile, jsonEpub, 0777)
		if err != nil {
			return
		}
		hasSaved = true
		// remove old index
		err = os.RemoveAll(getIndexPath())
		if err != nil {
			fmt.Println(err)
			return hasSaved, err
		}

		// indexing db
		numIndexed, err := ldb.Index()
		if err != nil {
			return hasSaved, err
		}
		fmt.Println("Saved and indexed " + strconv.FormatUint(numIndexed, 10) + " epubs.")
	}
	return
}

// generateID for a new Book
func (ldb *LibraryDB) generateID() (id int) {
	// id 0 for first Book
	if len(ldb.Books) == 0 {
		return
	}
	// find max ID of ldb.Books
	for _, book := range ldb.Books {
		if book.ID > id {
			id = book.ID
		}
	}
	id ++
	return
}

// FindByID among known Books
func (ldb *LibraryDB) FindByID(id int) (result *Book, err error) {
	for i, bk := range ldb.Books {
		if bk.ID == id {
			return &ldb.Books[i], nil
		}
	}
	return &Book{}, errors.New("Could not find book with ID " + strconv.Itoa(id))
}

//FindByFilename among known Books
func (ldb *LibraryDB) FindByFilename(filename string) (result *Book, err error) {
	for i, bk := range ldb.Books {
		if bk.RetailEpub.Filename == filename || bk.NonRetailEpub.Filename == filename {
			return &ldb.Books[i], nil
		}
	}
	return &Book{}, errors.New("Could not find book with epub " + filename)
}

func buildIndexMapping() (*bleve.IndexMapping, error) {
	// TODO index everything

	// a generic reusable mapping for english text
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = en.AnalyzerName
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword_analyzer.Name

	epubMapping := bleve.NewDocumentMapping()

	epubMapping.AddFieldMappingsAt("progress", textFieldMapping)
	epubMapping.AddFieldMappingsAt("description", textFieldMapping)
	epubMapping.AddFieldMappingsAt("language", textFieldMapping)
	epubMapping.AddFieldMappingsAt("creator", textFieldMapping)
	epubMapping.AddFieldMappingsAt("title", textFieldMapping)
	epubMapping.AddFieldMappingsAt("year", textFieldMapping)
	epubMapping.AddFieldMappingsAt("isbn", textFieldMapping)
	epubMapping.AddFieldMappingsAt("rating", textFieldMapping)
	epubMapping.AddFieldMappingsAt("tags", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.AddDocumentMapping("epub", epubMapping)

	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = "en"

	return indexMapping, nil
}

func openIndex() (index bleve.Index, isNew bool) {
	index, err := bleve.Open(getIndexPath())
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new index...")
		// create a mapping
		indexMapping, err := buildIndexMapping()
		index, err = bleve.New(getIndexPath(), indexMapping)
		if err != nil {
			log.Fatal(err)
		}
		isNew = true
	} else if err == nil {
		//log.Printf("Opening existing index...")
	} else {
		log.Fatal(err)
	}
	return index, isNew
}

// Index current DB
func (ldb *LibraryDB) Index() (numIndexed uint64, err error) {
	// open index
	index, _ := openIndex()
	defer index.Close()

	// read the bytes
	jsonBytes, err := ioutil.ReadFile(ldb.DatabaseFile)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonBytes, &ldb.Books)
	if err != nil {
		fmt.Print("Error:", err)
	}

	// index by filename
	for _, epub := range ldb.Books {
		// TODO: index epub.ShortString() instead?
		index.Index(epub.getMainFilename(), epub)
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
func (ldb *LibraryDB) Search(queryString string) (results []Book, err error) {
	// TODO make sure the index is up to date

	fmt.Println("Searching database for " + queryString + " ...")
	query := bleve.NewQueryStringQuery(queryString)
	search := bleve.NewSearchRequest(query)

	// open index
	index, isNew := openIndex()
	if isNew {
		index.Close()
		// indexing db
		fmt.Println("New index, populating...")
		numIndexed, err := ldb.Index()
		if err != nil {
			return nil, err
		}
		fmt.Println("Saved and indexed " + strconv.FormatUint(numIndexed, 10) + " epubs.")
		// reopening
		index, _ = openIndex()
	}
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
			var epub *Book
			epub, err = ldb.FindByFilename(hit.ID)
			if err != nil {
				return
			}
			results = append(results, *epub)
		}
	}
	return
}

// SearchJSON current DB and output as JSON
func (ldb *LibraryDB) SearchJSON() (jsonOutput string, err error) {
	// TODO Search() then get JSON output from each result Epub
	// TODO OR --- the opposite. bleve can return JSON, Search has to parse it and locate the relevant Epub objects
	fmt.Println("Searching database with JSON output...")
	return
}

// ListNonRetailOnly among known epubs.
func (ldb *LibraryDB) ListNonRetailOnly() (nonretail []Book, err error) {
	// TODO return Search for querying non retail epubs, removing the epubs with same title/author but retail
	return
}

// ListRetailOnly among known epubs.
func (ldb *LibraryDB) ListRetailOnly() (retail []Book, err error) {
	return
}

// ListAuthors among known epubs.
func (ldb *LibraryDB) ListAuthors() (authors []string, err error) {
	return
}

// ListTags associated with known epubs.
func (ldb *LibraryDB) ListTags() (tags []string, err error) {
	// TODO search for tags in all epubs, remove duplicates
	return
}

// ListUntagged among known epubs.
func (ldb *LibraryDB) ListUntagged() (untagged []Book, err error) {
	return
}

// ListWithTag among known epubs.
func (ldb *LibraryDB) ListWithTag(tag string) (tagged []Book, err error) {
	return
}
