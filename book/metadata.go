package book

import (
	"fmt"
	"strconv"
	"strings"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
)

// Metadata contains all of the known book metadata.
type Metadata struct {
	ID            string   `json:"-" xml:"id"`
	MainTitle     string   `json:"title" xml:"title"`
	OriginalTitle string   `json:"original_title" xml:"work>original_title"`
	ImageURL      string   `json:"image_url" xml:"image_url"`
	NumPages      string   `json:"num_pages" xml:"num_pages"`
	Authors       []string `json:"authors" xml:"authors>author>name"`
	ISBN          string   `json:"isbn" xml:"isbn13"`
	Year          string   `json:"year" xml:"work>original_publication_year"`
	Description   string   `json:"description" xml:"description"`
	Series        Series   `json:"series" xml:"series_works>series_work"`
	AverageRating string   `json:"average_rating" xml:"average_rating"`
	Tags          Tags     `json:"tags" xml:"popular_shelves>shelf"`
	Language      string   `json:"language" xml:"language_code"`
}

// String returns a representation of Metadata
func (i *Metadata) String() string {
	if len(i.Series) != 0 {
		return fmt.Sprintf("%s (%s) %s [%s]", i.Author(), i.Year, i.Title(), i.MainSeries().String())
	}
	return fmt.Sprintf("%s (%s) %s", i.Author(), i.Year, i.Title())
}

// HasAny checks if metadata was parsed.
func (i *Metadata) HasAny() (hasMetadata bool) {
	// if Metadata does not have a title and author, chances are it's empty.
	if i.Title() != "" && i.Author() != "" {
		return true
	}
	return
}

// Title returns Metadata's main title.
func (i *Metadata) Title() string {
	if i.OriginalTitle != "" {
		return i.OriginalTitle
	}
	return i.MainTitle
}

// Clean cleans up the Metadata
func (i *Metadata) Clean() {
	// default year
	if i.Year == "" {
		i.Year = "XXXX"
	}
	// clean description
	i.Description = cleanHTML(i.Description)
	// clean language
	i.Language = cleanLanguage(i.Language)
	// clean tags
	i.Tags.Clean()
	// clean series
	for j := range i.Series {
		i.Series[j].Name = strings.TrimSpace(i.Series[j].Name)
	}
}

// Author returns Metadata's main author.
func (i *Metadata) Author() (author string) {
	author = "Unknown"
	if len(i.Authors) != 0 {
		author = i.Authors[0]
	}
	return
}

// MainSeries return the main Series of Metadata.
func (i *Metadata) MainSeries() SingleSeries {
	if len(i.Series) != 0 {
		return i.Series[0]
	}
	return SingleSeries{}
}

// Refresh updates Metadata fields, using the configuration file.
func (i *Metadata) Refresh(c cfg.Config) (hasChanged bool) {
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

// IsSimilar checks if metadata is similar to known Metadata.
func (i *Metadata) IsSimilar(o Metadata) (isSimilar bool) {
	// TODO tests
	// check isbn
	if i.ISBN != "" && o.ISBN != "" && i.ISBN == o.ISBN {
		return true
	}
	// similar == same author/title, for now
	if i.Author() == o.Author() && i.Title() == o.Title() {
		return true
	}
	return
}

// Diff returns differences between Metadatas.
func (i *Metadata) Diff(o Metadata, firstHeader, secondHeader string) (diff string) {
	var rows [][]string
	rows = append(rows, []string{i.String(), o.String()})
	rows = append(rows, []string{i.Author(), o.Author()})
	rows = append(rows, []string{i.Title(), o.Title()})
	rows = append(rows, []string{i.Year, o.Year})
	rows = append(rows, []string{i.Description, o.Description})
	rows = append(rows, []string{i.Tags.String(), o.Tags.String()})
	rows = append(rows, []string{i.Series.String(), o.Series.String()})
	rows = append(rows, []string{i.Language, o.Language})
	rows = append(rows, []string{i.ISBN, o.ISBN})
	return h.TabulateRows(rows, firstHeader, secondHeader)
}

// Merge with another Metadata.
func (i *Metadata) Merge(o Metadata) (err error) {
	// TODO tests
	// TODO all fields
	if i.Author() != o.Author() {
		h.Subpart("Authors: ")
		index, userInput, err := h.Choose(i.Author(), o.Author())
		if err != nil {
			return err
		}
		if index == 1 {
			i.Authors = o.Authors
		}
		if index == -1 && userInput != "" {
			i.Authors = strings.Split(userInput, ",")
			// trim spaces
			for j := range i.Authors {
				i.Authors[j] = strings.TrimSpace(i.Authors[j])
			}
		}
	}
	if i.Title() != o.Title() {
		h.Subpart("Title:")
		index, userInput, err := h.Choose(i.Title(), o.Title())
		if err != nil {
			return err
		}
		if index == 1 {
			// TODO show both versions?
			i.MainTitle = o.MainTitle
			i.OriginalTitle = o.OriginalTitle
		}
		if index == -1 && userInput != "" {
			i.MainTitle = userInput
			i.OriginalTitle = userInput
		}
	}
	i.Year, err = h.ChooseVersion("Publication year", i.Year, o.Year)
	if err != nil {
		return
	}
	i.Description, err = h.ChooseVersion("Description", cleanHTML(i.Description), cleanHTML(o.Description))
	if err != nil {
		return
	}
	i.Language, err = h.ChooseVersion("Language", cleanLanguage(i.Language), cleanLanguage(o.Language))
	if err != nil {
		return
	}
	if i.Tags.String() != o.Tags.String() {
		h.Subpart("Tags: ")
		index, userInput, err := h.Choose(i.Tags.String(), o.Tags.String())
		if err != nil {
			return err
		}
		if index == 1 {
			i.Tags = o.Tags
		}
		if index == -1 && userInput != "" {
			i.Tags = Tags{}
			i.Tags.AddFromNames(strings.Split(userInput, ",")...)
			i.Tags.Clean()
		}
	}
	if i.Series.String() != o.Series.String() {
		h.Subpart("Series: ")
		index, userInput, err := h.Choose(i.Series.String(), o.Series.String())
		if err != nil {
			return err
		}
		if index == 1 {
			i.Series = o.Series
		}
		if index == -1 && userInput != "" {
			i.Series = Series{}
			for _, s := range strings.Split(userInput, ",") {
				// split again name:index
				parts := strings.Split(s, ":")
				switch len(parts) {
				case 1:
					i.Series.Add(strings.TrimSpace(s), 0)
				case 2:
					index, err := strconv.ParseFloat(parts[1], 32)
					if err != nil {
						h.Warning("Could not parse series " + s)
					} else {
						i.Series.Add(strings.TrimSpace(parts[0]), float32(index))
					}
				default:
					h.Warning("Could not parse series " + s)
				}
			}
		}
	}
	i.ISBN, err = h.ChooseVersion("ISBN", i.ISBN, o.ISBN)
	if err != nil {
		return
	}
	// automatically fill fields usually not found in epubs.
	i.ImageURL = o.ImageURL
	i.NumPages = o.NumPages
	i.AverageRating = o.AverageRating
	return
}
