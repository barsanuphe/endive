package book

import (
	"fmt"
	"strings"

	"errors"

	c "github.com/barsanuphe/endive/config"
	e "github.com/barsanuphe/endive/endive"
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

const (
	titleField         = "title"
	descriptionField   = "description"
	isbnField          = "isbn"
	yearField          = "year"
	editionYearField   = "edition_year"
	authorField        = "author"
	publisherField     = "publisher"
	tagsField          = "tags"
	seriesField        = "series"
	languageField      = "language"
	categoryField      = "category"
	genreField         = "genre"
	numPagesField      = "numpages"
	averageRatingField = "averagerating"
	fictionCategory    = "fiction"
	nonfictionCategory = "nonfiction"

	unknownYear = "XXXX"
	unknown     = "Unknown"

	epubExtension = ".epub"
)

// MetadataFieldNames is a list of valid field names
var MetadataFieldNames = []string{authorField, titleField, yearField, editionYearField, publisherField, descriptionField, languageField, categoryField, genreField, tagsField, seriesField, isbnField}

// String returns a representation of Metadata
func (i *Metadata) String() string {
	if len(i.Series) != 0 {
		return fmt.Sprintf("%s (%s) %s [%s]", i.Author(), i.OriginalYear, i.Title(), i.MainSeries().String())
	}
	return fmt.Sprintf("%s (%s) %s", i.Author(), i.OriginalYear, i.Title())
}

// HasAny checks if metadata was parsed.
func (i *Metadata) HasAny() bool {
	// if Metadata does not have a title and author, chances are it's empty.
	if i.Title() != "" && i.Author() != "" {
		return true
	}
	return false
}

// IsComplete checks if metadata looks complete
func (i *Metadata) IsComplete() bool {
	hasAuthor := i.Author() != ""
	hasTitle := i.Title() != ""
	hasYear := i.OriginalYear != "" && i.OriginalYear != unknownYear
	hasLanguage := i.Language != ""
	hasDescription := i.Description != ""
	hasCategory := i.Category != "" && i.Category != unknown
	hasGenre := i.MainGenre != "" && i.MainGenre != unknown
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
			i.OriginalYear = unknownYear
		}
	}
	if i.EditionYear == "" {
		if i.OriginalYear != "" {
			i.EditionYear = i.OriginalYear
		} else {
			i.EditionYear = unknownYear
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
		if isIn, _ := i.Tags.Has(Tag{Name: fictionCategory}); isIn {
			i.Category = fictionCategory
			i.Tags.RemoveFromNames(fictionCategory)
		}
		if isIn, _ := i.Tags.Has(Tag{Name: nonfictionCategory}); isIn {
			i.Category = nonfictionCategory
			i.Tags.RemoveFromNames(nonfictionCategory)
		}
	}
	// if nothing valid found...
	if i.Category == "" {
		i.Category = unknown
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
		i.MainGenre = unknown
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
func (i *Metadata) useAliases(cfg c.Config) {
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
}

// Author returns Metadata's main author.
func (i *Metadata) Author() string {
	if len(i.Authors) != 0 {
		return strings.Join(i.Authors, ", ")
	}
	return unknown
}

// MainSeries return the main Series of Metadata.
func (i *Metadata) MainSeries() SingleSeries {
	if len(i.Series) != 0 {
		return i.Series[0]
	}
	return SingleSeries{}
}

// IsSimilar checks if metadata is similar to known Metadata.
func (i *Metadata) IsSimilar(o Metadata) bool {
	// TODO tests
	// check isbn
	if i.ISBN != "" && o.ISBN != "" && i.ISBN == o.ISBN {
		return true
	}
	// similar == same author/title, for now
	if i.Author() == o.Author() && i.Title() == o.Title() {
		return true
	}
	return false
}

// Diff returns differences between Metadatas.
func (i *Metadata) Diff(o Metadata, firstHeader, secondHeader string) string {
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
func (i *Metadata) Merge(o Metadata, cfg c.Config, ui e.UserInterface) (err error) {
	for _, field := range MetadataFieldNames {
		err = i.MergeField(o, field, cfg, ui)
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
func (i *Metadata) MergeField(o Metadata, field string, cfg c.Config, ui e.UserInterface) (err error) {
	switch field {
	case tagsField:
		help := "Tags can be edited as a comma-separated list of strings."
		tagString, e := ui.Choose(strings.Title(tagsField), help, i.Tags.String(), o.Tags.String())
		if e != nil {
			return e
		}
		i.Tags = Tags{}
		i.Tags.AddFromNames(strings.Split(tagString, ",")...)
	case seriesField:
		help := "Series can be edited as a comma-separated list of 'series name:index' strings. Index can be empty, or a range."
		userInput, e := ui.Choose(strings.Title(seriesField), help, i.Series.rawString(), o.Series.rawString())
		if e != nil {
			return e
		}
		i.Series = Series{}
		userInput = strings.TrimSpace(userInput)
		if userInput != "" {
			for _, s := range strings.Split(userInput, ",") {
				_, errAdding := i.Series.AddFromString(s)
				if errAdding != nil {
					ui.Warning("Could not add series " + s + " , " + errAdding.Error())
				}
			}
		}
	case authorField:
		help := "Authors can be edited as a comma-separated list of strings."
		userInput, e := ui.Choose(strings.Title(authorField), help, i.Author(), o.Author())
		if e != nil {
			return e
		}
		i.Authors = strings.Split(userInput, ",")
		// trim spaces
		for j := range i.Authors {
			i.Authors[j] = strings.TrimSpace(i.Authors[j])
		}
	case yearField:
		i.OriginalYear, err = ui.Choose("Original Publication year", "", i.OriginalYear, o.OriginalYear)
		if err != nil {
			return
		}
	case editionYearField:
		i.EditionYear, err = ui.Choose("Publication year", "", i.EditionYear, o.EditionYear)
		if err != nil {
			return
		}
	case publisherField:
		i.Publisher, err = ui.Choose(strings.Title(publisherField), "", i.Publisher, o.Publisher)
		if err != nil {
			return
		}
	case languageField:
		i.Language, err = ui.Choose(strings.Title(languageField), "", cleanLanguage(i.Language), cleanLanguage(o.Language))
		if err != nil {
			return
		}
	case categoryField:
		i.Category, err = ui.Choose(strings.Title(categoryField), "Valid values: fiction/nonfiction.", i.Category, o.Category)
		if err != nil {
			return
		}
	case genreField:
		i.MainGenre, err = ui.Choose(strings.Title(genreField), "", i.MainGenre, o.MainGenre)
		if err != nil {
			return
		}
	case isbnField:
		i.ISBN, err = ui.Choose(strings.Title(isbnField), "ISBN13 for this epub.", i.ISBN, o.ISBN)
		if err != nil {
			return
		}
	case titleField:
		chosenTitle, e := ui.Choose(strings.Title(titleField), "Title, without series information.", i.Title(), o.Title())
		if e != nil {
			return e
		}
		i.MainTitle = chosenTitle
		i.OriginalTitle = chosenTitle
	case descriptionField:
		i.Description, err = ui.Choose(strings.Title(descriptionField), "", cleanHTML(i.Description), cleanHTML(o.Description))
		if err != nil {
			return
		}
	default:
		ui.Debug("Unknown field: " + field)
		return errors.New("Unknown field: " + field)
	}
	i.Clean(cfg)
	return
}
