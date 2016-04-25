/*
Package book is a subpackage of Endive, that aims to manipulate epub files and their metadata.

A Book can hold at most 2 epubs: a retail version and/or a non-retail version.

It keeps two versions of metadata:
	- EpubMetadata, which is read directly from the main epub file (retail if it exists, non-retail otherwise)
	- Metadata, which starts with EpubMetadata, but holds additionnal information retrieved from online sources (ie, Goodreads).

The Book struct controls where the files are and how they are named.
*/
package book

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
)

var validProgress = []string{"unread", "read", "reading", "shortlisted"}

// Book can manipulate a book.
// A Book can have multiple epub files.
type Book struct {
	Config cfg.Config `json:"-"`
	ID     int        `json:"id"`
	// associated files
	RetailEpub    Epub `json:"retail"`
	NonRetailEpub Epub `json:"nonretail"`
	// metadata
	EpubMetadata Info `json:"epub_metadata"`
	Metadata     Info `json:"metadata"`
	// user info
	Progress string `json:"progress"`
	ReadDate string `json:"readdate"`
	Rating   string `json:"rating"`
	Review   string `json:"review"`
}

// NewBook constucts a valid new Epub
// TODO remove
func NewBook(id int, filename string, c cfg.Config, isRetail bool) *Book {
	return NewBookWithMetadata(id, filename, c, isRetail, Info{})
}

// NewBookWithMetadata constucts a valid new Epub
func NewBookWithMetadata(id int, filename string, c cfg.Config, isRetail bool, i Info) *Book {
	f := Epub{Filename: filename, Config: c, NeedsReplacement: "false"}
	if isRetail {
		return &Book{ID: id, RetailEpub: f, Config: c, EpubMetadata: i, Metadata: i, Progress: "unread"}
	}
	return &Book{ID: id, NonRetailEpub: f, Config: c, EpubMetadata: i, Metadata: i, Progress: "unread"}
}

// ShortString returns a short string representation of Epub
func (e *Book) ShortString() (desc string) {
	return e.Metadata.Author() + " (" + e.Metadata.Year + ") " + e.Metadata.Title()
}

// String returns a string representation of Epub
func (e *Book) String() (desc string) {
	tags := ""
	if len(e.Metadata.Tags) != 0 {
		tags += "[ "
		for _, tag := range e.Metadata.Tags {
			tags += tag.Name + " "
		}
		tags += " ]"
	}
	return e.FullPath() + ":\t" + e.Metadata.Author() + " (" + e.Metadata.Year + ") " + e.Metadata.Title() + " [" + e.Metadata.Language + "] " + tags
}

// FullPath of the main Epub of a Book.
func (e *Book) FullPath() (filename string) {
	// assuming at least one epub is defined
	if e.HasRetail() {
		return e.RetailEpub.FullPath()
	} else if e.HasNonRetail() {
		return e.NonRetailEpub.FullPath()
	} else {
		panic(errors.New("Book has no epub file!"))
	}
}

