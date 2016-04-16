package library

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	b "github.com/barsanuphe/endive/book"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/language/en"
)

func buildIndexMapping() (*bleve.IndexMapping, error) {
	// TODO index everything

	// a generic reusable mapping for english text
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = en.AnalyzerName

	epubMapping := bleve.NewDocumentMapping()

	epubMapping.AddFieldMappingsAt("progress", textFieldMapping)
	epubMapping.AddFieldMappingsAt("description", textFieldMapping)
	epubMapping.AddFieldMappingsAt("language", textFieldMapping)
	epubMapping.AddFieldMappingsAt("creator", textFieldMapping)
	epubMapping.AddFieldMappingsAt("title", textFieldMapping)
	epubMapping.AddFieldMappingsAt("year", textFieldMapping)
	epubMapping.AddFieldMappingsAt("isbn", textFieldMapping)
	epubMapping.AddFieldMappingsAt("rating", textFieldMapping)
	epubMapping.AddFieldMappingsAt("tags", textFieldMapping)

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
func (ldb *DB) Index() (numIndexed uint64, err error) {
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
	for _, book := range ldb.Books {
		index.Index(book.GetMainFilename(), book)
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
func (ldb *DB) Search(queryString string) (results []b.Book, err error) {
	// TODO make sure the index is up to date

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
			var epub *b.Book
			epub, err = ldb.FindByFilename(hit.ID)
			if err != nil {
				return
			}
			results = append(results, *epub)
		}
	}
	return
}
