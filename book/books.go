package book

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/kylelemons/godebug/pretty"

	e "github.com/barsanuphe/endive/endive"
)

// Books is a slice of Book.
type Books []Book

func bookSliceToGeneric(x Books) (y []e.GenericBook) {
	y = make([]e.GenericBook, len(x))
	for i := range x {
		y[i] = &x[i]
	}
	return
}

// Books returns GenericBooks
func (bks *Books) Books() []e.GenericBook {
	return bookSliceToGeneric(*bks)
}

// Add a Book
func (bks *Books) Add(books ...e.GenericBook) {
	for _, b := range books {
		var book Book
		book = *b.(*Book)
		*bks = append(*bks, book)
	}
}

// Propagate for all Books to be aware of Config and UI.
func (bks *Books) Propagate(ui e.UserInterface, c e.Config) {
	// make each Book aware of current Config + UI
	for i := range *bks {
		(*bks)[i].Config = c
		(*bks)[i].UI = ui
		(*bks)[i].NonRetailEpub.Config = c
		(*bks)[i].NonRetailEpub.UI = ui
		(*bks)[i].RetailEpub.Config = c
		(*bks)[i].RetailEpub.UI = ui
	}
}

// filter Books with a given function
func (bks *Books) filter(f func(*Book) bool) (filteredBooks Books) {
	for _, v := range *bks {
		if f(&v) {
			filteredBooks.Add(&v)
		}
	}
	return
}

// findUnique Book with a given function
func (bks *Books) findUnique(f func(*Book) bool) *Book {
	for i, v := range *bks {
		if f(&v) {
			return &(*bks)[i]
		}
	}
	return &Book{}
}

// Incomplete among Books.
func (bks *Books) Incomplete() e.Collection {
	incomplete := bks.filter(func(b *Book) bool { return !b.Metadata.IsComplete() })
	var res e.Collection
	res = &incomplete
	return res
}

// Progress among Books.
func (bks *Books) Progress(progress string) e.Collection {
	prgrs := bks.filter(func(b *Book) bool { return b.Progress == progress })
	var res e.Collection
	res = &prgrs
	return res
}

// Retail among Books.
func (bks *Books) Retail() e.Collection {
	retail := bks.filter(func(b *Book) bool { return b.HasRetail() })
	var res e.Collection
	res = &retail
	return res
}

// NonRetailOnly among Books.
func (bks *Books) NonRetailOnly() e.Collection {
	nonretail := bks.filter(func(b *Book) bool { return !b.HasRetail() })
	var res e.Collection
	res = &nonretail
	return res
}

// Exported among Books.
func (bks *Books) Exported() e.Collection {
	exported := bks.filter(func(b *Book) bool { return b.IsExported == e.True })
	var res e.Collection
	res = &exported
	return res
}

// FindByID among known Books
func (bks *Books) FindByID(id int) (e.GenericBook, error) {
	b := bks.findUnique(func(b *Book) bool { return b.ID() == id })
	if b.ID() == 0 {
		return nil, errors.New("Could not find book with ID " + strconv.Itoa(id))
	}
	var result e.GenericBook
	result = b
	return result, nil
}

// RemoveByID a book
func (bks *Books) RemoveByID(id int) (err error) {
	var found bool
	removeIndex := -1
	for i := range *bks {
		if (*bks)[i].ID() == id {
			found = true
			removeIndex = i
			break
		}
	}
	if found {
		*bks = append((*bks)[:removeIndex], (*bks)[removeIndex+1:]...)
	} else {
		err = errors.New("Did not find book with ID " + strconv.Itoa(id))
	}
	return
}

// FindByFullPath among known Books
func (bks *Books) FindByFullPath(filename string) (e.GenericBook, error) {
	if filename == "" {
		return nil, errors.New("empty path")
	}
	b := bks.findUnique(func(b *Book) bool {
		return b.RetailEpub.FullPath() == filename || b.NonRetailEpub.FullPath() == filename
	})
	if b.ID() == 0 {
		return nil, errors.New("Could not find book with epub " + filename)
	}
	var result e.GenericBook
	result = b
	return result, nil
}

// FindByMetadata among known Books
func (bks *Books) FindByMetadata(isbn, authors, title string) (e.GenericBook, error) {
	isbnCandidate, err := e.CleanISBN(isbn)
	if (authors == "" && title == "") && err != nil {
		return nil, errors.New("invalid isbn and/or empty author and title")
	}
	o := Metadata{ISBN: isbnCandidate, Authors: []string{authors}, BookTitle: title}
	b := bks.findUnique(func(b *Book) bool {
		return b.Metadata.IsSimilar(o)
	})
	if b.ID() == 0 {
		return nil, errors.New("Could not find book with info " + o.String())
	}
	var result e.GenericBook
	result = b
	return result, nil
}