// MainEpub of a Book.
func (e *Book) MainEpub() (epub *Epub) {
	// assuming at least one epub is defined
	if e.HasRetail() {
		return &e.RetailEpub
	} else if e.HasNonRetail() {
		return &e.NonRetailEpub
	} else {
		panic(errors.New("Book has no epub file!"))
	}
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
	if fileTemplate == "" {
		return "", errors.New("Empty filename template")
	}

	// TODO add all replacements
	r := strings.NewReplacer(
		"$a", "{{$a}}",
		"$t", "{{$t}}",
		"$y", "{{$y}}",
		"$l", "{{$l}}",
	)

	// replace with all valid epub parameters
	tmpl := fmt.Sprintf(`{{$a := "%s"}}{{$y := "%s"}}{{$t := "%s"}}{{$l := "%s"}}%s`,
		h.CleanForPath(e.Metadata.Author()),
		e.Metadata.Year,
		h.CleanForPath(e.Metadata.Title()), e.Metadata.Language, r.Replace(fileTemplate))

	var doc bytes.Buffer
	// NOTE: use html/template for html output
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
		origin := epub.FullPath()
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
		// FIXME: should probably test EpubMetadata then Metadata
		info, ok := e.MainEpub().ReadMetadata()
		if ok != nil {
			err = ok
			return
		}
		e.Metadata = info
	}
	// refresh Metadata
	if e.Metadata.Refresh(e.Config) {
		fmt.Println("Found author alias: " + e.Metadata.Author())
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

// HasRetail checks if a retail epub is available.
func (e *Book) HasRetail() (hasRetail bool) {
	return e.RetailEpub.Filename != ""
}

// HasNonRetail checks if a non-retail epub is available.
func (e *Book) HasNonRetail() (hasNonRetail bool) {
	return e.NonRetailEpub.Filename != ""
}

// AddEpub to the Library
func (e *Book) AddEpub(path string, isRetail bool, hash string) (imported bool, err error) {
	// TODO tests
	if isRetail {
		if e.HasRetail() {
			h.Logger.Info("Trying to import retail epub although retail version already exists.")
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

		if imported && e.HasNonRetail() {
			// if a non-retail version existed, it is now trumped. Removing epub.
			h.Logger.Warning("Non-retail version trumped, removing.")
			// replace ,nonretail
			err = e.removeEpub(!isRetail)
			if err != nil {
				return
			}
		}
	} else {
		if e.HasRetail() {
			h.Logger.Info("Trying to import non-retail epub although retail version exists, ignoring.")
		} else {
			if e.HasNonRetail() {
				h.Logger.Info("Trying to import non-retail epub although a non-retail version already exists.")
				if e.NonRetailEpub.NeedsReplacement == "true" {
					// replace ,nonretail
					h.Logger.Warning("Replacing non-retail version, flagged for replacement.")
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
	h.Logger.Debug("Importing " + path)
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

	// get online data
	err = e.SearchOnline()
	if err != nil {
		return
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
		err = os.Remove(e.RetailEpub.FullPath())
		if err != nil {
			return
		}
		e.RetailEpub = Epub{}
	} else {
		// remove
		err = os.Remove(e.NonRetailEpub.FullPath())
		if err != nil {
			return
		}
		e.NonRetailEpub = Epub{}
	}
	return
}

// Check epubs integrity.
func (e *Book) Check() (retailHasChanged bool, nonRetailHasChanged bool, err error) {
	if e.HasNonRetail() {
		nonRetailHasChanged, err = e.RetailEpub.Check()
		if err != nil {
			return
		}
	}
	if e.HasRetail() {
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

// SearchOnline tries to find metadata from online sources.
func (e *Book) SearchOnline() (err error) {
	if e.Config.GoodReadsAPIKey == "" {
		h.Logger.Error("Goodreads API key not found, not getting online information.")
		return
	}

	// TODO tests
	// TODO: if e.Metadata.ISBN exists, GetBookIDByISBN(e.Metadata.ISBN, e.Config.GoodReadsAPIKey)
	// TODO: if unsure, show hits
	id := GetBookIDByQuery(e.Metadata.Author(), e.Metadata.Title(), e.Config.GoodReadsAPIKey)
	if id == "" {
		return errors.New("Could not find online data for " + e.ShortString())
	}
	onlineInfo := GetBook(id, e.Config.GoodReadsAPIKey)
	// show diff between epub and GR versions, then ask what to do.
	fmt.Println(e.Metadata.Diff(onlineInfo, "Local", "GoodReads"))
	h.GreenBold("Accept in (B)ulk? Choose (F)ield by field? (S)earch again? (A)bort? ")

	scanner := bufio.NewReader(os.Stdin)
	choice, _ := scanner.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "a", "A", "abort":
		return errors.New("Abort")
	case "b", "B", "Bulk":
		h.Logger.Info("Accepting online version.")
		e.Metadata = onlineInfo
	case "f", "F", "Field":
		h.Logger.Info("Going through every field.")
		err = e.Metadata.Merge(onlineInfo)
		if err != nil {
			return err
		}
	case "s", "S", "Search":
		h.Logger.Info("Searching again.")
		// TODO GetBookIDByQuery but show hits instead of choosing automatically
	default:
		h.Logger.Info("What was that?")
		// TODO ask again
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
