package book

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	e "github.com/barsanuphe/endive/endive"
)

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
	typeField          = "type"
	genreField         = "genre"
	numPagesField      = "numpages"
	averageRatingField = "averagerating"

	authorUsage      = "Authors can be edited as a comma-separated list of strings."
	categoryUsage    = "A book can be either fiction or nonfiction."
	typeUsage        = "The nature of this book."
	tagsUsage        = "Tags can be edited as a comma-separated list of strings."
	seriesUsage      = "Series can be edited as a comma-separated list of 'series name:index' strings. Index can be empty, or a range."
	yearUsage        = "The year in which the book was written."
	editionYearUsage = "The year in which this edition was published."
	publisherUsage   = "Publisher of this edition."
	languageUsage    = "Language of this edition."
	genreUsage       = "Main genre of this book."
	isbnUsage        = "ISBN13 for this edition."
	titleUsage       = "Title, without series information."
	descriptionUsage = "Description for this edition."

	unknownYear = "XXXX"
	unknown     = "Unknown"

	localSource  = "Epub"
	onlineSource = "Online"

	cannotSetField = "Cannot set field %s"
)

// MetadataFieldNames is a list of valid field names
var MetadataFieldNames = []string{authorField, titleField, yearField, editionYearField, publisherField, descriptionField, languageField, categoryField, typeField, genreField, tagsField, seriesField, isbnField}
var metadataFieldMap = map[string]string{
	authorField:      "Authors",
	titleField:       "BookTitle",
	yearField:        "OriginalYear",
	editionYearField: "EditionYear",
	publisherField:   "Publisher",
	descriptionField: "Description",
	languageField:    "Language",
	categoryField:    "Category",
	typeField:        "Type",
	genreField:       "Genre",
	tagsField:        "Tags",
	seriesField:      "Series",
	isbnField:        "ISBN",
}

// Metadata contains all of the known book metadata.
type Metadata struct {
	BookTitle     string   `json:"title" xml:"title"`
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
	Type          string   `json:"type"`
	Genre         string   `json:"genre"`
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
	hasType := i.Type != "" && i.Type != unknown
	hasGenre := i.Genre != "" && i.Genre != unknown
	hasISBN := i.ISBN != ""
	hasPublisher := i.Publisher != ""
	hasTags := i.Tags.String() != ""
	return hasAuthor && hasTitle && hasYear && hasLanguage && hasDescription && hasCategory && hasType && hasGenre && hasISBN && hasPublisher && hasTags
}

// Title returns Metadata's main title.
func (i *Metadata) Title() string {
	return i.BookTitle
}

// Clean cleans up the Metadata
func (i *Metadata) Clean(cfg e.Config) {
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
		for _, possibleCategory := range validCategories {
			if isIn, _ := i.Tags.Has(Tag{Name: possibleCategory}); isIn {
				i.Category = possibleCategory
				i.Tags.RemoveFromNames(possibleCategory)
				break
			}
		}
	}
	// if nothing valid found...
	if i.Category == "" {
		i.Category = unknown
	}
	if cat, err := cleanCategory(i.Category); err == nil {
		i.Category = cat
	}

	// autofill type
	if i.Type == "" {
		for _, possibleType := range validTypes {
			if isIn, _ := i.Tags.Has(Tag{Name: possibleType}); isIn {
				i.Type = possibleType
				i.Tags.RemoveFromNames(possibleType)
				break
			}
		}
	}
	// if nothing found, unknown.
	if i.Type == "" {
		i.Type = unknown
	}
	if tp, err := cleanType(i.Type); err == nil {
		i.Type = tp
	}

	// MainGenre
	if i.Genre == "" && len(i.Tags) != 0 {
		cleanName, err := cleanTagName(i.Tags[0].Name)
		if err == nil {
			i.Genre = cleanName
			i.Tags.RemoveFromNames(i.Genre)
		}
	}
	// if nothing valid found...
	if i.Genre == "" {
		i.Genre = unknown
	}
	if main, err := cleanTagName(i.Genre); err == nil {
		i.Genre = main
	}

	// clean series
	for j := range i.Series {
		i.Series[j].Name = strings.TrimSpace(i.Series[j].Name)
	}
	// clean publisher
	i.Publisher = strings.TrimSpace(i.Publisher)
	// use config aliases, again, to clean up new values for maingenre, category, etc
	i.useAliases(cfg)
}

