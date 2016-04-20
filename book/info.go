package book

import (
	"fmt"
	"strings"
)

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
	Tags          Tags     `xml:"popular_shelves>shelf"`
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
