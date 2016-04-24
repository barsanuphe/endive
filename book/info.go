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
	Language      string   `json:"language" xml:"language_code"`
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
func (i *Info) Clean() {
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

// IsSimilar checks if metadata is similar to known Info.
func (i *Info) IsSimilar(o Info) (isSimilar bool) {
	// TODO do much better, try with isbn if available on both sides
	// similar == same author/title, for now
	if i.Author() == o.Author() && i.Title() == o.Title() {
		return true
	}
	return
}

// Diff returns differences between Infos.
func (i *Info) Diff(o Info, firstHeader, secondHeader string) (diff string) {
	var rows [][]string
	// TODO
	rows = append(rows, []string{i.String(), o.String()})
	rows = append(rows, []string{i.Author(), o.Author()})
	rows = append(rows, []string{i.Title(), o.Title()})
	rows = append(rows, []string{i.Year, o.Year})
	rows = append(rows, []string{i.Description, o.Description})
	rows = append(rows, []string{i.Tags.String(), o.Tags.String()})
	rows = append(rows, []string{i.Series.String(), o.Series.String()})
	rows = append(rows, []string{i.Language, o.Language})
	return h.TabulateRows(rows, firstHeader, secondHeader)
}

// Merge with another Info.
func (i *Info) Merge(o Info) (err error) {
	// TODO tests
	// TODO all fields
	if i.Author() != o.Author() {
		index, err := h.Choose(i.Author(), o.Author())
		if err != nil {
			return err
		}
		if index == 1 {
			i.Authors = o.Authors
		}
	}

	if i.Title() != o.Title() {
		index, err := h.Choose(i.Title(), o.Title())
		if err != nil {
			return err
		}
		if index == 1 {
			// TODO show both versions?
			i.MainTitle = o.MainTitle
			i.OriginalTitle = o.OriginalTitle
		}
	}

	i.Year, err = chooseFieldVersion(i.Year, o.Year)
	if err != nil {
		return
	}
	i.Description, err = chooseFieldVersion(i.Description, o.Description)
	if err != nil {
		return
	}
	i.Language, err = chooseFieldVersion(i.Language, o.Language)
	if err != nil {
		return
	}
	if i.Tags.String() != o.Tags.String() {
		index, err := h.Choose(i.Tags.String(), o.Tags.String())
		if err != nil {
			return err
		}
		if index == 1 {
			i.Tags = o.Tags
		}
	}
	if i.Series.String() != o.Series.String() {
		index, err := h.Choose(i.Series.String(), o.Series.String())
		if err != nil {
			return err
		}
		if index == 1 {
			i.Series = o.Series
		}
	}
	i.ISBN, err = chooseFieldVersion(i.ISBN, o.ISBN)
	if err != nil {
		return
	}
	// automatically fill fields usually not found in epubs.
	i.ImageURL = o.ImageURL
	i.NumPages = o.NumPages
	i.AverageRating = o.AverageRating
	return
}

func chooseFieldVersion(local, remote string) (choice string, err error) {
	if local == remote {
		return local, err
	}
	index, err := h.Choose(local, remote)
	if err != nil {
		// in case of error, return original version
		return local, err
	}
	switch index {
	case 0:
		return local, err
	case 1:
		return remote, err
	}
	return
}