// useAliases updates Metadata fields, using the configuration file.
func (i *Metadata) useAliases(cfg e.Config) {
	// author aliases
	for j, author := range i.Authors {
		for mainAlias, aliases := range cfg.AuthorAliases {
			_, isIn := e.StringInSlice(author, aliases)
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
			_, isIn := e.StringInSlice(tag.Name, aliases)
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
	// genre aliases (same as tags)
	for mainAlias, aliases := range cfg.TagAliases {
		_, isIn := e.StringInSlice(i.Genre, aliases)
		if isIn {
			i.Genre = mainAlias
			break
		}
	}
	// type aliases (same as tags)
	for mainAlias, aliases := range cfg.TagAliases {
		_, isIn := e.StringInSlice(i.Type, aliases)
		if isIn {
			i.Type = mainAlias
			break
		}
	}
	// publisher aliases
	for mainAlias, aliases := range cfg.PublisherAliases {
		_, isIn := e.StringInSlice(i.Publisher, aliases)
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
func (i *Metadata) Diff(o *Metadata, firstHeader, secondHeader string) string {
	var rows [][]string
	rows = append(rows, []string{i.String(), o.String()})
	rows = append(rows, []string{i.Author(), o.Author()})
	rows = append(rows, []string{i.Title(), o.Title()})
	rows = append(rows, []string{i.OriginalYear, o.OriginalYear})
	rows = append(rows, []string{i.EditionYear, o.EditionYear})
	rows = append(rows, []string{i.Publisher, o.Publisher})
	rows = append(rows, []string{i.Description, o.Description})
	rows = append(rows, []string{i.Category, o.Category})
	rows = append(rows, []string{i.Type, o.Type})
	rows = append(rows, []string{i.Genre, o.Genre})
	rows = append(rows, []string{i.Tags.String(), o.Tags.String()})
	rows = append(rows, []string{i.Series.String(), o.Series.String()})
	rows = append(rows, []string{i.Language, o.Language})
	rows = append(rows, []string{i.ISBN, o.ISBN})
	return e.TabulateRows(rows, firstHeader, secondHeader)
}

// Merge with another Metadata.
func (i *Metadata) Merge(o *Metadata, cfg e.Config, ui e.UserInterface) (err error) {
	for _, field := range MetadataFieldNames {
		err = i.MergeField(o, field, cfg, ui)
		if err != nil {
			return
		}
	}
	// automatically fill fields usually not found in epubs.
	i.ImageURL = getLargeGRUrl(o.ImageURL)
	i.NumPages = o.NumPages
	i.AverageRating = o.AverageRating
	i.Clean(cfg)
	return
}

// Set Metadata field with a string value
func (i *Metadata) Set(field, value string) error {
	structFieldName := ""
	publicFieldName := ""

	// try to find struct name from public name
	for k, v := range metadataFieldMap {
		if v == field || k == field {
			structFieldName = v
			publicFieldName = k
		}
	}
	if structFieldName == "" {
		// nothing was found, invalid field
		return errors.New("Invalid field " + field)
	}

	structField := reflect.ValueOf(i).Elem().FieldByName(structFieldName)
	if !structField.IsValid() || !structField.CanSet() {
		return fmt.Errorf(cannotSetField, field)
	}
	// set value
	switch publicFieldName {
	case tagsField:
		value = strings.ToLower(value)
		i.Tags = Tags{}
		i.Tags.AddFromNames(strings.Split(value, ",")...)
	case seriesField:
		i.Series = Series{}
		if value != "" {
			for _, s := range strings.Split(value, ",") {
				if _, err := i.Series.AddFromString(s); err != nil {
					return err
				}
			}
		}
	case authorField:
		i.Authors = strings.Split(value, ",")
		for j := range i.Authors {
			i.Authors[j] = strings.TrimSpace(i.Authors[j])
		}
	case yearField, editionYearField:
		// check it's a correct year
		_, err := strconv.Atoi(value)
		if err != nil {
			return errors.New("Invalid year value: " + value)
		}
		structField.SetString(value)
	case isbnField:
		// check it's a correct isbn
		isbn, err := e.CleanISBN(value)
		if err != nil {
			return err
		}
		structField.SetString(isbn)
	case categoryField:
		cleanCategory, err := cleanCategory(value)
		if err != nil {
			return err
		}
		structField.SetString(cleanCategory)
	case typeField:
		value = strings.ToLower(value)
		// check it's a valid type
		if _, isIn := e.StringInSlice(value, validTypes); !isIn {
			return errors.New("Invalid type: " + value)
		}
		structField.SetString(value)
	case descriptionField:
		structField.SetString(cleanHTML(value))
	case languageField:
		structField.SetString(cleanLanguage(value))
	default:
		structField.SetString(value)
	}
	return nil
}

// MergeField with another Metadata.
func (i *Metadata) MergeField(o *Metadata, field string, cfg e.Config, ui e.UserInterface) (err error) {
	userInput := ""
	options := []string{}
	switch field {
	case tagsField:
		options := append(options, i.Tags.String(), o.Tags.String())
		CleanSliceAndTagEntries(ui, i.Tags.String(), o.Tags.String(), &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), tagsUsage, options, false)
	case seriesField:
		options := append(options, i.Series.rawString(), o.Series.rawString())
		CleanSliceAndTagEntries(ui, i.Series.rawString(), o.Series.rawString(), &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), seriesUsage, options, false)
	case authorField:
		options := append(options, i.Author(), o.Author())
		CleanSliceAndTagEntries(ui, i.Author(), o.Author(), &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), authorUsage, options, false)
	case yearField:
		options := append(options, i.OriginalYear, o.OriginalYear)
		CleanSliceAndTagEntries(ui, i.OriginalYear, o.OriginalYear, &options, unknownYear)
		userInput, err = ui.SelectOption("Original Publication year", yearUsage, options, false)
	case editionYearField:
		options := append(options, i.EditionYear, o.EditionYear)
		CleanSliceAndTagEntries(ui, i.EditionYear, o.EditionYear, &options, unknownYear)
		userInput, err = ui.SelectOption("Publication year", editionYearUsage, options, false)
	case publisherField:
		options := append(options, i.Publisher, o.Publisher)
		CleanSliceAndTagEntries(ui, i.Publisher, o.Publisher, &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), publisherUsage, options, false)
	case languageField:
		options := append(options, cleanLanguage(i.Language), cleanLanguage(o.Language))
		CleanSliceAndTagEntries(ui, cleanLanguage(i.Language), cleanLanguage(o.Language), &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), languageUsage, options, false)
	case categoryField:
		options = append(options, validCategories...)
		CleanSliceAndTagEntries(ui, i.Category, o.Category, &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), categoryUsage, options, false)
	case typeField:
		options = append(options, validTypes...)
		CleanSliceAndTagEntries(ui, i.Type, o.Type, &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), typeUsage, options, false)
	case genreField:
		options := append(options, i.Genre, o.Genre)
		CleanSliceAndTagEntries(ui, i.Genre, o.Genre, &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), genreUsage, options, false)
	case isbnField:
		options := append(options, i.ISBN, o.ISBN)
		CleanSliceAndTagEntries(ui, i.ISBN, o.ISBN, &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), isbnUsage, options, false)
	case titleField:
		options := append(options, i.Title(), o.Title())
		CleanSliceAndTagEntries(ui, i.Title(), o.Title(), &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), titleUsage, options, false)
	case descriptionField:
		options := append(options, cleanHTML(i.Description), cleanHTML(o.Description))
		CleanSliceAndTagEntries(ui, cleanHTML(i.Description), cleanHTML(o.Description), &options, unknown)
		userInput, err = ui.SelectOption(strings.Title(field), descriptionUsage, options, true)
	default:
		ui.Debug("Unknown field: " + field)
		return errors.New("Unknown field: " + field)
	}

	// checking SelectOption err
	if err != nil {
		return
	}
	// set the field
	err = i.Set(field, userInput)
	if err != nil {
		return
	}
	i.Clean(cfg)
	return
}

