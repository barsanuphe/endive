package book

import (
	"sort"
	"strings"

	h "github.com/barsanuphe/endive/helpers"
)

// TODO add all possible orders
var validSortOrder = []string{idField, authorField, titleField, seriesField, genreField, categoryField, ratingField, averageRatingField, yearField}

// CheckValidSortOrder checks if a sorting field is valid.
func CheckValidSortOrder(sortBy string) (valid bool) {
	_, valid = h.StringInSlice(strings.ToLower(sortBy), validSortOrder)
	return
}

// By is the type of a "less" function that defines the ordering of its Book arguments.
type By func(p1, p2 *Book) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(books []Book) {
	ps := &bookSorter{
		books: books,
		by:    by,
	}
	sort.Sort(ps)
}

// bookSorter joins a By function and a slice of Books to be sorted.
type bookSorter struct {
	books []Book
	by    func(p1, p2 *Book) bool
}

// Len is part of sort.Interface.
func (s *bookSorter) Len() int {
	return len(s.books)
}

// Swap is part of sort.Interface.
func (s *bookSorter) Swap(i, j int) {
	s.books[i], s.books[j] = s.books[j], s.books[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *bookSorter) Less(i, j int) bool {
	return s.by(&s.books[i], &s.books[j])
}

// SortBooks using a specific field.
func SortBooks(books []Book, orderBy string) {
	// TODO : tests

	if !CheckValidSortOrder(orderBy) {
		return
	}

	id := func(p1, p2 *Book) bool {
		return p1.ID < p2.ID
	}
	originalYear := func(p1, p2 *Book) bool {
		return p1.Metadata.OriginalYear < p2.Metadata.OriginalYear
	}
	year := func(p1, p2 *Book) bool {
		return p1.Metadata.EditionYear < p2.Metadata.EditionYear
	}
	author := func(p1, p2 *Book) bool {
		return p1.Metadata.Author() < p2.Metadata.Author()
	}
	title := func(p1, p2 *Book) bool {
		return p1.Metadata.Title() < p2.Metadata.Title()
	}

	switch orderBy {
	case idField:
		By(id).Sort(books)
	case authorField:
		By(author).Sort(books)
	case titleField:
		By(title).Sort(books)
	case yearField:
		By(originalYear).Sort(books)
	case editionYearField:
		By(year).Sort(books)
		// TODO all cases
	}
}
