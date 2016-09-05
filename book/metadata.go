package book

import (
	"fmt"
	"strconv"
	"strings"

	"errors"

	c "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
)

// MetadataFieldNames is a list of valid field names
var MetadataFieldNames = []string{"author", "title", "year", "edition_year", "publisher", "description", "language", "category", "genre", "tags", "series", "isbn"}

// Metadata contains all of the known book metadata.
type Metadata struct {
	ID            string   `json:"-" xml:"id"`
	MainTitle     string   `json:"title" xml:"title"`
	OriginalTitle string   `json:"original_title" xml:"work>original_title"`
	ImageURL      string   `json:"image_url" xml:"image_url"`
	NumPages      string   `json:"num_pages" xml:"num_pages"`
	Authors       []string `json:"authors" xml:"authors>author>name"`
	ISBN          string   `json:"isbn" xml:"isbn13"`
	OriginalYear  string   `json:"year" xml:"work>original_publication_year"`
	EditionYear   string   `json:"edition_year" xml:"publication_year"`
	Description   string   `json:"description" xml:"description"`
	Series        Series   `json:"series" xml:"series_works>series_work"`
	AverageRating string   `json:"average_rating" xml:"average_rating"`
	Tags          Tags     `json:"tags" xml:"popular_shelves>shelf"`
	Category      string   `json:"category"`
	MainGenre     string   `json:"main_genre"`
	Language      string   `json:"language" xml:"language_code"`
	Publisher     string   `json:"publisher" xml:"publisher"`
}

// String returns a representation of Metadata
func (i *Metadata) String() string {
	if len(i.Series) != 0 {
		return fmt.Sprintf("%s (%s) %s [%s]", i.Author(), i.OriginalYear, i.Title(), i.MainSeries().String())
	}
	return fmt.Sprintf("%s (%s) %s", i.Author(), i.OriginalYear, i.Title())
}

// HasAny checks if metadata was parsed.
func (i *Metadata) HasAny() (hasMetadata bool) {
	// if Metadata does not have a title and author, chances are it's empty.
	if i.Title() != "" && i.Author() != "" {
		return true
	}
	return
}

// IsComplete checks if metadata looks complete
func (i *Metadata) IsComplete() bool {
	hasAuthor := i.Author() != ""
	hasTitle := i.Title() != ""
	hasYear := i.OriginalYear != "" && i.OriginalYear != "XXXX"
	hasLanguage := i.Language != ""
	hasDescription := i.Description != ""
	hasCategory := i.Category != "" && i.Category != "Unknown"
	hasGenre := i.MainGenre != "" && i.MainGenre != "Unknown"
	hasISBN := i.ISBN != ""
	hasPublisher := i.Publisher != ""
	hasTags := i.Tags.String() != ""
	return hasAuthor && hasTitle && hasYear && hasLanguage && hasDescription && hasCategory && hasGenre && hasISBN && hasPublisher && hasTags
}

// Title returns Metadata's main title.
func (i *Metadata) Title() string {
	if i.OriginalTitle != "" {
		return i.OriginalTitle
	}
	return i.MainTitle
}

// Clean cleans up the Metadata
func (i *Metadata) Clean(cfg c.Config) {
	// default year
	if i.OriginalYear == "" {
		if i.EditionYear != "" {
			i.OriginalYear = i.EditionYear
		} else {
			i.OriginalYear = "XXXX"
		}
	}
	if i.EditionYear == "" {
		if i.OriginalYear != "" {
			i.EditionYear = i.OriginalYear
		} else {
			i.EditionYear = "XXXX"
		}
	}
	// clean description
	i.Description = cleanHTML(i.Description)
	// clean language
	i.Language = cleanLanguage(i.Language)
	// use config aliases
	i.useAliases(cfg)
	// clean tags
	i.Tags.Clean()
	// autofill category
	if i.Category == "" {
		if isIn, _ := i.Tags.Has(Tag{Name: "fiction"}); isIn {
			i.Category = "fiction"
			i.Tags.RemoveFromNames("fiction")
		}
		if isIn, _ := i.Tags.Has(Tag{Name: "nonfiction"}); isIn {
			i.Category = "nonfiction"
			i.Tags.RemoveFromNames("nonfiction")
		}
	}
	// if nothing valid found...
	if i.Category == "" {
		i.Category = "Unknown"
	}
	if cat, err := cleanCategory(i.Category); err == nil {
		i.Category = cat
	}

	// MainGenre
	if i.MainGenre == "" && len(i.Tags) != 0 {
		cleanName, err := cleanTagName(i.Tags[0].Name)
		if err == nil {
			i.MainGenre = cleanName
			i.Tags.RemoveFromNames(i.MainGenre)
		}
	}
	// if nothing valid found...
	if i.MainGenre == "" {
		i.MainGenre = "Unknown"
	}
	if main, err := cleanTagName(i.MainGenre); err == nil {
		i.MainGenre = main
	}

	// clean series
	for j := range i.Series {
		i.Series[j].Name = strings.TrimSpace(i.Series[j].Name)
	}
	// clean publisher
	i.Publisher = strings.TrimSpace(i.Publisher)
}

