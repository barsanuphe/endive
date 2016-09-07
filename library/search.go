package library

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"

	b "github.com/barsanuphe/endive/book"
	h "github.com/barsanuphe/endive/helpers"

	"github.com/blevesearch/bleve"
)

func openIndex() (index bleve.Index, isNew bool) {
	index, err := bleve.Open(getIndexPath())
	if err == bleve.ErrorIndexPathDoesNotExist {
		h.Debug("Creating new index...")
		index, err = bleve.New(getIndexPath(), bleve.NewIndexMapping())
		if err != nil {
			h.Error(err.Error())
		}
		isNew = true
	} else if err == nil {
		//log.Printf("Opening existing index...")
	} else {
		h.Error(err.Error())
	}
	return index, isNew
}

// IndexAll current DB
func (ldb *DB) IndexAll() (numIndexed uint64, err error) {
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
		h.Errorf("Error: %s", err.Error())
	}

	// index by path
	for _, book := range ldb.Books {
		err = index.Index(book.FullPath(), book)
		if err != nil {
			return
		}
	}

	// check number of indexed documents
	numIndexed, err = index.DocCount()
	if err != nil {
		return
	}

	h.Debug("Indexed: " + strconv.FormatUint(numIndexed, 10) + " epubs.")
	return
}

// indexAdd add Books to index
func (ldb *DB) indexAdd(books map[string]b.Book) (err error) {
	// open index
	index, _ := openIndex()
	defer index.Close()

	for k, v := range books {
		err = index.Index(k, v)
		if err != nil {
			return
		}
	}
	return
}

// indexDelete delete Books from index
func (ldb *DB) indexDelete(books map[string]b.Book) (err error) {
	// open index
	index, _ := openIndex()
	defer index.Close()

	for k := range books {
		err = index.Delete(k)
		if err != nil {
			return
		}
	}

	return
}

// IndexDiff indexes incremental modifications
func (ldb *DB) IndexDiff(newB map[string]b.Book, modifiedB map[string]b.Book, deletedB map[string]b.Book) (err error) {
	// delete books
	err = ldb.indexDelete(deletedB)
	if err != nil {
		return
	}
	// remove index for modified books too
	err = ldb.indexDelete(modifiedB)
	if err != nil {
		return
	}
	// add new books
	err = ldb.indexAdd(newB)
	if err != nil {
		return
	}
	// add modified books
	err = ldb.indexAdd(modifiedB)
	if err != nil {
		return
	}
	return
}

// RunQuery on current DB
func (ldb *DB) RunQuery(queryString string) (results []b.Book, err error) {
	queryString = ldb.prepareQuery(queryString)
	query := bleve.NewQueryStringQuery(queryString)
	// NOTE: second argument is max number of hits
	search := bleve.NewSearchRequestOptions(query, 1000, 0, false)
	// open index
	index, isNew := openIndex()
	if isNew {
		index.Close()
		// indexing db
		h.Debug("New index, populating...")
		numIndexed, err := ldb.IndexAll()
		if err != nil {
			return nil, err
		}
		h.Debug("Saved and indexed " + strconv.FormatUint(numIndexed, 10) + " epubs.")
		// reopening
		index, _ = openIndex()
	}
	defer index.Close()

	searchResults, err := index.Search(search)
	if err != nil {
		h.Error(err.Error())
		return
	}
	//fmt.Println(searchResults.Total)
	if searchResults.Total != 0 {
		for _, hit := range searchResults.Hits {
			var epub *b.Book
			epub, err = ldb.Books.FindByFilename(hit.ID)
			if err != nil {
				return
			}
			results = append(results, *epub)
		}
	}
	return
}

// prepareQuery before search
func (ldb *DB) prepareQuery(queryString string) (newQuery string) {
	// replace fields for simpler queries
	r := strings.NewReplacer(
		"author:", "metadata.authors:",
		"title:", "metadata.title:",
		"year:", "metadata.year:",
		"language:", "metadata.language:",
		"series:", "metadata.series.seriesname:",
		"tags:", "metadata.tags.name:",
		"tag:", "metadata.tags.name:",
		"publisher:", "metadata.publisher:",
		"category:", "metadata.category:",
		"genre:", "metadata.main_genre:",
		"description:", "metadata.description:",
	)
	return r.Replace(queryString)
}
