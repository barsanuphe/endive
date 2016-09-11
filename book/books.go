package book

import (
	"errors"
	"reflect"
	"strconv"

	e "github.com/barsanuphe/endive/endive"
)

// Books is a slice of Book.
type Books []Book

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
func (bks Books) Diff(o Books) (newB map[string]Book, modifiedB map[string]Book, deletedB map[string]Book) {
	newB = make(map[string]Book)
	modifiedB = make(map[string]Book)
	deletedB = make(map[string]Book)

	if len(bks) != 0 {
		// updating config, ui in other Books for comparison
		// otherwise FullPath will always be different.
		config := bks[0].Config
		ui := bks[0].UI
		for i := range o {
			o[i].UI = ui
			o[i].Config = config
			o[i].RetailEpub.Config = config
			o[i].RetailEpub.UI = ui
			o[i].NonRetailEpub.Config = config
			o[i].NonRetailEpub.UI = ui
		}
	}

	// list current
	knownIndexed := make(map[string]Book)
	knownFullPaths := []string{}
	for _, b := range bks {
		knownIndexed[b.FullPath()] = b
		knownFullPaths = append(knownFullPaths, b.FullPath())
	}

	// list other
	otherIndexed := make(map[string]Book)
	otherFullPaths := []string{}
	for _, b := range o {
		otherIndexed[b.FullPath()] = b
		otherFullPaths = append(otherFullPaths, b.FullPath())
	}

	// if in current and not in other, append to new
	commonPaths := []string{}
	for _, k := range knownFullPaths {
		if _, isIn := e.StringInSlice(k, otherFullPaths); !isIn {
			newB[k] = knownIndexed[k]
		} else {
			commonPaths = append(commonPaths, k)
		}
	}
	// if in other and not in current, append to deleted
	for _, p := range otherFullPaths {
		if _, isIn := e.StringInSlice(p, knownFullPaths); !isIn {
			deletedB[p] = otherIndexed[p]
		}
	}
	// if in both, compare them directly, if different, append to modified
	for _, v := range commonPaths {
		// compare
		if !reflect.DeepEqual(knownIndexed[v], otherIndexed[v]) {
			modifiedB[v] = knownIndexed[v]
		}
	}
	return
}
