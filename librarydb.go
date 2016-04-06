package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"encoding/json"

	"strconv"

	"bytes"
	"os"

	"errors"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzers/keyword_analyzer"
	"github.com/blevesearch/bleve/analysis/language/en"
)

const indexName string = "endive.index"

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
		if os.IsNotExist(err) {
			// first run
			return nil
		}
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
func (ldb *LibraryDB) Save() (hasSaved bool, err error) {
	jsonEpub, err := json.Marshal(ldb.Epubs)
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
		err = os.RemoveAll(indexName)
		if err != nil {
			fmt.Println(err)
			return
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
	epubMapping.AddFieldMappingsAt("author", textFieldMapping)
	epubMapping.AddFieldMappingsAt("title", textFieldMapping)
	epubMapping.AddFieldMappingsAt("publicationyear", textFieldMapping)
	epubMapping.AddFieldMappingsAt("isbn", textFieldMapping)
	epubMapping.AddFieldMappingsAt("rating", textFieldMapping)
	epubMapping.AddFieldMappingsAt("tags", keywordFieldMapping)

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
	index := openIndex(indexName)
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
		index.Index(epub.Filename, epub)
	}

	// check number of indexed documents
	numIndexed, err = index.DocCount()
	if err != nil {
		return
	}

	fmt.Println("Indexed: " + strconv.FormatUint(numIndexed, 10) + " epubs.")

	return
}

func (ldb *LibraryDB) FindByFilename(filename string) (result Epub, err error) {
	for _, result = range ldb.Epubs {
		if result.Filename == filename {
			return
		}
	}
	return Epub{}, errors.New("Could not find epub " + filename)
}

func (ldb *LibraryDB) hasCopy(e Epub, isRetail bool) (result bool) {
	// TODO tests

	// TODO make sur e.IsRetail is set

	// loop over ldb.Epubs,
	for _, epub := range ldb.Epubs {
		isDuplicate, canTrump := e.IsDuplicate(epub, isRetail)
		if canTrump {
			return false
		}
		if isDuplicate {
			return true
		}
	}
	return
}

// Search current DB
func (ldb *LibraryDB) Search(queryString string) (results []Epub, err error) {
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
			var epub Epub
			epub, err = ldb.FindByFilename(hit.ID)
			if err != nil {
				return
			}
			results = append(results, epub)
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
func (ldb *LibraryDB) ListNonRetailOnly() (nonretail []Epub, err error) {
	// TODO return Search for querying non retail epubs, removing the epubs with same title/author but retail
	return
}

// ListRetailOnly among known epubs.
func (ldb *LibraryDB) ListRetailOnly() (retail []Epub, err error) {
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
func (ldb *LibraryDB) ListUntagged() (untagged []Epub, err error) {
	return
}

// ListWithTag among known epubs.
func (ldb *LibraryDB) ListWithTag(tag string) (tagged []Epub, err error) {
	return
}
