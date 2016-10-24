package book

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	e "github.com/barsanuphe/endive/endive"
)

const (
	progressUsage = "Your progress for this book: unread, shortlisted, reading or read."
	readDateUsage = "When you finished reading this book."
	ratingUsage   = "Give a rating between 0 and 5."
	reviewUsage   = "Your review of this book."
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
	switch len(args) {
	case 0:
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
		return nil
	case 1:
		return b.editSpecificField(strings.ToLower(args[0]), "")
	case 2:
		return b.editSpecificField(strings.ToLower(args[0]), args[1])
	}
	return nil
}

// Set a field value for Book or Metadata
func (b *Book) Set(field, value string) error {
	// try to set Metadata fields first
	if err := b.Metadata.Set(field, value); err != nil {
		// probably failed because field was not a Metadata field
		// try to set Book fields
		structFieldName := ""
		publicFieldName := ""

		// try to find struct name from public name
		for k, v := range bookFieldMap {
			if v == field || k == field {
				structFieldName = v
				publicFieldName = k
			}
		}
		if structFieldName == "" {
			// nothing was found, invalid field
			return errors.New("Invalid field " + field)
		}
		// setting the field
		structField := reflect.ValueOf(b).Elem().FieldByName(structFieldName)
		if !structField.IsValid() || !structField.CanSet() {
			return fmt.Errorf(cannotSetField, field)
		}

		switch publicFieldName {
		case progressField:
			// check it's a valid progress
			if _, isIn := e.StringInSlice(value, validProgress); !isIn {
				return errors.New("Invalid reading progress: " + value)
			}
			structField.SetString(value)
		case readDateField:
			// check it's a valid date
			if _, err = time.Parse("2006-01-02", value); err != nil {
				return errors.New("Invalid read date: " + value)
			}
			structField.SetString(value)
		case ratingField:
			// checking rating is between 0 and 10
			val, err := strconv.ParseFloat(value, 32)
			if err != nil || val > 5 || val < 0 {
				return errors.New("Rating must be between 0 and 5.")
			}
			structField.SetString(value)
		default:
			structField.SetString(value)
		}
	}
	return nil
}

func (b *Book) editSpecificField(field string, value string) (err error) {
	if value == "" {
		switch field {
		case tagsField:
			value, err = b.UI.UpdateValue(field, tagsUsage, b.Metadata.Tags.String(), false)
		case seriesField:
			value, err = b.UI.UpdateValue(field, seriesUsage, b.Metadata.Series.rawString(), false)
		case authorField:
			value, err = b.UI.UpdateValue(field, authorUsage, b.Metadata.Author(), false)
		case yearField:
			value, err = b.UI.UpdateValue(field, yearUsage, b.Metadata.OriginalYear, false)
		case editionYearField:
			value, err = b.UI.UpdateValue(field, editionYearUsage, b.Metadata.EditionYear, false)
		case languageField:
			value, err = b.UI.UpdateValue(field, languageUsage, b.Metadata.Language, false)
		case categoryField:
			value, err = b.UI.UpdateValue(field, categoryUsage, b.Metadata.Category, false)
		case typeField:
			value, err = b.UI.UpdateValue(field, typeUsage, b.Metadata.Type, false)
		case genreField:
			value, err = b.UI.UpdateValue(field, genreUsage, b.Metadata.Genre, false)
		case isbnField:
			value, err = b.UI.UpdateValue(field, isbnUsage, b.Metadata.ISBN, false)
		case titleField:
			value, err = b.UI.UpdateValue(field, titleUsage, b.Metadata.BookTitle, false)
		case descriptionField:
			value, err = b.UI.UpdateValue(field, descriptionUsage, b.Metadata.Description, true)
		case publisherField:
			value, err = b.UI.UpdateValue(field, publisherUsage, b.Metadata.Publisher, false)
		case progressField:
			value, err = b.UI.UpdateValue(field, progressUsage, b.Progress, false)
		case readDateField:
			value, err = b.UI.UpdateValue(field, readDateUsage, b.ReadDate, false)
		case ratingField:
			value, err = b.UI.UpdateValue(field, ratingUsage, b.Rating, false)
		case reviewField:
			value, err = b.UI.UpdateValue(field, reviewUsage, b.Review, true)
		default:
			b.UI.Debug("Unknown field: " + field)
			return errors.New("Unknown field: " + field)
		}
		if err != nil {
			return err
		}
	}

	// set the field
	err = b.Set(field, value)
	if err != nil {
		return
	}

	// cleaning all metadata
	b.Metadata.Clean(b.Config)
	return
}