// getOnlineMetadata retrieves the online info for this book.
func (i *Metadata) getOnlineMetadata(ui e.UserInterface, cfg e.Config) (*Metadata, error) {
	if cfg.GoodReadsAPIKey == "" {
		return nil, e.WarningGoodReadsAPIKeyMissing
	}
	var err error
	var g RemoteLibraryAPI
	g = GoodReads{}
	id := ""

	// If not ISBN is found, ask for input
	if i.ISBN == "" {
		ui.Warning("Could not find ISBN.")
		isbn, err := e.AskForISBN(ui)
		if err == nil {
			i.ISBN = isbn
		}
	}
	// search by ISBN preferably
	if i.ISBN != "" {
		id, err = g.GetBookIDByISBN(i.ISBN, cfg.GoodReadsAPIKey)
		if err != nil {
			return nil, err
		}
	}
	// if no ISBN or nothing was found
	if id == "" {
		// TODO: if unsure, show hits
		id, err = g.GetBookIDByQuery(i.Author(), i.Title(), cfg.GoodReadsAPIKey)
		if err != nil {
			return nil, err
		}
	}
	// if still nothing was found...
	if id == "" {
		return nil, errors.New("Could not find online data for " + i.String())
	}
	// get book info
	onlineInfo, err := g.GetBook(id, cfg.GoodReadsAPIKey)
	if err == nil {
		onlineInfo.Clean(cfg)
	}
	return &onlineInfo, nil
}

