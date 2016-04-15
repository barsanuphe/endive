package book

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

	"github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
)

var validProgress = []string{"unread", "read", "reading", "shortlisted"}

// Book can manipulate a book.
// A Book can have multiple epub files.
type Book struct {
	Config config.Config `json:"-"`
	ID     int           `json:"id"`

	RetailEpub    Epub     `json:"retail"`
	NonRetailEpub Epub     `json:"nonretail"`
	Metadata      Metadata `json:"metadata"`
	Series        Series   `json:"series"`
	Tags          []string `json:"tags"`

	Progress    string `json:"progress"`
	ReadDate    string `json:"readdate"`
	Rating      string `json:"rating"`
	Review      string `json:"review"`
	Description string `json:"description"`
}

// NewBook constucts a valid new Epub
func NewBook(id int, filename string, c config.Config, isRetail bool) *Book {
	return NewBookWithMetadata(id, filename, c, isRetail, NewMetadata())
}

// NewBookWithMetadata constucts a valid new Epub
func NewBookWithMetadata(id int, filename string, c config.Config, isRetail bool, m *Metadata) *Book {
	f := Epub{Filename: filename, Config: c, NeedsReplacement: "false"}
	if isRetail {
		return &Book{ID: id, RetailEpub: f, Config: c, Metadata: *m, Progress: "unread"}
	}
	return &Book{ID: id, NonRetailEpub: f, Config: c, Metadata: *m, Progress: "unread"}
}

// ShortString returns a short string representation of Epub
func (e *Book) ShortString() (desc string) {
	return e.Metadata.Get("creator")[0] + " (" + e.Metadata.Get("year")[0] + ") " + e.Metadata.Get("title")[0]
}

// String returns a string representation of Epub
func (e *Book) String() (desc string) {
	tags := ""
	if len(e.Tags) != 0 {
		tags = "[ " + strings.Join(e.Tags, " ") + " ]"
	}
	return e.GetMainFilename() + ":\t" + e.Metadata.Get("creator")[0] + " (" + e.Metadata.Get("year")[0] + ") " + e.Metadata.Get("title")[0] + " [" + e.Metadata.Get("language")[0] + "] " + tags
}

// GetMainFilename of a Book.
func (e *Book) GetMainFilename() (filename string) {
	// assuming at least one epub is defined
	if e.RetailEpub.Filename == "" && e.NonRetailEpub.Filename != "" {
		return e.NonRetailEpub.getPath()
	}
	if e.RetailEpub.Filename != "" && e.NonRetailEpub.Filename == "" {
		return e.RetailEpub.getPath()
	}
	// TODO return err
	return "ERROR"
}

// SetProgress sets reading progress
func (e *Book) SetProgress(progress string) (err error) {
	progress = strings.ToLower(progress)
	if _, isIn := h.StringInSlice(progress, validProgress); isIn {
		e.Progress = progress
	} else {
		err = errors.New("Unknown reading progress: " + progress)
	}
	return
}

// AddTag adds a tag
func (e *Book) AddTag(tagName string) (err error) {
	_, isIn := h.StringInSlice(tagName, e.Tags)
	if !isIn {
		e.Tags = append(e.Tags, tagName)
	}
	return
}

