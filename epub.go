package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/barsanuphe/epubgo"
)

// Series can track a series and an epub's position.
type Series struct {
	Name  string
	Index int
}

// Epub can manipulate an epub file.
// TODO map directly to JSON
type Epub struct {
	Filename        string
	RelativePath    string
	NewFilename     string
	NewRelativePath string
	Hash            string
	IsRetail        bool
	Progress        int
	Series          []Series
	Author          string
	Title           string
	Language	string
	PublicationYear int
	ReadDate        string // month
	Tags            []string
	Rating          int
	Review          string
}

// String returns a string representation of Epub
func (e *Epub) String() (desc string) {
	return e.Filename + ":\t" + e.Author + " (" + strconv.Itoa(e.PublicationYear) + ") " + e.Title + " [" + e.Language + "]"
}

// GetHash calculates an epub's current hash
func (e *Epub) GetHash() (err error) {
	return
}

// SetProgress sets reading progress
func (e *Epub) SetProgress(progress int) (err error) {
	return
}

// GetProgress sets reading progress
func (e *Epub) GetProgress() (progress int, err error) {
	return
}

// AddSeries adds a series
func (e *Epub) AddSeries(seriesName string, index int) (err error) {
	return
}

// RemoveSeries removes a series
func (e *Epub) RemoveSeries(seriesName string) (err error) {
	return
}

// HasSeries checks if epub is part of a series
func (e *Epub) HasSeries(seriesName string) (isInThisSeries bool, index int, err error) {
	return
}

// AddTag adds a tag
func (e *Epub) AddTag(tagName string) (err error) {
	return
}

// RemoveTag removes a series
func (e *Epub) RemoveTag(tagName string) (err error) {
	return
}

// HasTag checks if epub is part of a series
func (e *Epub) HasTag(tagName string) (hasThisTag bool, err error) {
	return
}

// SetReadDate sets date when finished reading
func (e *Epub) SetReadDate(date string) (err error) {
	return
}

// SetReadDateToday sets date when finished reading
func (e *Epub) SetReadDateToday() (err error) {
	// TODO find today, call SetReadDate(today)
	return
}

// GetMetadata from the epub
func (e *Epub) GetMetadata() (err error) {
	fmt.Println("Reading metadata from OPF for Epub " + e.Filename)
	book, err := epubgo.Open(e.Filename)
	if err != nil {
		fmt.Println("Error parsing EPUB")
		return
	}
	defer book.Close()

	title, err := book.Metadata("title")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.Title = "Uknown"
	} else {
		e.Title = title[0]
	}

	author, err := book.Metadata("creator")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.Author = "Unknown"
	} else {
		e.Author = author[0]
	}

	language, err := book.Metadata("language")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.Language = "Unknown"
	} else {
		e.Language = language[0]
	}

	dateEvents, err := book.MetadataAttr("date")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.PublicationYear = 0
	} else {
		dates, err := book.Metadata("date")
		if err != nil {
			fmt.Println("Error parsing EPUB")
			e.PublicationYear = 0
		} else {
			// find publication event
			found := false
			for i, d := range dateEvents {
				for _, evt := range d {
					if evt == "publication" {
						found = true
						// MetadataAttr are in the same order than Metadata
						e.PublicationYear, err = strconv.Atoi(dates[i][0:4])
						if err != nil {
							panic(err)
						}
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				fmt.Println("Error parsing EPUB, no publication year")
				err = errors.New("No publication date")
			}
		}
	}
	return
}

// Refresh filename.
func (e *Epub) Refresh() (wasRenamed bool, newName string, err error) {
	fmt.Println("Refreshing Epub " + e.Filename)
	// TODO the first time (ie if author, title, year are blank), run GetMetadata
	// TODO otherwise, just use the db
	return
}

// FromJSON fills the Epub info from JSON text.
func (e *Epub) FromJSON(json string) (err error) {
	fmt.Println("Filling Epub from DB for " + e.Filename)
	return
}

// JSON returns a JSON representation of the Epub and its metadata.
func (e *Epub) JSON() (JSONPart string, err error) {
	fmt.Println("Generationg JSON for " + e.Filename)
	return
}
