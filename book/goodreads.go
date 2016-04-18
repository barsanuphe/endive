package book

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	h "github.com/barsanuphe/endive/helpers"
)

const apiRoot = "http://www.goodreads.com/"

type Response struct {
	Book GoodreadsBook `xml:"book"`
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
	ID          string           `xml:"id"`
	Title       string           `xml:"work>original_title"`
	ImageURL    string           `xml:"image_url"`
	NumPages    string           `xml:"num_pages"`
	Format      string           `xml:"format"`
	Authors     []GoodreadAuthor `xml:"authors>author"`
	ISBN        string           `xml:"isbn"`
	Year        string           `xml:"publication_year"`
	Description string           `xml:"description"`
	Series      []GoodreadSeries `xml:"series_works>series_work"`
	Rating      string           `xml:"average_rating"`
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
	return fmt.Sprintf("%s (%s) %s [%s]", b.Author().Name, b.Year, b.Title, b.SeriesString())
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
	response := &Response{}
	getData(uri, response)
	return response.Book
}
