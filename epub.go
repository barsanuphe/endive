package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/barsanuphe/epubgo"
)

var validProgress = []string{"unread", "read", "reading", "shortlisted"}

// Series can track a series and an epub's position.
type Series struct {
	Name  string  `json:"name"`
	Index float32 `json:"index"`
}

// Epub can manipulate an epub file.
type Epub struct {
	Filename         string   `json:"filename"` // relative to LibraryRoot
	Config           Config   `json:"-"`
	Hash             string   `json:"hash"`
	IsRetail         string   `json:"isretail"`
	Progress         string   `json:"progress"`
	Series           []Series `json:"series"`
	Author           string   `json:"author"`
	Title            string   `json:"title"`
	Language         string   `json:"language"`
	PublicationYear  string   `json:"publicationyear"`
	ReadDate         string   `json:"readdate"`
	Tags             []string `json:"tags"`
	Rating           string   `json:"rating"`
	Review           string   `json:"review"`
	Description      string   `json:"description"`
	NeedsReplacement string   `json:"replace"`
	ISBN             string   `json:"isbn"`
}

// NewEpub constucts a valid new Epub
func NewEpub(filename string, c Config) *Epub {
	return &Epub{Filename: filename, Config: c, NeedsReplacement: "false", IsRetail: "false", Progress: "unread"}
}

// getPath returns the absolute file path.
// if it is in the library, prepends LibraryRoot.
// if it is outside, return Filename directly.
func (e *Epub) getPath() (path string) {
	// TODO: tests
	if filepath.IsAbs(path) {
		return e.Filename
	} else {
		return filepath.Join(e.Config.LibraryRoot, e.Filename)
	}
}

// String returns a string representation of Epub
func (e *Epub) String() (desc string) {
	tags := ""
	if len(e.Tags) != 0 {
		tags = "[ " + strings.Join(e.Tags, " ") + " ]"
	}
	return e.Filename + ":\t" + e.Author + " (" + e.PublicationYear + ") " + e.Title + " [" + e.Language + "] " + tags
}

// GetHash calculates an epub's current hash
func (e *Epub) GetHash() (err error) {
	hash, err := calculateSHA256(e.getPath())
	if err != nil {
		return
	}
	e.Hash = hash
	return
}

// SetProgress sets reading progress
func (e *Epub) SetProgress(progress string) (err error) {
	progress = strings.ToLower(progress)
	if _, isIn := stringInSlice(progress, validProgress); isIn {
		e.Progress = progress
	} else {
		err = errors.New("Unknown reading progress: " + progress)
	}
	return
}

// FlagForReplacement an epub of insufficient quality
func (e *Epub) FlagForReplacement() (err error) {
	e.NeedsReplacement = "true"
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
	e.ReadDate = date
	return
}

// SetReadDateToday sets date when finished reading
func (e *Epub) SetReadDateToday() (err error) {
	currentDate := time.Now().Local()
	return e.SetReadDate(currentDate.Format("2006-01-02"))
}

