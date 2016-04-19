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

// Response is the top xml element in goodreads response.
type Response struct {
	Book   GoodreadsBook   `xml:"book"`
	Search GoodreadsSearch `xml:"search"`
}

// GoodreadsSearch is the main xml element in goodreads search.
type GoodreadsSearch struct {
	ResultsNumber string          `xml:"total-results"`
	Works         []GoodreadWorks `xml:"results>work"`
}

// GoodreadWorks holds the work information in the xml reponse.
type GoodreadWorks struct {
	ID     string `xml:"best_book>id"`
	Title  string `xml:"best_book>title"`
	Author string `xml:"best_book>author>name"`
}

// GoodreadSeries has information about the series a book is part of.
type GoodreadSeries struct {
	ID       string `xml:"id"`
	Title    string `xml:"series>title"`
	Position string `xml:"user_position"`
}

// GoodreadAuthor is the author of a book.
type GoodreadAuthor struct {
	ID   string `xml:"id"`
	Name string `xml:"name"`
}

// GoodreadsBook contains all of the known book metadata.
type GoodreadsBook struct {
	ID            string           `xml:"id"`
	Title         string           `xml:"title"`
	OriginalTitle string           `xml:"work>original_title"`
	ImageURL      string           `xml:"image_url"`
	NumPages      string           `xml:"num_pages"`
	Format        string           `xml:"format"`
	Authors       []GoodreadAuthor `xml:"authors>author"`
	ISBN          string           `xml:"isbn"`
	Year          string           `xml:"work>original_publication_year"`
	Description   string           `xml:"description"`
	Series        []GoodreadSeries `xml:"series_works>series_work"`
	Rating        string           `xml:"average_rating"`
}

// Author return a GoodreadsBook's main author.
func (b GoodreadsBook) Author() GoodreadAuthor {
	return b.Authors[0]
}

// MainSeries return a GoodreadsBook's main series.
func (b GoodreadsBook) MainSeries() GoodreadSeries {
	return b.Series[0]
}

// SeriesString returns a representation of a GoodreadsBook's main series.
func (b GoodreadsBook) SeriesString() string {
	return fmt.Sprintf("%s #%s", strings.TrimSpace(b.MainSeries().Title), b.MainSeries().Position)
}

// String returns a representation of a GoodreadsBook
func (b GoodreadsBook) String() string {
	if len(b.Series) != 0 {
		return fmt.Sprintf("%s (%s) %s [%s]", b.Author().Name, b.Year, b.OriginalTitle, b.SeriesString())
	}
	return fmt.Sprintf("%s (%s) %s", b.Author().Name, b.Year, b.OriginalTitle)
}

//------------------------

// GetBook returns a GoodreadsBook from its Goodreads ID
func GetBook(id, key string) GoodreadsBook {
	defer h.TimeTrack(time.Now(), "Getting Book info")
	uri := apiRoot + "book/show/" + id + ".xml?key=" + key
	response := Response{}
	h.GetXMLData(uri, &response)
	return response.Book
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
	response := Response{}
	h.GetXMLData(uri, &response)
	// parsing results
	hits, err := strconv.Atoi(response.Search.ResultsNumber)
	if err != nil {
		fmt.Println("error")
	}
	if hits != 0 {
		for _, work := range response.Search.Works {
			if work.Author == author && work.Title == title {
				return work.ID
			}
		}
		fmt.Println("Could not find exact match, returning first hit.")
		return response.Search.Works[0].ID
	}
	return
}

// GetBookIDByISBN gets a Goodreads ID from an ISBN
func GetBookIDByISBN(isbn, key string) (id string) {
	defer h.TimeTrack(time.Now(), "Getting Book ID")
	// TODO
	return
}