// SearchOnline tries to find metadata from online sources.
func (i *Metadata) SearchOnline(ui e.UserInterface, cfg e.Config, fields ...string) (err error) {
	onlineInfo, err := i.getOnlineMetadata(ui, cfg)
	if err != nil {
		ui.Debug(err.Error())
		ui.Warning("Could not retrieve information from GoodReads. Manual review.")
		err = i.Merge(&Metadata{}, cfg, ui)
		if err != nil {
			ui.Error(err.Error())
		}
		return err
	}

	// show diff between epub and GR versions, then ask what to do.
	fmt.Println(i.Diff(onlineInfo, localSource, onlineSource))
	ui.Choice("[E]dit or [A]bort : ")
	validChoice := false
	errs := 0
	for !validChoice {
		choice, scanErr := ui.GetInput()
		if scanErr != nil {
			return scanErr
		}
		switch strings.ToLower(choice) {
		case "a":
			err = errors.New("Abort")
			validChoice = true
		case "e":
			if len(fields) == 0 {
				if err := i.Merge(onlineInfo, cfg, ui); err != nil {
					return err
				}
			} else {
				for _, f := range fields {
					if err := i.MergeField(onlineInfo, f, cfg, ui); err != nil {
						return err
					}
				}
				// automatically fill fields usually not found in epubs.
				i.ImageURL = getLargeGRUrl(onlineInfo.ImageURL)
				i.NumPages = onlineInfo.NumPages
				i.AverageRating = onlineInfo.AverageRating
				i.Clean(cfg)
			}
			validChoice = true
		default:
			fmt.Println("Invalid choice.")
			errs++
			if errs > 10 {
				return errors.New("Too many invalid choices.")
			}
		}
	}
	return
}
