package book

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	e "github.com/barsanuphe/endive/endive"
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
		err = errors.New("Missing main epub for " + b.ShortString())
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
		err = errors.New("Missing main epub for " + b.ShortString())
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
	// cleaning all metadata
	b.Metadata.Clean(b.Config)
	return
}

// EditField in current Metadata associated with the Book.
func (b *Book) EditField(args ...string) (err error) {
	if len(args) == 0 {
		// completely interactive edit
		for _, field := range []string{"author", "title", "year", "edition_year", "category", "genre", "tags", "series", "language", "isbn", "description", "progress"} {
			err = b.editSpecificField(field, []string{})
			if err != nil {
				fmt.Println("Could not assign new value to field " + field + ", continuing.")
			}
		}
	} else {
		field := strings.ToLower(args[0])
		values := []string{}
		if len(args) > 1 {
			values = args[1:]
		}
		err = b.editSpecificField(field, values)
	}
	return
}

func (b *Book) editSpecificField(field string, values []string) error {
	switch field {
	case tagsField:
		fmt.Println("NOTE: tags can be edited as a comma-separated list of strings.")
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Tags.String(), values, false)
		if err != nil {
			return err
		}
		// if user input was entered, we have to split the line
		if len(newValues) == 1 {
			newValues = strings.Split(newValues[0], ",")
		}
		for i := range newValues {
			newValues[i] = strings.TrimSpace(newValues[i])
		}
		// remove all tags
		b.Metadata.Tags = Tags{}
		// add new ones
		if b.Metadata.Tags.AddFromNames(newValues...) {
			b.UI.Infof("Tags added to %s\n", b.ShortString())
		}
	case seriesField:
		fmt.Println("NOTE: series can be edited as a comma-separated list of 'series name:index' strings. Index can be empty, or a range.")
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Series.rawString(), values, false)
		if err != nil {
			return err
		}
		// if user input was entered, we have to split the line
		if len(newValues) == 1 && newValues[0] != b.Metadata.Series.rawString() && strings.TrimSpace(newValues[0]) != "" {
			// remove all Series
			b.Metadata.Series = Series{}
			for _, s := range strings.Split(newValues[0], ",") {
				_, err := b.Metadata.Series.AddFromString(s)
				if err != nil {
					b.UI.Warning("Could not parse series " + s + ", " + err.Error())
				}
			}
		}
	case authorField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Author(), values, false)
		if err != nil {
			return err
		}
		b.Metadata.Authors = strings.Split(newValues[0], ",")
		// trim spaces
		for j := range b.Metadata.Authors {
			b.Metadata.Authors[j] = strings.TrimSpace(b.Metadata.Authors[j])
		}
	case yearField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.OriginalYear, values, false)
		if err != nil {
			return err
		}
		// check it's a valid date!
		_, err = strconv.Atoi(newValues[0])
		if err != nil {
			return err
		}
		b.Metadata.OriginalYear = newValues[0]
	case editionYearField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.EditionYear, values, false)
		if err != nil {
			return err
		}
		// check it's a valid date!
		_, err = strconv.Atoi(newValues[0])
		if err != nil {
			return err
		}
		b.Metadata.EditionYear = newValues[0]
	case languageField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Language, values, false)
		if err != nil {
			return err
		}
		b.Metadata.Language = newValues[0]
	case categoryField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Category, values, false)
		if err != nil {
			return err
		}
		b.Metadata.Category = newValues[0]
	case genreField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Genre, values, false)
		if err != nil {
			return err
		}
		b.Metadata.Genre = newValues[0]
	case isbnField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.ISBN, values, false)
		if err != nil {
			return err
		}
		// check it's a valid ISBN
		isbn, err := e.CleanISBN(newValues[0])
		if err != nil {
			return err
		}
		b.Metadata.ISBN = isbn
	case titleField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.BookTitle, values, false)
		if err != nil {
			return err
		}
		b.Metadata.BookTitle = newValues[0]
	case descriptionField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Description, values, true)
		if err != nil {
			return err
		}
		b.Metadata.Description = newValues[0]
	case publisherField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Publisher, values, false)
		if err != nil {
			return err
		}
		b.Metadata.Publisher = newValues[0]
	case progressField:
		newValues, err := b.UI.UpdateValues(field, b.Progress, values, false)
		if err != nil {
			return err
		}
		if _, isIn := e.StringInSlice(newValues[0], validProgress); isIn {
			b.Progress = newValues[0]
		} else {
			return errors.New(newValues[0] + " is not a valid reading progress")
		}
	case readDateField:
		newValues, err := b.UI.UpdateValues(field, b.ReadDate, values, false)
		if err != nil {
			return err
		}
		// check it's a valid date
		_, err = time.Parse("2006-01-02", newValues[0])
		if err != nil {
			return err
		}
		b.ReadDate = newValues[0]
	case ratingField:
		newValues, err := b.UI.UpdateValues(field, b.Rating, values, false)
		if err != nil {
			return err
		}
		// checking rating is between 0 and 10
		val, err := strconv.ParseFloat(newValues[0], 32)
		if err != nil || val > 5 || val < 0 {
			b.UI.Error("Rating must be between 0 and 5.")
			return err
		}
		b.Rating = newValues[0]
	case reviewField:
		newValues, err := b.UI.UpdateValues(field, b.Review, values, true)
		if err != nil {
			return err
		}
		b.Review = newValues[0]
	default:
		b.UI.Debug("Unknown field: " + field)
		return errors.New("Unknown field: " + field)
	}
	// cleaning all metadata
	b.Metadata.Clean(b.Config)
	return nil
}
