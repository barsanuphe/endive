package book

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	h "github.com/barsanuphe/endive/helpers"
)

const apiRoot = "https://www.goodreads.com/"

// response is the top xml element in goodreads response.
type response struct {
	book   Info          `xml:"book"`
	search searchResults `xml:"search"`
}

// searchResults is the main xml element in goodreads search.
type searchResults struct {
	resultsNumber string `xml:"total-results"`
	works         []work `xml:"results>work"`
}

// works holds the work information in the xml reponse.
type work struct {
	id     string `xml:"best_book>id"`
	author string `xml:"best_book>author>name"`
	title  string `xml:"best_book>title"`
}

//------------------------

// GetBook returns a GoodreadsBook from its Goodreads ID
func GetBook(id, key string) Info {
	defer h.TimeTrack(time.Now(), "Getting Book info")
	uri := apiRoot + "book/show/" + id + ".xml?key=" + key
	r := response{}
	h.GetXMLData(uri, &r)
	r.book.Tags.Clean()
	return r.book
}

func makeSearchQ(parts ...string) (query string) {
	query = strings.Join(parts, "+")
	r := strings.NewReplacer(" ", "+")
	return html.EscapeString(r.Replace(query))
}

// GetBookIDByQuery gets a Goodreads ID from a query
func GetBookIDByQuery(author, title, key string) (id string) {
	defer h.TimeTrack(time.Now(), "Getting Book ID")

	uri := apiRoot + "search/index.xml?key=" + key + "&q=" + makeSearchQ(author, title)
	fmt.Println(uri)
	r := response{}
	h.GetXMLData(uri, &r)

	// parsing results
	hits, err := strconv.Atoi(r.search.resultsNumber)
	if err != nil {
		fmt.Println("error")
	}
	if hits != 0 {
		for _, work := range r.search.works {
			if work.author == author && work.title == title {
				return work.id
			}
		}
		fmt.Println("Could not find exact match, returning first hit.")
		return r.search.works[0].id
	}
	return
}

// GetBookIDByISBN gets a Goodreads ID from an ISBN
func GetBookIDByISBN(isbn, key string) (id string) {
	defer h.TimeTrack(time.Now(), "Getting Book ID")
	// TODO
	return
}
