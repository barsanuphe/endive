package book

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	e "github.com/barsanuphe/endive/endive"
)

const (
	authorUsage      = "Authors can be edited as a comma-separated list of strings."
	categoryUsage    = "A book can be either fiction or nonfiction."
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
)

// ForceMetadataRefresh overwrites current Metadata
func (b *Book) ForceMetadataRefresh() (err error) {
	_, exists := e.FileExists(b.MainEpub().FullPath())
	if exists == nil {
		info, ok := b.MainEpub().ReadMetadata()
		if ok != nil {
			err = ok
			return
		}
		b.Metadata = info
	} else {
		err = errors.New("Missing main epub for " + b.String())
		return
	}

	// get online data
	err = b.Metadata.SearchOnline(b.UI, b.Config)
	if err != nil {
		b.UI.Warning(err.Error())
	}
	return
}

// ForceMetadataFieldRefresh overwrites current Metadata for a specific field only.
func (b *Book) ForceMetadataFieldRefresh(field string) (err error) {
	info := Metadata{}
	_, exists := e.FileExists(b.MainEpub().FullPath())
	if exists == nil {
		info, err = b.MainEpub().ReadMetadata()
		if err != nil {
			return
		}
	} else {
		err = errors.New("Missing main epub for " + b.String())
		return
	}
	// get online data
	onlineInfo, err := b.Metadata.getOnlineMetadata(b.UI, b.Config)
	if err != nil {
		return err
	}
	// merge field
	err = info.MergeField(onlineInfo, field, b.Config, b.UI)
	if err != nil {
		return err
	}
	switch field {
	case tagsField:
		b.Metadata.Tags = info.Tags
	case seriesField:
		b.Metadata.Series = info.Series
	case authorField:
		b.Metadata.Authors = info.Authors
	case yearField:
		b.Metadata.OriginalYear = info.OriginalYear
	case editionYearField:
		b.Metadata.EditionYear = info.EditionYear
	case publisherField:
		b.Metadata.Publisher = info.Publisher
	case languageField:
		b.Metadata.Language = info.Language
	case categoryField:
		b.Metadata.Category = info.Category
	case typeField:
		b.Metadata.Type = info.Type
	case genreField:
		b.Metadata.Genre = info.Genre
	case isbnField:
		b.Metadata.ISBN = info.ISBN
	case titleField:
		b.Metadata.BookTitle = info.BookTitle
	case descriptionField:
		b.Metadata.Description = info.Description
	default:
		return errors.New("Unknown field: " + field)
	}
	return
}

// EditField in current Metadata associated with the Book.
func (b *Book) EditField(args ...string) error {
	if len(args) == 0 {
		// completely interactive edit over all fields
		atLeastOneWrong := false
		for _, field := range allFields {
			if err := b.editSpecificField(field, ""); err != nil {
				atLeastOneWrong = true
				b.UI.Warning("Could not assign new value to field " + field + ", continuing.")
			}
		}
		if atLeastOneWrong {
			return errors.New("Could not set at least one field.")
		}
	} else {
		return b.editSpecificField(strings.ToLower(args[0]), args[1])
	}
	return nil
}

func (b *Book) editSpecificField(field string, value string) error {
	switch field {
	case tagsField:
		fmt.Println(tagsUsage)
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Tags.String(), value, false)
		if err != nil {
			return err
		}
		// if user input was entered
		if strings.TrimSpace(newValue) != "" {
			// remove all tags
			b.Metadata.Tags = Tags{}
			// add new ones
			b.Metadata.Tags.AddFromNames(strings.Split(newValue, ",")...)
		}
	case seriesField:
		fmt.Println(seriesUsage)
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Series.rawString(), value, false)
		if err != nil {
			return err
		}
		// if user input was entered, we have to split the line
		if strings.TrimSpace(newValue) != "" {
			// remove all Series
			b.Metadata.Series = Series{}
			for _, s := range strings.Split(newValue, ",") {
				if _, err := b.Metadata.Series.AddFromString(s); err != nil {
					b.UI.Warning("Could not parse series " + s + ", " + err.Error())
				}
			}
		}
	case authorField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Author(), value, false)
		if err != nil {
			return err
		}
		b.Metadata.Authors = strings.Split(newValue, ",")
		// trim spaces
		for j := range b.Metadata.Authors {
			b.Metadata.Authors[j] = strings.TrimSpace(b.Metadata.Authors[j])
		}
	case yearField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.OriginalYear, value, false)
		if err != nil {
			return err
		}
		// check it's a valid date!
		_, err = strconv.Atoi(newValue)
		if err != nil {
			return err
		}
		b.Metadata.OriginalYear = newValue
	case editionYearField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.EditionYear, value, false)
		if err != nil {
			return err
		}
		// check it's a valid date!
		_, err = strconv.Atoi(newValue)
		if err != nil {
			return err
		}
		b.Metadata.EditionYear = newValue
	case languageField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Language, value, false)
		if err != nil {
			return err
		}
		b.Metadata.Language = newValue
	case categoryField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Category, value, false)
		if err != nil {
			return err
		}
		b.Metadata.Category = newValue
	case typeField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Type, value, false)
		if err != nil {
			return err
		}
		b.Metadata.Type = newValue
	case genreField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Genre, value, false)
		if err != nil {
			return err
		}
		b.Metadata.Genre = newValue
	case isbnField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.ISBN, value, false)
		if err != nil {
			return err
		}
		// check it's a valid ISBN
		isbn, err := e.CleanISBN(newValue)
		if err != nil {
			return err
		}
		b.Metadata.ISBN = isbn
	case titleField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.BookTitle, value, false)
		if err != nil {
			return err
		}
		b.Metadata.BookTitle = newValue
	case descriptionField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Description, value, true)
		if err != nil {
			return err
		}
		b.Metadata.Description = newValue
	case publisherField:
		newValue, err := b.UI.UpdateValues(field, b.Metadata.Publisher, value, false)
		if err != nil {
			return err
		}
		b.Metadata.Publisher = newValue
	case progressField:
		newValue, err := b.UI.UpdateValues(field, b.Progress, value, false)
		if err != nil {
			return err
		}
		if _, isIn := e.StringInSlice(newValue, validProgress); isIn {
			b.Progress = newValue
		} else {
			return errors.New(newValue + " is not a valid reading progress")
		}
	case readDateField:
		newValue, err := b.UI.UpdateValues(field, b.ReadDate, value, false)
		if err != nil {
			return err
		}
		// check it's a valid date
		_, err = time.Parse("2006-01-02", newValue)
		if err != nil {
			return err
		}
		b.ReadDate = newValue
	case ratingField:
		newValue, err := b.UI.UpdateValues(field, b.Rating, value, false)
		if err != nil {
			return err
		}
		// checking rating is between 0 and 10
		val, err := strconv.ParseFloat(newValue, 32)
		if err != nil || val > 5 || val < 0 {
			b.UI.Error("Rating must be between 0 and 5.")
			return err
		}
		b.Rating = newValue
	case reviewField:
		newValue, err := b.UI.UpdateValues(field, b.Review, value, true)
		if err != nil {
			return err
		}
		b.Review = newValue
	default:
		b.UI.Debug("Unknown field: " + field)
		return errors.New("Unknown field: " + field)
	}
	// cleaning all metadata
	b.Metadata.Clean(b.Config)
	return nil
}