// RemoveTag removes a series
func (e *Book) RemoveTag(tagName string) (err error) {
	i, isIn := h.StringInSlice(tagName, e.Tags)
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
	_, hasThisTag = h.StringInSlice(tagName, e.Tags)
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
	if fileTemplate == "" {
		return "", errors.New("Empty filename template")
	}

	r := strings.NewReplacer(
		"$a", "{{$a}}",
		"$t", "{{$t}}",
		"$y", "{{$y}}",
		"$l", "{{$l}}",
	)

	// replace with all valid epub parameters
	tmpl := fmt.Sprintf(`{{$a := "%s"}}{{$y := "%s"}}{{$t := "%s"}}{{$l := "%s"}}%s`,
		h.CleanForPath(e.Metadata.Get("creator")[0]),
		e.Metadata.Get("year")[0],
		h.CleanForPath(e.Metadata.Get("title")[0]), e.Metadata.Get("language")[0], r.Replace(fileTemplate))

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

// refreshEpub one specific epub associated with this Book
func (e *Book) refreshEpub(epub Epub, isRetail bool) (wasRenamed bool, newName string, err error) {
	// do nothing if file does not exist
	if epub.Filename == "" {
		err = errors.New("Does not exist")
		return
	}
	newName, err = e.generateNewName(e.Config.EpubFilenameFormat, isRetail)
	if err != nil {
		return
	}

	if epub.Filename != newName {
		origin := epub.getPath()
		fmt.Println("Renaming " + origin + " to: " + newName)
		// move to c.LibraryRoot + new name

		destination := filepath.Join(e.Config.LibraryRoot, newName)
		// if parent directory does not exist, create
		err = os.MkdirAll(filepath.Dir(destination), os.ModePerm)
		if err != nil {
			return
		}
		err = os.Rename(origin, destination)
		if err != nil {
			return
		}
		wasRenamed = true
	}
	return
}

// Refresh the filenames of the Epubs associated with this Book.
func (e *Book) Refresh() (wasRenamed []bool, newName []string, err error) {
	fmt.Println("Refreshing Epub " + e.ShortString())
	// metadata is blank, run GetMetadata
	if hasMetadata := e.Metadata.HasAny(); !hasMetadata {
		err = e.Metadata.Read(e.GetMainFilename())
		if err != nil {
			return
		}
	}
	// refresh both epubs
	wasRenamedR, newNameR, errR := e.refreshEpub(e.RetailEpub, true)
	if wasRenamedR {
		e.RetailEpub.Filename = newNameR
	}
	wasRenamedNR, newNameNR, errNR := e.refreshEpub(e.NonRetailEpub, false)
	if wasRenamedNR {
		e.NonRetailEpub.Filename = newNameNR
	}
	wasRenamed = []bool{wasRenamedR, wasRenamedNR}
	newName = []string{newNameR, newNameNR}
	// TODO do better
	if errR != nil && errNR != nil {
		err = errors.New(errR.Error() + errNR.Error())
	}
	return
}

func (e *Book) hasRetail() (hasRetail bool) {
	return e.RetailEpub.Filename != ""
}

func (e *Book) hasNonRetail() (hasNonRetail bool) {
	return e.NonRetailEpub.Filename != ""
}

// AddEpub to the Library
func (e *Book) AddEpub(path string, isRetail bool, hash string) (imported bool, err error) {
	// TODO tests
	if isRetail {
		if e.hasRetail() {
			fmt.Println("Trying to import retail epub although retail version already exists.")
			if e.RetailEpub.NeedsReplacement == "true" {
				// replace retail
				err = e.removeEpub(isRetail)
				if err != nil {
					return
				}
				imported, err = e.Import(path, isRetail, hash)
			}
		} else {
			// if no retail version exists, import
			imported, err = e.Import(path, isRetail, hash)
		}

		if imported && e.hasNonRetail() {
			// if a non-retail version existed, it is now trumped. Removing epub.
			fmt.Println("Non-retail version trumped, removing.")
			// replace ,nonretail
			err = e.removeEpub(!isRetail)
			if err != nil {
				return
			}
		}
	} else {
		if e.hasRetail() {
			fmt.Println("Trying to import non-retail epub although retail version exists, ignoring.")
		} else {
			if e.hasNonRetail() {
				fmt.Println("Trying to import non-retail epub although a non-retail version already exists.")
				if e.NonRetailEpub.NeedsReplacement == "true" {
					// replace ,nonretail
					fmt.Println("Replacing non-retail version, flagged for replacement.")
					err = e.removeEpub(isRetail)
					if err != nil {
						return
					}
					imported, err = e.Import(path, isRetail, hash)
				}
			} else {
				// import non retail if no version exists
				imported, err = e.Import(path, isRetail, hash)
			}
		}
	}
	return
}

// Import an Epub to the Library
func (e *Book) Import(path string, isRetail bool, hash string) (imported bool, err error) {
	fmt.Println("Importing " + path)
	// copy
	dest := filepath.Join(e.Config.LibraryRoot, filepath.Base(path))
	err = h.CopyFile(path, dest)
	if err != nil {
		return
	}
	// make epub
	ep := Epub{Filename: dest, Hash: hash, Config: e.Config}
	if isRetail {
		e.RetailEpub = ep
	} else {
		e.NonRetailEpub = ep
	}
	// rename
	_, _, err = e.Refresh()
	if err != nil {
		return
	}
	return true, nil
}

// Remove an Epub from the library
func (e *Book) removeEpub(isRetail bool) (err error) {
	if isRetail {
		// remove
		err = os.Remove(e.RetailEpub.getPath())
		if err != nil {
			return
		}
		e.RetailEpub = Epub{}
	} else {
		// remove
		err = os.Remove(e.NonRetailEpub.getPath())
		if err != nil {
			return
		}
		e.NonRetailEpub = Epub{}
	}
	return
}

// Check epubs integrity.
func (e *Book) Check() (retailHasChanged bool, nonRetailHasChanged bool, err error) {
	if e.hasNonRetail() {
		nonRetailHasChanged, err = e.RetailEpub.Check()
		if err != nil {
			return
		}
	}
	if e.hasRetail() {
		retailHasChanged, err = e.RetailEpub.Check()
		if err != nil {
			return
		}
		if retailHasChanged {
			err = errors.New("Retail Epub hash has changed")
		}
	}
	return
}

// ----------------------------------------------
// FromJSON & JSON are used mainly for unit tests

// FromJSON fills the Epub info from JSON text.
func (e *Book) FromJSON(jsonBytes []byte) (err error) {
	fmt.Println("Filling Epub from DB...")
	err = json.Unmarshal(jsonBytes, e)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Loaded " + e.ShortString())
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
