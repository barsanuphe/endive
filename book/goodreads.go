package book

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"html"

	h "github.com/barsanuphe/endive/helpers"
)

const apiRoot = "https://www.goodreads.com/"

type Response struct {
	Book   GoodreadsBook   `xml:"book"`
	Search GoodreadsSearch `xml:"search"`
}

type GoodreadsSearch struct {
	ResultsNumber string          `xml:"total-results"`
	Works         []GoodreadWorks `xml:"results>work"`
}

type GoodreadWorks struct {
	ID     string `xml:"best_book>id"`
	Title  string `xml:"best_book>title"`
	Author string `xml:"best_book>author>name"`
}

type GoodreadSeries struct {
	ID       string `xml:"id"`
	Title    string `xml:"series>title"`
	Position string `xml:"user_position"`
}

type GoodreadAuthor struct {
	ID   string `xml:"id"`
	Name string `xml:"name"`
}

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

func (b GoodreadsBook) Author() GoodreadAuthor {
	return b.Authors[0]
}
func (b GoodreadsBook) MainSeries() GoodreadSeries {
	return b.Series[0]
}

func (b GoodreadsBook) SeriesString() string {
	return fmt.Sprintf("%s #%s", strings.TrimSpace(b.MainSeries().Title), b.MainSeries().Position)
}

func (b GoodreadsBook) FullTitle() string {
	if len(b.Series) != 0 {
		return fmt.Sprintf("%s (%s) %s [%s]", b.Author().Name, b.Year, b.OriginalTitle, b.SeriesString())
	}
	return fmt.Sprintf("%s (%s) %s", b.Author().Name, b.Year, b.OriginalTitle)
}

//------------------------

func getData(uri string, i interface{}) {
	data := getRequest(uri)
	xmlUnmarshal(data, i)
}

func getRequest(uri string) []byte {
	res, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func xmlUnmarshal(b []byte, i interface{}) {
	err := xml.Unmarshal(b, i)
	if err != nil {
		log.Fatal(err)
	}
}

func GetBook(id, key string) GoodreadsBook {
	defer h.TimeTrack(time.Now(), "Getting Book info")
	uri := apiRoot + "book/show/" + id + ".xml?key=" + key
	response := Response{}
	getData(uri, &response)
	return response.Book
}

func makeSearchQ(parts ...string) (query string) {
	query = strings.Join(parts, "+")
	r := strings.NewReplacer(" ", "+")
	return html.EscapeString(r.Replace(query))
}

func GetBookIDByQuery(author, title, key string) (id string) {
	defer h.TimeTrack(time.Now(), "Getting Book ID")

	uri := apiRoot + "search/index.xml?key=" + key + "&q=" + makeSearchQ(author, title)
	response := Response{}
	getData(uri, &response)
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

// GetBookIDByISBN using search
func GetBookIDByISBN(isbn, key string) (id string) {
	defer h.TimeTrack(time.Now(), "Getting Book ID")
	// TODO
	return
}
