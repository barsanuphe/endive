package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var validProgress = []string{"unread", "read", "reading", "shortlisted"}

// Book can manipulate a book.
// A Book can have multiple epub files.
type Book struct {
	Config Config `json:"-"`

	RetailEpub    Epub    `json:"retail"`
	NonRetailEpub    Epub    `json:"nonretail"`
	Metadata Metadata `json:"metadata"`
	Series   Series   `json:"series"`
	Tags     []string `json:"tags"`

	Progress    string `json:"progress"`
	ReadDate    string `json:"readdate"`
	Rating      string `json:"rating"`
	Review      string `json:"review"`
	Description string `json:"description"`
}

// NewBook constucts a valid new Epub
func NewBook(filename string, c Config, isRetail bool) *Book {
	m := NewMetadata()
	if isRetail {
		f := Epub{Filename: filename, Config: c, NeedsReplacement: "false", Retail: "true"}
		return &Book{RetailEpub:f, Config: c, Metadata: *m, Progress: "unread"}
	} else {
		f := Epub{Filename: filename, Config: c, NeedsReplacement: "false", Retail: "false"}
		return &Book{NonRetailEpub:f, Config: c, Metadata: *m, Progress: "unread"}
	}
}

// ShortString returns a short string representation of Epub
func (e *Book) ShortString() (desc string) {
	return e.Metadata.Get("author")[0] + " (" + e.Metadata.Get("year")[0] + ") " + e.Metadata.Get("title")[0]
}

func (e *Book) getMainFilename() (filename string ) {
	// assuming at least one epub is defined
	if e.RetailEpub.Filename == "" && e.NonRetailEpub.Filename != "" {
		return e.NonRetailEpub.Filename
	}
	if e.RetailEpub.Filename != "" && e.NonRetailEpub.Filename == "" {
		return e.RetailEpub.Filename
	}
	// TODO return err
	return "ERROR"
}

// String returns a string representation of Epub
func (e *Book) String() (desc string) {
	tags := ""
	if len(e.Tags) != 0 {
		tags = "[ " + strings.Join(e.Tags, " ") + " ]"
	}
	return e.getMainFilename() + ":\t" + e.Metadata.Get("creator")[0] + " (" + e.Metadata.Get("year")[0] + ") " + e.Metadata.Get("title")[0] + " [" + e.Metadata.Get("language")[0] + "] " + tags
}

// SetProgress sets reading progress
func (e *Book) SetProgress(progress string) (err error) {
	progress = strings.ToLower(progress)
	if _, isIn := stringInSlice(progress, validProgress); isIn {
		e.Progress = progress
	} else {
		err = errors.New("Unknown reading progress: " + progress)
	}
	return
}

// AddTag adds a tag
func (e *Book) AddTag(tagName string) (err error) {
	_, isIn := stringInSlice(tagName, e.Tags)
	if !isIn {
		e.Tags = append(e.Tags, tagName)
	}
	return
}

// RemoveTag removes a series
func (e *Book) RemoveTag(tagName string) (err error) {
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
func (e *Book) HasTag(tagName string) (hasThisTag bool) {
	_, hasThisTag = stringInSlice(tagName, e.Tags)
	return
}

// SetReadDate sets date when finished reading
func (e *Book) SetReadDate(date string) (err error) {
	e.ReadDate = date
	return
}

// SetReadDateToday sets date when finished reading
func (e *Book) SetReadDateToday() (err error) {
	currentDate := time.Now().Local()
	return e.SetReadDate(currentDate.Format("2006-01-02"))
}

func (e *Book) generateNewName(fileTemplate string, isRetail bool) (newName string, err error) {
	// TODO add all replacements
	r := strings.NewReplacer(
		"$a", "{{$a}}",
		"$t", "{{$t}}",
		"$y", "{{$y}}",
		"$l", "{{$l}}",
	)

	// replace with all valid epub parameters
	tmpl := fmt.Sprintf(`{{$a := "%s"}}{{$y := "%s"}}{{$t := "%s"}}{{$l := "%s"}}%s`,
		cleanForPath(e.Metadata.Get("author")[0]),
		             e.Metadata.Get("year")[0],
		cleanForPath(e.Metadata.Get("title")[0]), e.Metadata.Get("language")[0], r.Replace(fileTemplate))

	var doc bytes.Buffer
	te := template.Must(template.New("hop").Parse(tmpl))
	err = te.Execute(&doc, nil)
	if err != nil {
		return
	}
	newName = doc.String()
	if isRetail {
		newName += " [retail]"
	}
	// adding extension
	if filepath.Ext(newName) != ".epub" {
		newName += ".epub"
	}
	return
}

// Refresh filename.
func (e *Book) Refresh() (wasRenamed bool, newName string, err error) {
	fmt.Println("Refreshing Epub " + e.ShortString())
	// loop over files
	for _, epub := range []Epub{e.RetailEpub, e.NonRetailEpub} {
		fmt.Println("...  " + epub.Filename)
		// metadata is blank, run GetMetadata
		if hasMetadata := e.Metadata.HasAny(); !hasMetadata {
			// TODO read from retail only if available
			err = e.Metadata.Read(epub.Filename)
			if err != nil {
				return
			}
		}

		newName, err = e.generateNewName(e.Config.EpubFilenameFormat, epub.IsRetail())
		if err != nil {
			return
		}

		if epub.Filename != newName {
			fmt.Println("Renaming to: " + newName)
			// move to c.LibraryRoot + new name
			origin := epub.getPath()
			destination := filepath.Join(e.Config.LibraryRoot, newName)
			err = os.Rename(origin, destination)
			if err != nil {
				return
			}
			wasRenamed = true
			epub.Filename = newName
		}
	}
	return
}


// Import an Epub to the Library
func (e *Book) Import(isRetail bool) (err error) {
	// TODO try to find similar book in library
	// if found, see if duplicate or not
	// if it trumps, replace old one
	// if not, merge both Books (keep retail Book, add non-retail Epub)

	// TODO tests
/*
	if e.Hash == "" {
		err = e.GetHash()
		if err != nil {
			return
		}
	}
	// TODO check
	if !e.HasMetadata() {
		err = e.GetMetadata()
		if err != nil {
			return
		}
	}
	// get newName
	newName, err := e.generateNewName(e.Config.EpubFilenameFormat, isRetail)
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
*/
	return
}


// IsDuplicate checks if current objet is duplicate of another
func (e *Book) IsDuplicate(o Book, isRetail bool) (isDupe bool, trumps bool) {
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

// FromJSON & JSON are used mainly for unit tests

// FromJSON fills the Epub info from JSON text.
func (e *Book) FromJSON(jsonBytes []byte) (err error) {
	fmt.Println("Filling Epub from DB for " + e.ShortString())
	err = json.Unmarshal(jsonBytes, e)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// JSON returns a JSON representation of the Epub and its metadata.
func (e *Book) JSON() (JSONPart string, err error) {
	fmt.Println("Generationg JSON for " + e.ShortString())
	jsonEpub, err := json.Marshal(e)
	if err != nil {
		fmt.Println(err)
		return
	}
	JSONPart = string(jsonEpub)
	return
}
