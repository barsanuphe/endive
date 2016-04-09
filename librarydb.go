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

// Refresh current DB
func (ldb *LibraryDB) Refresh() (renamed int, err error) {
	fmt.Println("Refreshing database...")
	for _, epub := range ldb.Epubs {
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