//FindByHash among known Books
func (bks *Books) FindByHash(hash string) (e.GenericBook, error) {
	if hash == "" {
		return nil, errors.New("empty hash")
	}
	b := bks.findUnique(func(b *Book) bool {
		return b.RetailEpub.Hash == hash || b.NonRetailEpub.Hash == hash
	})
	if b.ID() == 0 {
		return nil, errors.New("Could not find book with hash " + hash)
	}
	var result e.GenericBook
	result = b
	return result, nil
}

// Authors among known epubs.
func (bks *Books) Authors() (authors map[string]int) {
	authors = make(map[string]int)
	for _, book := range *bks {
		author := book.Metadata.Author()
		authors[author]++
	}
	return
}

// Publishers among known epubs.
func (bks *Books) Publishers() (publishers map[string]int) {
	publishers = make(map[string]int)
	for _, book := range *bks {
		if book.Metadata.Publisher != "" {
			publisher := book.Metadata.Publisher
			publishers[publisher]++
		} else {
			publishers["Unknown"]++
		}
	}
	return
}

// Tags associated with known epubs.
func (bks *Books) Tags() (tags map[string]int) {
	tags = make(map[string]int)
	for _, book := range *bks {
		for _, tag := range book.Metadata.Tags {
			tags[tag.Name]++
		}
	}
	return
}

// Series associated with known epubs.
func (bks *Books) Series() (series map[string]int) {
	series = make(map[string]int)
	for _, book := range *bks {
		for _, s := range book.Metadata.Series {
			series[s.Name]++
		}
	}
	return
}

// Diff detects differences between two sets of Books.
func (bks Books) Diff(o e.Collection, newB e.Collection, modifiedB e.Collection, deletedB e.Collection) {
	// convert o from Collection to Books
	var oBooks Books
	oBooks.Add(o.Books()...)

	if len(bks) != 0 {
		// updating config, ui in other Books for comparison
		// otherwise FullPath will always be different.
		config := bks[0].Config
		ui := bks[0].UI
		oBooks.Propagate(ui, config)
	}

	// list current
	knownFullPaths := []string{}
	for _, b := range bks {
		knownFullPaths = append(knownFullPaths, b.FullPath())
	}
	// list other
	otherFullPaths := []string{}
	for _, b := range oBooks {
		otherFullPaths = append(otherFullPaths, b.FullPath())
	}

	// if in current and not in other, append to new
	commonBooks := []Book{}
	for _, k := range bks {
		if _, isIn := e.StringInSlice(k.FullPath(), otherFullPaths); !isIn {
			newB.Add(&k)
		} else {
			commonBooks = append(commonBooks, k)
		}
	}
	// if in other and not in current, append to deleted
	for _, p := range oBooks {
		if _, isIn := e.StringInSlice(p.FullPath(), knownFullPaths); !isIn {
			deletedB.Add(&p)
		}
	}
	// if in both, compare them directly, if different, append to modified
	for _, v := range commonBooks {
		// compare
		ov, err := oBooks.FindByFullPath(v.FullPath())
		if err != nil {
			v.UI.Error(err.Error())
		} else {
			var ob *Book
			ob = ov.(*Book)
			// textual diff in struct fields: ignores UI, Config, etc
			// since we're only interested in string/int fields really, this is enough
			if pretty.Compare(ob, v) != "" {
				modifiedB.Add(&v)
			}
		}
	}
	return
}

// Table of books
func (bks Books) Table() string {
	if len(bks.Books()) == 0 {
		return ""
	}
	var rows [][]string
	for _, res := range bks {
		relativePath, err := filepath.Rel(res.Config.LibraryRoot, res.FullPath())
		if err != nil {
			panic(errors.New("File " + res.FullPath() + " not in library?"))
		}
		id := fmt.Sprintf("%d", res.ID())
		if res.IsExported == e.True {
			id += " ⇲"
		}
		rows = append(rows, []string{id, res.Metadata.Author(), res.Metadata.Title(), res.Metadata.OriginalYear, relativePath})
	}
	return e.TabulateRows(rows, "ID", "Author", "Title", "Year", "Filename")
}

// Sort books
func (bks Books) Sort(sortBy string) {
	SortBooks(bks, sortBy)
}

// First books
func (bks Books) First(nb int) e.Collection {
	var res e.Collection
	if nb > 0 && len(bks) > nb {
		bks = bks[:nb]
	}
	res = &bks
	return res
}

// Last books
func (bks Books) Last(nb int) e.Collection {
	var res e.Collection
	if nb > 0 && len(bks) > nb {
		bks = bks[len(bks)-nb:]
	}
	res = &bks
	return res
}
