package index

import (
	"errors"
	"os"

	"github.com/blevesearch/bleve"

	e "github.com/barsanuphe/endive/endive"
)

// Index implements Indexer
type Index struct {
	Path            string
	NeedsRebuilding bool
}

// SetPath for Index
func (i *Index) SetPath(path string) {
	// TODO check exists, make parents
	i.Path = path
}

// Count the number of indexed GenericBooks.
func (i *Index) Count() uint64 {
	index, _, err := i.open()
	if err != nil {
		return 0
	}
	defer index.Close()

	// check number of indexed documents
	count, err := index.DocCount()
	if err != nil {
		return 0
	}
	return count
}

// Rebuild for all GenericBooks
func (i *Index) Rebuild(all []e.GenericBook) error {
	// remove old index
	err := os.RemoveAll(i.Path)
	if err != nil {
		return err
	}
	// indexing db
	books := make(map[string]e.GenericBook)
	for _, b := range all {
		books[b.FullPath()] = b
	}
	return i.add(books)
}

// Update existing index
func (i *Index) Update(newB map[string]e.GenericBook, modB map[string]e.GenericBook, delB map[string]e.GenericBook) (err error) {
	// delete books
	err = i.delete(delB)
	if err != nil {
		return
	}
	// remove index for modified books too
	err = i.delete(modB)
	if err != nil {
		return
	}
	// add new books
	err = i.add(newB)
	if err != nil {
		return
	}
	// add modified books
	err = i.add(modB)
	if err != nil {
		return
	}
	return
}

// Query on current Index
func (i *Index) Query(queryString string) (resultsPaths []string, err error) {
	query := bleve.NewQueryStringQuery(queryString)
	// NOTE: second argument is max number of hits
	search := bleve.NewSearchRequestOptions(query, 1000, 0, false)
	// open index
	index, isNew, err := i.open()
	if err != nil {
		return
	}
	defer index.Close()
	if isNew {
		return resultsPaths, errors.New("Index is empty")
	}

	searchResults, err := index.Search(search)
	if err != nil {
		return
	}
	//fmt.Println(searchResults.Total)
	if searchResults.Total != 0 {
		for _, hit := range searchResults.Hits {
			resultsPaths = append(resultsPaths, hit.ID)
		}
	}
	return
}

func (i *Index) open() (index bleve.Index, isNew bool, err error) {
	// TODO check Path is set
	index, err = bleve.Open(i.Path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		index, err = bleve.New(i.Path, bleve.NewIndexMapping())
		if err != nil {
			return
		}
		isNew = true
	}
	return index, isNew, err
}

// indexAdd add Books to index
func (i *Index) add(books map[string]e.GenericBook) (err error) {
	// open index
	index, _, err := i.open()
	if err != nil {
		return
	}
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
func (i *Index) delete(books map[string]e.GenericBook) (err error) {
	// open index
	index, _, err := i.open()
	if err != nil {
		return
	}
	defer index.Close()

	for k := range books {
		err = index.Delete(k)
		if err != nil {
			return
		}
	}

	return
}
