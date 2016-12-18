package book

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	h "github.com/barsanuphe/helpers"
)

const (
	progressUsage = "Your progress for this book: unread, shortlisted, reading or read."
	readDateUsage = "When you finished reading this book."
	ratingUsage   = "Give a rating between 0 and 5."
	reviewUsage   = "Your review of this book."
)

var usageMap = map[string]string{
	tagsField:        tagsUsage,
	seriesField:      seriesUsage,
	authorField:      authorUsage,
	yearField:        yearUsage,
	editionYearField: editionYearUsage,
	languageField:    languageUsage,
	categoryField:    categoryUsage,
	typeField:        typeUsage,
	genreField:       genreUsage,
	isbnField:        isbnUsage,
	titleField:       titleUsage,
	descriptionField: descriptionUsage,
	publisherField:   publisherUsage,
	progressField:    progressUsage,
	readDateField:    readDateUsage,
	ratingField:      ratingUsage,
	reviewField:      reviewUsage,
}

// ForceMetadataRefresh overwrites current Metadata
func (b *Book) ForceMetadataRefresh() (err error) {
	_, exists := h.FileExists(b.MainEpub().FullPath())
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
	_, exists := h.FileExists(b.MainEpub().FullPath())
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
	err = info.MergeField(onlineInfo, field, b.Config, b.UI, false)
	if err != nil {
		return err
	}
	// get field value from info
	value, err := info.Get(field)
	if err != nil {
		return err
	}
	// set value
	return b.Metadata.Set(field, value)
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

// Get Book field value
func (b *Book) Get(field string) (value string, err error) {
	var structField reflect.Value
	value, err = b.Metadata.Get(field)
	if err != nil {
		_, structField, _, err = getField(b, bookFieldMap, field)
		if err != nil {
			return "", err
		}
		value = structField.String()
	}
	return value, nil
}

// Set a field value for Book or Metadata
func (b *Book) Set(field, value string) error {
	// try to set Metadata fields first
	err := b.Metadata.Set(field, value)
	if err != nil {
		publicFieldName, structField, canBeSet, err := getField(b, bookFieldMap, field)
		if err != nil {
			return err
		}
		if !canBeSet {
			return fmt.Errorf(cannotSetField, field)
		}

		switch publicFieldName {
		case progressField:
			// check it's a valid progress
			if _, isIn := h.StringInSlice(value, validProgress); !isIn {
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
			// checking rating is between 0 and 5
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
		currentValue, err := b.Get(field)
		if err != nil {
			return err
		}
		usage, ok := usageMap[field]
		if !ok {
			usage = ""
		}
		longField := false
		if field == reviewField || field == descriptionField {
			longField = true
		}
		value, err = b.UI.UpdateValue(field, usage, currentValue, longField)
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

// OutputDiffTable returns differences between Books.
func (b *Book) OutputDiffTable(o *Book, diffOnly bool) [][]string {
	var rows [][]string
	if !diffOnly {
		rows = append(rows, []string{b.String(), o.String()})
	}
	for _, field := range allFields {
		iValue, err := b.Get(field)
		if err != nil {
			iValue = couldNotRetrieveValue
		}
		oValue, err := o.Get(field)
		if err != nil {
			oValue = couldNotRetrieveValue
		}
		if !diffOnly || iValue != oValue {
			rows = append(rows, []string{iValue, oValue})
		}
	}
	return rows
}

// AddIDToDiff returns a DiffTable with ID
func (b *Book) AddIDToDiff(rowsIn [][]string) [][]string {
	var rows [][]string
	for _, row := range rowsIn {
		newRow := append([]string{fmt.Sprintf("%d", b.ID())}, row...)
		rows = append(rows, newRow)
	}
	return rows
}
