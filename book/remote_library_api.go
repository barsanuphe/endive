package book

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	h "github.com/barsanuphe/endive/helpers"
)

// RemoteLibraryAPI is the interface for accessing remote library information.
type RemoteLibraryAPI interface {
	GetBook(id, key string) (Metadata, error)
	GetBookIDByQuery(author, title, key string) (id string, err error)
	GetBookIDByISBN(isbn, key string) (id string, err error)
}

// GoodReads implements RemoteLibraryAPI and retrieves information from goodreads.com.
type GoodReads struct {
}

const apiRoot = "https://www.goodreads.com/"

// response is the top xml element in goodreads response.
type response struct {
	Book   Metadata      `xml:"book"`
	Search searchResults `xml:"search"`
}

// searchResults is the main xml element in goodreads search.
type searchResults struct {
	ResultsNumber string `xml:"total-results"`
	Works         []work `xml:"results>work"`
}

// works holds the work information in the xml response.
type work struct {
	ID     string `xml:"best_book>id"`
	Author string `xml:"best_book>author>name"`
	Title  string `xml:"best_book>title"`
}

// GetBook returns a GoodreadsBook from its Goodreads ID
func (g GoodReads) GetBook(id, key string) (Metadata, error) {
	uri := apiRoot + "book/show/" + id + ".xml?key=" + key
	r := response{}
	err := h.GetXMLData(uri, &r)
	return r.Book, err
}

func makeSearchQuery(parts ...string) (query string) {
	query = strings.Join(parts, "+")
	r := strings.NewReplacer(" ", "+")
	return html.EscapeString(r.Replace(query))
}

// GetBookIDByQuery gets a Goodreads ID from a query
func (g GoodReads) GetBookIDByQuery(author, title, key string) (id string, err error) {
	uri := apiRoot + "search/index.xml?key=" + key + "&q=" + makeSearchQuery(author, title)
	r := response{}
	err = h.GetXMLData(uri, &r)
	if err != nil {
		return
	}
	// parsing results
	numberOfHits, err := strconv.Atoi(r.Search.ResultsNumber)
	if err != nil {
		return
	}
	if numberOfHits != 0 {
		// TODO: if more than 1 hit, give the user a choice, as in beets import.
		for _, work := range r.Search.Works {
			if work.Author == author && work.Title == title {
				return work.ID, nil
			}
		}
		fmt.Println("Could not find exact match, returning first hit.")
		return r.Search.Works[0].ID, nil
	}
	return
}

// GetBookIDByISBN gets a Goodreads ID from an ISBN
func (g GoodReads) GetBookIDByISBN(isbn, key string) (id string, err error) {
	uri := apiRoot + "search/index.xml?key=" + key + "&q=" + isbn
	r := response{}
	err = h.GetXMLData(uri, &r)
	if err != nil {
		return
	}
	// parsing results
	numberOfHits, err := strconv.Atoi(r.Search.ResultsNumber)
	if err != nil {
		return
	}
	if numberOfHits != 0 {
		id = r.Search.Works[0].ID
		if numberOfHits > 1 {
			fmt.Println("Got more than 1 hit while searching by ISBN! Returned first hit.")
		}
	}
	return
}
