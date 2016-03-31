package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"

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
	Name  string  `json:"name"`
	Index float32 `json:"index"`
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
func (e *Epub) AddSeries(seriesName string, index float32) (seriesModified bool) {
	hasSeries, seriesIndex, currentIndex := e.HasSeries(seriesName)
	// if not HasSeries, create new Series and add
	if !hasSeries {
		s := Series{Name: seriesName, Index: index}
		e.Series = append(e.Series, s)
		seriesModified = true
	} else {
		// if hasSeries, if index is different, update index
		if currentIndex != index {
			e.Series[seriesIndex].Index = index
			seriesModified = true
		}
	}
	return
}

// RemoveSeries removes a series
func (e *Epub) RemoveSeries(seriesName string) (seriesRemoved bool) {
	hasSeries, seriesIndex, _ := e.HasSeries(seriesName)
	if hasSeries {
		e.Series[seriesIndex] = e.Series[len(e.Series)-1]
		e.Series = e.Series[:len(e.Series)-1]
		seriesRemoved = true
	}
	return
}

// HasSeries checks if epub is part of a series
func (e *Epub) HasSeries(seriesName string) (hasSeries bool, index int, seriesIndex float32) {
	for i, series := range e.Series {
		if series.Name == seriesName {
			return true, i, series.Index
		}
	}
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
	// TODO
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

func (e *Epub) generateNewName(fileTemplate string) string {
	// add all replacements
	r := strings.NewReplacer(
		"$a", "{{$a}}",
		"$t", "{{$t}}",
		"$y", "{{$y}}",
		"$l", "{{$l}}",
	)

	// replace with all valid epub parameters
	tmpl := fmt.Sprintf(`{{$a := "%s"}}{{$y := "%d"}}{{$t := "%s"}}{{$l := "%s"}}%s`,
		e.Author, e.PublicationYear, e.Title, e.Language, r.Replace(fileTemplate))

	var doc bytes.Buffer
	te := template.Must(template.New("hop").Parse(tmpl))
	err := te.Execute(&doc, nil)
	if err != nil {
		panic(err)
	}
	return doc.String()
}

// Refresh filename.
func (e *Epub) Refresh(c Config) (wasRenamed bool, newName string, err error) {
	fmt.Println("Refreshing Epub " + e.Filename)
	// TODO the first time (ie if author, title, year are blank), run GetMetadata
	// TODO otherwise, just use the db

	// TODO isolate filename: filepath.Base(e.Filename)
	// TODO calculate new name from c.EpubFilenameFormat

	if e.Filename != newName {
		//destination := filepath.Join(c.LibraryRoot, newName)
		// TODO move to c.LibraryRoot + new name
		wasRenamed = true
		e.Filename = newName
	}

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
