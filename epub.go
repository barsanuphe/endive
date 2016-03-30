package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/barsanuphe/epubgo"
)

const (
	Unread = iota
	Read
	Reading
	Shortlisted
)

// Series can track a series and an epub's position.
type Series struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

// Epub can manipulate an epub file.
type Epub struct {
	Filename         string   `json:"filename"`
	RelativePath     string   `json:"relativepath"`
	Hash             string   `json:"hash"`
	IsRetail         bool     `json:"isretail"`
	Progress         int      `json:"progress"`
	Series           []Series `json:"series"`
	Author           string   `json:"author"`
	Title            string   `json:"title"`
	Language         string   `json:"language"`
	PublicationYear  int      `json:"publicationyear"`
	ReadDate         string   `json:"readdate"`
	Tags             []string `json:"tags"`
	Rating           int      `json:"rating"`
	Review           string   `json:"review"`
	NeedsReplacement bool     `json:"replace"`
}

// String returns a string representation of Epub
func (e *Epub) String() (desc string) {
	tags := ""
	if len(e.Tags) != 0 {
		tags = "[ " + strings.Join(e.Tags, " ") + " ]"
	}
	return e.Filename + ":\t" + e.Author + " (" + strconv.Itoa(e.PublicationYear) + ") " + e.Title + " [" + e.Language + "] " + tags
}

// GetHash calculates an epub's current hash
func (e *Epub) GetHash() (err error) {
	var result []byte
	file, err := os.Open(e.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	e.Hash = hex.EncodeToString(hash.Sum(result))
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

func stringInSlice(a string, list []string) (index int, isIn bool) {
	for i, b := range list {
		if b == a {
			return i, true
		}
	}
	return -1, false
}

// AddTag adds a tag
func (e *Epub) AddTag(tagName string) (err error) {
	_, isIn := stringInSlice(tagName, e.Tags)
	if !isIn {
		e.Tags = append(e.Tags, tagName)
	}
	return
}

// RemoveTag removes a series
func (e *Epub) RemoveTag(tagName string) (err error) {
	i, isIn := stringInSlice(tagName, e.Tags)
	if isIn {
		e.Tags[i] = e.Tags[len(e.Tags)-1]
		e.Tags = e.Tags[:len(e.Tags)-1]
	} else {
		err = errors.New(tagName + " not in tags")
	}
	return
}

// HasTag checks if epub is part of a series
func (e *Epub) HasTag(tagName string) (hasThisTag bool) {
	_, hasThisTag = stringInSlice(tagName, e.Tags)
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

	title, err := book.MetadataElement("title")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.Title = "Unknown"
	} else {
		e.Title = title[0].Content
	}

	author, err := book.MetadataElement("creator")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.Author = "Unknown"
	} else {
		e.Author = author[0].Content
	}

	language, err := book.MetadataElement("language")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.Language = "Unknown"
	} else {
		e.Language = language[0].Content
	}

	dateEvents, err := book.MetadataElement("date")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.PublicationYear = 0
	} else {
		found := false
		for _, el := range dateEvents {
			for _, evt := range el.Attr {
				if evt == "publication" {
					e.PublicationYear, err = strconv.Atoi(el.Content[0:4])
					if err != nil {
						panic(err)
					}
					found = true
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
	return
}

// HasMetadata checks if metadata was parsed
func (e *Epub) HasMetadata() (hasMetadata bool) {
	if e.Author != "" && e.Title != "" {
		hasMetadata = true
	}
	return
}

// Refresh filename.
func (e *Epub) Refresh(c Config) (wasRenamed bool, newName string, err error) {
	fmt.Println("Refreshing Epub " + e.Filename)
	// TODO the first time (ie if author, title, year are blank), run GetMetadata
	// TODO otherwise, just use the db

	// TODO isolate filename
	// TODO calculate new name from c.EpubFilenameFormat
	// TODO move to c.LibraryRoot + new name

	// TODO if old directory (c.LibraryRoot - epub filename) is empty, delete
	return
}

// FromJSON fills the Epub info from JSON text.
func (e *Epub) FromJSON(jsonBytes []byte) (err error) {
	fmt.Println("Filling Epub from DB for " + e.Filename)
	err = json.Unmarshal(jsonBytes, e)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// JSON returns a JSON representation of the Epub and its metadata.
func (e *Epub) JSON() (JSONPart string, err error) {
	fmt.Println("Generationg JSON for " + e.Filename)
	jsonEpub, err := json.Marshal(e)
	if err != nil {
		fmt.Println(err)
		return
	}
	JSONPart = string(jsonEpub)
	return
}
