package book

import (
	"errors"
	"reflect"
	"strconv"

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

// filter Books with a given function
func (bks Books) filter(f func(*Book) bool) (filteredBooks Books) {
	filteredBooks = make(Books, 0)
	for _, v := range bks {
		if f(&v) {
			filteredBooks = append(filteredBooks, v)
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

// FilterIncomplete among Books.
func (bks *Books) FilterIncomplete() Books {
	return bks.filter(func(b *Book) bool { return !b.Metadata.IsComplete() })
}

// FilterByProgress among Books.
func (bks *Books) FilterByProgress(progress string) Books {
	return bks.filter(func(b *Book) bool { return b.Progress == progress })
}

// FilterUntagged among Books.
func (bks *Books) FilterUntagged() Books {
	return bks.filter(func(b *Book) bool { return len(b.Metadata.Tags) == 0 })
}

// FilterRetail among Books.
func (bks *Books) FilterRetail() Books {
	return bks.filter(func(b *Book) bool { return b.HasRetail() })
}

// FilterNonRetailOnly among Books.
func (bks *Books) FilterNonRetailOnly() Books {
	return bks.filter(func(b *Book) bool { return !b.HasRetail() })
}

// FindByID among known Books
func (bks *Books) FindByID(id int) (b *Book, err error) {
	b = bks.findUnique(func(b *Book) bool { return b.ID == id })
	if b.ID == 0 {
		err = errors.New("Could not find book with ID " + strconv.Itoa(id))
	}
	return
}

// FindByFullPath among known Books
func (bks *Books) FindByFullPath(filename string) (b *Book, err error) {
	b = bks.findUnique(func(b *Book) bool {
		return b.RetailEpub.FullPath() == filename || b.NonRetailEpub.FullPath() == filename
	})
	if b.ID == 0 {
		err = errors.New("Could not find book with epub " + filename)
	}
	return
}

// FindByMetadata among known Books
func (bks *Books) FindByMetadata(i Metadata) (b *Book, err error) {
	b = bks.findUnique(func(b *Book) bool {
		return b.Metadata.IsSimilar(i) || b.EpubMetadata.IsSimilar(i)
	})
	if b.ID == 0 {
		err = errors.New("Could not find book with info " + i.String())
	}
	return
}

//FindByHash among known Books
func (bks *Books) FindByHash(hash string) (b *Book, err error) {
	b = bks.findUnique(func(b *Book) bool {
		return b.RetailEpub.Hash == hash || b.NonRetailEpub.Hash == hash
	})
	if b.ID == 0 {
		err = errors.New("Could not find book with hash " + hash)
	}
	return
}

// Diff detects differences between two sets of Books.
func (bks Books) Diff(o e.Collection) (newB e.Collection, modifiedB e.Collection, deletedB e.Collection) {
	// convert o from Collection to Books
	var oBooks Books
	oBooks.Add(o.Books()...)

	if len(bks) != 0 {
		// updating config, ui in other Books for comparison
		// otherwise FullPath will always be different.
		config := bks[0].Config
		ui := bks[0].UI
		for i := range oBooks {
			oBooks[i].UI = ui
			oBooks[i].Config = config
			oBooks[i].RetailEpub.Config = config
			oBooks[i].RetailEpub.UI = ui
			oBooks[i].NonRetailEpub.Config = config
			oBooks[i].NonRetailEpub.UI = ui
		}
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
		ov, _ := oBooks.FindByFullPath(v.FullPath())
		if !reflect.DeepEqual(v, ov) {
			modifiedB.Add(&v)
		}
	}
	return
}