// GetMetadata from the epub
func (e *Epub) GetMetadata() (err error) {
	fmt.Println("Reading metadata from OPF for Epub " + e.Filename)
	book, err := epubgo.Open(e.getPath())
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

	description, err := book.MetadataElement("description")
	if err == nil {
		e.Description = description[0].Content
	}
	isbn, err := book.MetadataElement("source")
	if err == nil {
		e.ISBN = isbn[0].Content
	}

	language, err := book.MetadataElement("language")
	if err == nil {
		e.Language = language[0].Content
	}

	dateEvents, err := book.MetadataElement("date")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		e.PublicationYear = "XXXX"
	} else {
		found := false
		for _, el := range dateEvents {
			for _, evt := range el.Attr {
				if evt == "publication" {
					e.PublicationYear = el.Content[0:4]
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
	// check if we have the metadata required for renaming
	// TODO ie what it used in c.EpubFilenameFormat
	if e.Author != "" && e.Title != "" {
		hasMetadata = true
	}
	return
}

func (e *Epub) generateNewName(fileTemplate string) (newName string, err error) {
	// TODO add all replacements
	r := strings.NewReplacer(
		"$a", "{{$a}}",
		"$t", "{{$t}}",
		"$y", "{{$y}}",
		"$l", "{{$l}}",
	)

	// replace with all valid epub parameters
	tmpl := fmt.Sprintf(`{{$a := "%s"}}{{$y := "%s"}}{{$t := "%s"}}{{$l := "%s"}}%s`,
		cleanForPath(e.Author), e.PublicationYear, cleanForPath(e.Title), e.Language, r.Replace(fileTemplate))

	var doc bytes.Buffer
	te := template.Must(template.New("hop").Parse(tmpl))
	err = te.Execute(&doc, nil)
	if err != nil {
		return
	}
	newName = doc.String()
	if e.IsRetail == "true" {
		newName += " [retail]"
	}
	// adding extension
	if filepath.Ext(newName) != ".epub" {
		newName += ".epub"
	}
	return
}

// Refresh filename.
func (e *Epub) Refresh() (wasRenamed bool, newName string, err error) {
	fmt.Println("Refreshing Epub " + e.Filename)
	// metadata is blank, run GetMetadata
	if hasMetadata := e.HasMetadata(); !hasMetadata {
		err = e.GetMetadata()
		if err != nil {
			return
		}
	}

	newName, err = e.generateNewName(e.Config.EpubFilenameFormat)
	if err != nil {
		return
	}

	if e.Filename != newName {
		fmt.Println("Renaming to: " + newName)
		// move to c.LibraryRoot + new name
		origin := e.getPath()
		destination := filepath.Join(e.Config.LibraryRoot, newName)
		err = os.Rename(origin, destination)
		if err != nil {
			return
		}
		wasRenamed = true
		e.Filename = newName
	}

	return
}

// SetRetail a retail epub ebook.
func (e *Epub) SetRetail() (err error) {
	// set read-only
	err = os.Chmod(e.getPath(), 0444)
	if err == nil {
		e.IsRetail = "true"
	}
	return
}

// SetNonRetail a non retail epub ebook.
func (e *Epub) SetNonRetail() (err error) {
	// set read-write
	err = os.Chmod(e.getPath(), 0777)
	if err == nil {
		e.IsRetail = "false"
	}
	return
}

// Check the retail epub integrity.
func (e *Epub) Check() (hasChanged bool, err error) {
	// get current hash
	currentHash, err := calculateSHA256(e.getPath())
	if err != nil {
		return
	}
	// compare with old
	if currentHash != e.Hash {
		hasChanged = true
		if e.IsRetail == "true" {
			return hasChanged, errors.New("Retail Epub hash has changed")
		} else {
			return hasChanged, nil
		}
	}
	return
}

// IsDuplicate checks if current objet is duplicate of another
func (e *Epub) IsDuplicate(o Epub, isRetail bool) (isDupe bool, trumps bool) {
	// TODO tests

	// TODO if isDuplicate but e.IsRetail == "true" and o.IsRetail == "false" => trumps!

	// TODO: compare isbn if both are not empty
	//if e.ISBN != "" && e.ISBN == o.ISBN {
	//	return true,
	//}
	// TODO: else compare author/title
	// TODO: also compare if retail or not
	return
}

// Import an Epub to the Library
func (e *Epub) Import(isRetail bool) (err error) {
	// TODO tests

	if e.Hash == "" {
		err = e.GetHash()
		if err != nil {
			return
		}
	}
	if !e.HasMetadata() {
		err = e.GetMetadata()
		if err != nil {
			return
		}
	}
	// get newName
	newName, err := e.generateNewName(e.Config.EpubFilenameFormat)
	if err != nil {
		return
	}
	// copy
	err = CopyFile(e.Filename, filepath.Join(e.Config.LibraryRoot, newName))
	if err != nil {
		return
	}

	// update Filename with path relative to LibraryRoot
	e.Filename = newName

	// set retail
	if isRetail {
		err = e.SetRetail()
		if err != nil {
			return
		}
	} else {
		err = e.SetNonRetail()
		if err != nil {
			return
		}
	}
	// TODO if e trumps another ebook (is retail and trumps non-retail, or trumps an ebook needing replacement)
	// TOOD remove the other version, set NeedsReplacement back to "false" if necessary

	return
}

// FromJSON & JSON are used mainly for unit tests
// TODO: see if they can be removed

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