// useAliases updates Metadata fields, using the configuration file.
func (i *Metadata) useAliases(cfg c.Config) (hasChanged bool) {
	// author aliases
	for j, author := range i.Authors {
		for mainAlias, aliases := range cfg.AuthorAliases {
			_, isIn := h.StringInSlice(author, aliases)
			if isIn {
				i.Authors[j] = mainAlias
				break
			}
		}
	}
	// tag aliases
	cleanTags := Tags{}
	for _, tag := range i.Tags {
		added := false
		for mainAlias, aliases := range cfg.TagAliases {
			_, isIn := h.StringInSlice(tag.Name, aliases)
			if isIn {
				cleanTags.AddFromNames(mainAlias)
				added = true
				break
			}
		}
		// if no alias found, add directly
		if !added {
			cleanTags.AddFromNames(tag.Name)
		}
	}
	i.Tags = cleanTags

	// publisher aliases
	for mainAlias, aliases := range cfg.PublisherAliases {
		_, isIn := h.StringInSlice(i.Publisher, aliases)
		if isIn {
			i.Publisher = mainAlias
			break
		}
	}
	return
}

// Author returns Metadata's main author.
func (i *Metadata) Author() (author string) {
	author = "Unknown"
	if len(i.Authors) != 0 {
		author = strings.Join(i.Authors, ", ")
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
	rows = append(rows, []string{i.OriginalYear, o.OriginalYear})
	rows = append(rows, []string{i.EditionYear, o.EditionYear})
	rows = append(rows, []string{i.Publisher, o.Publisher})
	rows = append(rows, []string{i.Description, o.Description})
	rows = append(rows, []string{i.Category, o.Category})
	rows = append(rows, []string{i.MainGenre, o.MainGenre})
	rows = append(rows, []string{i.Tags.String(), o.Tags.String()})
	rows = append(rows, []string{i.Series.String(), o.Series.String()})
	rows = append(rows, []string{i.Language, o.Language})
	rows = append(rows, []string{i.ISBN, o.ISBN})
	return h.TabulateRows(rows, firstHeader, secondHeader)
}

// Merge with another Metadata.
func (i *Metadata) Merge(o Metadata, cfg c.Config) (err error) {
	for _, field := range MetadataFieldNames {
		err = i.MergeField(o, field, cfg)
		if err != nil {
			return
		}
	}
	// automatically fill fields usually not found in epubs.
	i.ImageURL = o.ImageURL
	i.NumPages = o.NumPages
	i.AverageRating = o.AverageRating
	i.Clean(cfg)
	return
}

// MergeField with another Metadata.
func (i *Metadata) MergeField(o Metadata, field string, cfg c.Config) (err error) {
	switch field {
	case "tags", "tag":
		h.Subpart("Tags: ")
		fmt.Println("NOTE: tags can be edited as a comma-separated list of strings.")
		tagString, e := h.Choose(i.Tags.String(), o.Tags.String())
		if e != nil {
			return e
		}
		i.Tags = Tags{}
		i.Tags.AddFromNames(strings.Split(tagString, ",")...)
	case "series":
		h.Subpart("Series: ")
		fmt.Println("NOTE: series can be edited as a comma-separated list of 'series name:index' strings. Index can be empty.")
		userInput, e := h.Choose(i.Series.rawString(), o.Series.rawString())
		if e != nil {
			return e
		}
		if strings.TrimSpace(userInput) != "" {
			i.Series = Series{}
			for _, s := range strings.Split(userInput, ",") {
				// split again name:index
				parts := strings.Split(s, ":")
				switch len(parts) {
				case 1:
					i.Series.Add(strings.TrimSpace(s), 0)
				case 2:
					index, e := strconv.ParseFloat(parts[1], 32)
					if e != nil {
						h.Warning("Index must be a float, or empty.")
					} else {
						i.Series.Add(strings.TrimSpace(parts[0]), float32(index))
					}
				default:
					h.Warning("Could not parse series " + s)
				}
			}
		}
	case "author", "authors":
		h.Subpart("Authors: ")
		fmt.Println("NOTE: authors can be edited as a comma-separated list of strings.")
		userInput, e := h.Choose(i.Author(), o.Author())
		if e != nil {
			return e
		}
		i.Authors = strings.Split(userInput, ",")
		// trim spaces
		for j := range i.Authors {
			i.Authors[j] = strings.TrimSpace(i.Authors[j])
		}
	case "year":
		i.OriginalYear, err = h.ChooseVersion("Original Publication year", i.OriginalYear, o.OriginalYear)
		if err != nil {
			return
		}
	case "edition_year":
		i.EditionYear, err = h.ChooseVersion("Publication year", i.EditionYear, o.EditionYear)
		if err != nil {
			return
		}
	case "publisher":
		i.Publisher, err = h.ChooseVersion("Publisher", i.Publisher, o.Publisher)
		if err != nil {
			return
		}
	case "language":
		i.Language, err = h.ChooseVersion("Language", cleanLanguage(i.Language), cleanLanguage(o.Language))
		if err != nil {
			return
		}
	case "category":
		i.Category, err = h.ChooseVersion("Category (fiction/nonfiction)", i.Category, o.Category)
		if err != nil {
			return
		}
	case "maingenre", "main_genre", "genre":
		i.MainGenre, err = h.ChooseVersion("Main Genre", i.MainGenre, o.MainGenre)
		if err != nil {
			return
		}
	case "isbn":
		i.ISBN, err = h.ChooseVersion("ISBN", i.ISBN, o.ISBN)
		if err != nil {
			return
		}
	case "title":
		h.Subpart("Title:")
		chosenTitle, e := h.Choose(i.Title(), o.Title())
		if e != nil {
			return e
		}
		i.MainTitle = chosenTitle
		i.OriginalTitle = chosenTitle
	case "description":
		i.Description, err = h.ChooseVersion("Description", cleanHTML(i.Description), cleanHTML(o.Description))
		if err != nil {
			return
		}
	default:
		h.Debug("Unknown field: " + field)
		return errors.New("Unknown field: " + field)
	}
	i.Clean(cfg)
	return
}
