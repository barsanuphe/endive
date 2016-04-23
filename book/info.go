package book

import (
	"fmt"
	"strings"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
)

// Info contains all of the known book metadata.
type Info struct {
	ID            string   `json:"-" xml:"id"` // TODO see if useful to even parse from xml
	MainTitle     string   `json:"title" xml:"title"`
	OriginalTitle string   `json:"original_title" xml:"work>original_title"`
	ImageURL      string   `json:"image_url" xml:"image_url"`
	NumPages      string   `json:"num_pages" xml:"num_pages"`
	Authors       []string `json:"authors" xml:"authors>author>name"`
	ISBN          string   `json:"isbn" xml:"isbn"`
	Year          string   `json:"year" xml:"work>original_publication_year"`
	Description   string   `json:"description" xml:"description"`
	Series        Series   `json:"series" xml:"series_works>series_work"`
	AverageRating string   `json:"average_rating" xml:"average_rating"`
	Tags          Tags     `json:"tags" xml:"popular_shelves>shelf"`
	Language      string   `json:"-" xml:"TODO: FIND XML PATH"`
}

// String returns a representation of a GoodreadsBook
func (i *Info) String() string {
	if len(i.Series) != 0 {
		return fmt.Sprintf("%s (%s) %s [%s]", i.Author(), i.Year, i.Title(), i.MainSeries().String())
	}
	return fmt.Sprintf("%s (%s) %s", i.Author(), i.Year, i.Title())
}

// HasAny checks if metadata was parsed.
func (i *Info) HasAny() (hasInfo bool) {
	// TODO: better check
	// if Info does not have a title, chances are it's empty.
	if i.Title() != "" {
		return true
	}
	return
}

// Title returns Info's main title.
func (i *Info) Title() string {
	if i.OriginalTitle != "" {
		return i.OriginalTitle
	}
	return i.MainTitle
}

// Clean cleans up the Info
func (i *Info) Clean()  {
	// default year
	if i.Year == "" {
		i.Year = "XXXX"
	}
	// clean tags
	i.Tags.Clean()
	// clean series
	for j := range i.Series {
		i.Series[j].Name = strings.TrimSpace(i.Series[j].Name)
	}
}

// Author returns Info's main author.
func (i *Info) Author() (author string) {
	author = "Unknown"
	if len(i.Authors) != 0 {
		author = i.Authors[0]
	}
	return
}

// MainSeries return the main Series of Info.
func (i *Info) MainSeries() SingleSeries {
	if len(i.Series) != 0 {
		return i.Series[0]
	}
	return SingleSeries{}
}

// IsSimilar checks if metadata is similar to known Info.
func (i *Info) IsSimilar(o Info) (isSimilar bool) {
	// TODO do much better, try with isbn if available on both sides
	// similar == same author/title, for now
	if i.Author() == o.Author() && i.Title() == o.Title() {
		return true
	}
	return
}

// Refresh updates Info fields, using the configuration file.
func (i *Info) Refresh(c cfg.Config) (hasChanged bool) {
	// for now, only taking into account author aliases
	for j, author := range i.Authors {
		for mainalias, aliases := range c.AuthorAliases {
			_, isIn := h.StringInSlice(author, aliases)
			if isIn {
				i.Authors[j] = mainalias
				break
			}
		}
	}
	return
}
