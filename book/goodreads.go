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
	Book   Info          `xml:"book"`
	Search searchResults `xml:"search"`
}

// searchResults is the main xml element in goodreads search.
type searchResults struct {
	ResultsNumber string `xml:"total-results"`
	Works         []work `xml:"results>work"`
}

// works holds the work information in the xml reponse.
type work struct {
	ID     string `xml:"best_book>id"`
	Title  string `xml:"best_book>title"`
	Author string `xml:"best_book>author>name"`
}

// Tag holds the name of a tag.
type Tag struct {
	Name string `json:"tagname" xml:"name,attr"`
}

// Info contains all of the known book metadata.
type Info struct {
	ID            string   `xml:"id"`
	Title         string   `xml:"title"`
	OriginalTitle string   `xml:"work>original_title"`
	ImageURL      string   `xml:"image_url"`
	NumPages      string   `xml:"num_pages"`
	Authors       []string `xml:"authors>author>name"`
	ISBN          string   `xml:"isbn"`
	Year          string   `xml:"work>original_publication_year"`
	Description   string   `xml:"description"`
	Series        Series   `xml:"series_works>series_work"`
	AverageRating string   `xml:"average_rating"`
	Tags          []Tag    `xml:"popular_shelves>shelf"`
}

// Author return a GoodreadsBook's main author.
func (b Info) Author() string {
	return b.Authors[0]
}

// MainSeries return a GoodreadsBook's main series.
func (b Info) MainSeries() SingleSeries {
	return b.Series[0]
}

// SeriesString returns a representation of a GoodreadsBook's main series.
func (b Info) SeriesString() string {
	return fmt.Sprintf("%s #%s", strings.TrimSpace(b.MainSeries().Name), b.MainSeries().Position)
}

// String returns a representation of a GoodreadsBook
func (b Info) String() string {
	if len(b.Series) != 0 {
		return fmt.Sprintf("%s (%s) %s [%s]", b.Author(), b.Year, b.OriginalTitle, b.SeriesString())
	}
	return fmt.Sprintf("%s (%s) %s", b.Author(), b.Year, b.OriginalTitle)
}

//------------------------

func cleanTags(g *Info) {
	cleanTags := []Tag{}
	// TODO: names of months, dates
	// remove shelf names that are obviously not genres
	forbiddenTags := []string{
		"own", "school", "favorite", "favourite", "book",
		"read", "kindle", "borrowed", "classic", "novel", "buy",
		"star", "release", "wait", "soon", "wish", "published", "want",
		"tbr", "series", "finish", "to-", "not-", "library", "audible",
		"coming", "anticipated", "default", "recommended", "-list", "sequel",
	}
	for _, tag := range g.Tags {
		clean := true
		for _, ft := range forbiddenTags {
			if strings.Contains(tag.Name, ft) {
				clean = false
				break
			}
		}
		if clean {
			cleanTags = append(cleanTags, tag)
		}
	}
	g.Tags = cleanTags
}

// GetBook returns a GoodreadsBook from its Goodreads ID
func GetBook(id, key string) Info {
	defer h.TimeTrack(time.Now(), "Getting Book info")
	uri := apiRoot + "book/show/" + id + ".xml?key=" + key
	response := response{}
	h.GetXMLData(uri, &response)
	cleanTags(&response.Book)
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
	response := response{}
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
