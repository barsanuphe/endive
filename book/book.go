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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	e "github.com/barsanuphe/endive/endive"
)

var validProgress = []string{"unread", "read", "reading", "shortlisted"}
var validCategories = []string{"fiction", "nonfiction"}

// Book can manipulate a book.
// A Book can have multiple epub files.
type Book struct {
	Config e.Config        `json:"-"`
	UI     e.UserInterface `json:"-"`
	BookID int             `json:"id"`
	// associated files
	RetailEpub    Epub `json:"retail"`
	NonRetailEpub Epub `json:"nonretail"`
	// metadata
	EpubMetadata Metadata `json:"epub_metadata"`
	Metadata     Metadata `json:"metadata"`
	// user info
	Progress string `json:"progress"`
	ReadDate string `json:"readdate"`
	Rating   string `json:"rating"`
	Review   string `json:"review"`
}

const (
	idField       = "id"
	progressField = "progress"
	readDateField = "readdate"
	ratingField   = "rating"
	reviewField   = "review"

	trueValue = "true"
)

// NewBook constructs a valid new Epub
func NewBook(ui e.UserInterface, id int, filename string, c e.Config, isRetail bool) *Book {
	return NewBookWithMetadata(ui, id, filename, c, isRetail, Metadata{})
}

// NewBookWithMetadata constructs a valid new Epub
func NewBookWithMetadata(ui e.UserInterface, id int, filename string, c e.Config, isRetail bool, i Metadata) *Book {
	f := Epub{Filename: filename, Config: c, UI: ui, NeedsReplacement: "false"}
	if isRetail {
		return &Book{BookID: id, RetailEpub: f, Config: c, UI: ui, EpubMetadata: i, Metadata: i, Progress: "unread"}
	}
	return &Book{BookID: id, NonRetailEpub: f, Config: c, UI: ui, EpubMetadata: i, Metadata: i, Progress: "unread"}
}

// ID returns the Books ID according to the GenericBook interface
func (b *Book) ID() int {
	return b.BookID
}

// String returns a string representation of Epub
func (b *Book) String() string {
	tags := ""
	if len(b.Metadata.Tags) != 0 {
		tags += "[ "
		for _, tag := range b.Metadata.Tags {
			tags += tag.Name + " "
		}
		tags += " ]"
	}
	return b.FullPath() + ":\t" + b.Metadata.Author() + " (" + b.Metadata.OriginalYear + ") " + b.Metadata.Title() + " [" + b.Metadata.Language + "] " + tags
}

// ShortString returns a short string representation of Epub
func (b *Book) ShortString() string {
	return b.Metadata.Author() + " (" + b.Metadata.OriginalYear + ") " + b.Metadata.Title()
}

// ShowInfo returns a table with relevant information about a book.
func (b *Book) ShowInfo(fields ...string) string {
	if len(fields) == 0 {
		// select all fields
		fields = []string{idField, "filename", authorField, titleField, yearField, editionYearField, publisherField, isbnField, descriptionField, numPagesField, languageField, categoryField, genreField, tagsField, seriesField, "versions", progressField, readDateField, averageRatingField, ratingField, reviewField}
	}
	var rows [][]string
	for _, field := range fields {
		switch field {
		case idField:
			rows = append(rows, []string{"ID", strconv.Itoa(b.BookID)})
		case "filename":
			rows = append(rows, []string{"Filename", b.MainEpub().Filename})
		case authorField:
			rows = append(rows, []string{strings.Title(authorField), b.Metadata.Author()})
		case titleField:
			rows = append(rows, []string{strings.Title(titleField), b.Metadata.Title()})
		case yearField:
			rows = append(rows, []string{"Original Publication Year", b.Metadata.OriginalYear})
		case editionYearField:
			rows = append(rows, []string{"Publication Year", b.Metadata.EditionYear})
		case publisherField:
			rows = append(rows, []string{strings.Title(publisherField), b.Metadata.Publisher})
		case isbnField:
			rows = append(rows, []string{strings.Title(isbnField), b.Metadata.ISBN})
		case descriptionField:
			rows = append(rows, []string{strings.Title(descriptionField), b.Metadata.Description})
		case numPagesField:
			if b.Metadata.NumPages != "" {
				rows = append(rows, []string{"Number of pages", b.Metadata.NumPages})
			}
		case languageField:
			rows = append(rows, []string{strings.Title(languageField), b.Metadata.Language})
		case categoryField:
			rows = append(rows, []string{strings.Title(categoryField), b.Metadata.Category})
		case genreField:
			rows = append(rows, []string{strings.Title(genreField), b.Metadata.MainGenre})
		case tagsField:
			if len(b.Metadata.Tags) != 0 {
				rows = append(rows, []string{strings.Title(tagsField), b.Metadata.Tags.String()})
			}
		case seriesField:
			if len(b.Metadata.Series) != 0 {
				rows = append(rows, []string{strings.Title(seriesField), b.Metadata.Series.String()})
			}
		case "versions":
			available := ""
			if b.HasRetail() {
				available += "retail "
				rows = append(rows, []string{"Retail hash", b.RetailEpub.Hash})
				if b.RetailEpub.NeedsReplacement == trueValue {
					rows = append(rows, []string{"Retail needs replacement", "TRUE"})
				}
			}
			if b.HasNonRetail() {
				available += "non-retail"
				rows = append(rows, []string{"Non-Retail hash", b.NonRetailEpub.Hash})
				if b.NonRetailEpub.NeedsReplacement == trueValue {
					rows = append(rows, []string{"Non-Retail needs replacement", "TRUE"})
				}
			}
			rows = append(rows, []string{"Available versions", available})
		case progressField:
			rows = append(rows, []string{strings.Title(progressField), b.Progress})
		case readDateField:
			if b.ReadDate != "" {
				rows = append(rows, []string{"Read Date", b.ReadDate})
			}
		case averageRatingField:
			if b.Metadata.AverageRating != "" {
				rows = append(rows, []string{"Average Rating", b.Metadata.AverageRating})
			}
		case ratingField:
			if b.Rating != "" {
				rows = append(rows, []string{strings.Title(ratingField), b.Rating})
			}
		case reviewField:
			if b.Review != "" {
				rows = append(rows, []string{strings.Title(reviewField), b.Review})
			}
		}
	}
	return e.TabulateRows(rows, "Info", "Book")
}

// FullPath of the main Epub of a Book.
func (b *Book) FullPath() string {
	// assuming at least one epub is defined
	return b.MainEpub().FullPath()
}

// MainEpub of a Book.
func (b *Book) MainEpub() *Epub {
	// assuming at least one epub is defined
	if b.HasRetail() {
		return &b.RetailEpub
	} else if b.HasNonRetail() {
		return &b.NonRetailEpub
	} else {
		panic(errors.New("Book has no epub file!"))
	}
}

// SetProgress sets reading progress
func (b *Book) SetProgress(progress string) (err error) {
	progress = strings.ToLower(progress)
	if _, isIn := e.StringInSlice(progress, validProgress); isIn {
		b.Progress = progress
	} else {
		err = errors.New("Unknown reading progress: " + progress)
	}
	return
}

// SetReadDate sets date when finished reading
func (b *Book) SetReadDate(date string) {
	b.ReadDate = date
}

// SetReadDateToday sets date when finished reading
func (b *Book) SetReadDateToday() {
	currentDate := time.Now().Local()
	b.SetReadDate(currentDate.Format("2006-01-02"))
}

func (b *Book) generateNewName(fileTemplate string, isRetail bool) (newName string, err error) {
	if fileTemplate == "" {
		return "", errors.New("Empty filename template")
	}

	r := strings.NewReplacer(
		"$a", "{{$a}}",
		"$t", "{{$t}}",
		"$y", "{{$y}}",
		"$l", "{{$l}}",
		"$i", "{{$i}}",
		"$s", "{{$s}}",
		"$p", "{{$p}}",
		"$c", "{{$c}}",
		"$g", "{{$g}}",
		"$r", "{{$r}}",
	)
	seriesString := ""
	if len(b.Metadata.Series) != 0 {
		seriesString = cleanPath("[" + b.Metadata.Series.String() + "]")
	}
	retail := "nonretail"
	if isRetail {
		retail = "retail"
	}
	// replace with all valid epub parameters
	tmpl := fmt.Sprintf(`{{$a := "%s"}}{{$y := "%s"}}{{$t := "%s"}}{{$l := "%s"}}{{$i := "%s"}}{{$s := "%s"}}{{$p := "%s"}}{{$c := "%s"}}{{$g := "%s"}}{{$r := "%s"}}%s`,
		cleanPath(b.Metadata.Author()), b.Metadata.OriginalYear,
		cleanPath(b.Metadata.Title()), b.Metadata.Language,
		b.Metadata.ISBN, seriesString, b.Progress, b.Metadata.Category,
		b.Metadata.MainGenre, retail, r.Replace(fileTemplate))

	var doc bytes.Buffer
	te := template.Must(template.New("hop").Parse(tmpl))
	err = te.Execute(&doc, nil)
	if err != nil {
		return
	}
	newName = strings.TrimSpace(doc.String())
	if !strings.Contains(fileTemplate, "$r") && isRetail {
		newName += " [retail]"
	}
	// adding extension
	if filepath.Ext(newName) != epubExtension {
		newName += epubExtension
	}
	// making sure the path is relative
	if strings.HasPrefix(newName, "/") {
		newName = newName[1:]
	}
	// making sure the final filename is valid
	filename := filepath.Base(newName)
	if filename != cleanPath(filename) {
		newName = filepath.Join(filepath.Dir(newName), strings.TrimSpace(cleanPath(filename)))
	}
	return
}

// RefreshEpub one specific epub associated with this Book
func (b *Book) RefreshEpub(epub Epub, isRetail bool) (wasRenamed bool, newName string, err error) {
	// do nothing if file does not exist
	if epub.Filename == "" {
		err = errors.New("Does not exist")
		return
	}
	newName, err = b.generateNewName(b.Config.EpubFilenameFormat, isRetail)
	if err != nil {
		return
	}

	if epub.Filename != newName {
		origin := epub.FullPath()
		b.UI.Info("Renaming: \n\t" + origin + "\n   =>\n\t" + newName)
		// move to c.LibraryRoot + new name
		destination := filepath.Join(b.Config.LibraryRoot, newName)
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
func (b *Book) Refresh() (wasRenamed []bool, newName []string, err error) {
	b.UI.Debug("Refreshing Epub " + b.ShortString())

	// metadata is blank, run GetMetadata
	if hasMetadata := b.Metadata.HasAny(); !hasMetadata {
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
	}
	// refresh and clean Metadata
	b.Metadata.Clean(b.Config)

	// refresh both epubs
	var wasRenamedR, wasRenamedNR bool
	var newNameR, newNameNR string
	var errR, errNR error
	if b.HasRetail() {
		_, exists := e.FileExists(b.RetailEpub.FullPath())
		if exists == nil {
			wasRenamedR, newNameR, errR = b.RefreshEpub(b.RetailEpub, true)
			if wasRenamedR {
				b.RetailEpub.Filename = newNameR
			}
		} else {
			b.UI.Warning("MISSING EPUB " + b.RetailEpub.FullPath())
			b.RetailEpub = Epub{}
		}
	}
	if b.HasNonRetail() {
		_, exists := e.FileExists(b.NonRetailEpub.FullPath())
		if exists == nil {
			wasRenamedNR, newNameNR, errNR = b.RefreshEpub(b.NonRetailEpub, false)
			if wasRenamedNR {
				b.NonRetailEpub.Filename = newNameNR
			}
		} else {
			b.UI.Warning("MISSING EPUB " + b.NonRetailEpub.FullPath())
			b.NonRetailEpub = Epub{}
		}
	}

	// preparing output
	wasRenamed = []bool{wasRenamedR, wasRenamedNR}
	newName = []string{newNameR, newNameNR}
	if errR != nil || errNR != nil {
		errorMsg := ""
		if errR != nil {
			errorMsg += errR.Error()
		}
		if errNR != nil {
			errorMsg += errNR.Error()
		}
		err = errors.New(errorMsg)
	}
	return
}

// HasRetail checks if a retail epub is available.
func (b *Book) HasRetail() (hasRetail bool) {
	return b.RetailEpub.Filename != ""
}

// HasNonRetail checks if a non-retail epub is available.
func (b *Book) HasNonRetail() (hasNonRetail bool) {
	return b.NonRetailEpub.Filename != ""
}

// HasEpub checks if the book has at least one epub
func (b *Book) HasEpub() bool {
	return b.HasRetail() || b.HasNonRetail()
}

// AddEpub to the Library
func (b *Book) AddEpub(path string, isRetail bool, hash string) (imported bool, err error) {
	if isRetail {
		if b.HasRetail() {
			b.UI.Info("Trying to import retail epub although retail version already exists.")
			if b.RetailEpub.NeedsReplacement == trueValue {
				// replace retail
				err = b.removeEpub(isRetail)
				if err != nil {
					return
				}
				imported, err = b.Import(path, isRetail, hash)
			}
		} else {
			// if no retail version exists, import
			imported, err = b.Import(path, isRetail, hash)
		}

		if imported && b.HasNonRetail() {
			// if a non-retail version existed, it is now trumped. Removing epub.
			b.UI.Warning("Non-retail version trumped, removing.")
			// replace ,nonretail
			err = b.removeEpub(!isRetail)
			if err != nil {
				return
			}
		}
	} else {
		if b.HasRetail() {
			b.UI.Info("Trying to import non-retail epub although retail version exists, ignoring.")
		} else {
			if b.HasNonRetail() {
				b.UI.Info("Trying to import non-retail epub although a non-retail version already exists.")
				if b.NonRetailEpub.NeedsReplacement == trueValue {
					// replace ,nonretail
					b.UI.Warning("Replacing non-retail version, flagged for replacement.")
					err = b.removeEpub(isRetail)
					if err != nil {
						return
					}
					imported, err = b.Import(path, isRetail, hash)
				}
			} else {
				// import non retail if no version exists
				imported, err = b.Import(path, isRetail, hash)
			}
		}
	}
	return
}

// Remove an Epub from the library
func (b *Book) removeEpub(isRetail bool) (err error) {
	if isRetail {
		// remove
		err = os.Remove(b.RetailEpub.FullPath())
		if err != nil {
			return
		}
		b.RetailEpub = Epub{}
	} else {
		// remove
		err = os.Remove(b.NonRetailEpub.FullPath())
		if err != nil {
			return
		}
		b.NonRetailEpub = Epub{}
	}
	return
}

// Import an Epub to the Library
func (b *Book) Import(path string, isRetail bool, hash string) (imported bool, err error) {
	b.UI.Debug("Importing " + path)
	// copy
	dest := filepath.Join(b.Config.LibraryRoot, filepath.Base(path))
	err = e.CopyFile(path, dest)
	if err != nil {
		return
	}
	// make epub
	ep := Epub{Filename: dest, Hash: hash, Config: b.Config, UI: b.UI}
	if isRetail {
		b.RetailEpub = ep
	} else {
		b.NonRetailEpub = ep
	}

	// get online data
	err = b.SearchOnline()
	if err != nil {
		b.UI.Debug(err.Error())
		b.UI.Warning("Could not retrieve information from GoodReads. Manual review.")
		err = b.EditField()
		if err != nil {
			b.UI.Error(err.Error())
		}
	}

	// rename
	_, _, err = b.Refresh()
	if err != nil {
		return
	}
	return true, nil
}

// Check epubs integrity.
func (b *Book) Check() (retailHasChanged bool, nonRetailHasChanged bool, err error) {
	if b.HasNonRetail() {
		nonRetailHasChanged, err = b.NonRetailEpub.Check()
		if err != nil {
			return
		}
	}
	if b.HasRetail() {
		retailHasChanged, err = b.RetailEpub.Check()
		if err != nil {
			return
		}
		if retailHasChanged {
			err = errors.New("Retail Epub hash has changed")
		}
	}
	return
}

// cleanPath makes sure a string can be used as part of a path
func cleanPath(md string) string {
	md = strings.TrimSpace(md)
	// if it starts with a dot, remove it so it does not become
	// a hidden file. if it starts with /, weird things happen.
	if strings.HasPrefix(md, ".") || strings.HasPrefix(md, "/") {
		md = md[1:]
	}
	// clean characters which would be problematic in a filename
	r := strings.NewReplacer(
		"/", "-",
		"\\", "-",
	)
	return r.Replace(md)
}

// cleanPathForVFAT makes sure a string can be used as part of a path
func cleanPathForVFAT(md string) string {
	// clean characters which would be problematic in a filename
	r := strings.NewReplacer(
		":", "-",
		"?", "",
	)
	return r.Replace(md)
}

// CleanFilename returns a filename
func (b *Book) CleanFilename() string {
	return cleanPathForVFAT(b.MainEpub().Filename)
}

// ----------------------------------------------
// FromJSON & JSON are used mainly for unit tests

// FromJSON fills the Epub info from JSON text.
func (b *Book) FromJSON(jsonBytes []byte) (err error) {
	fmt.Println("Filling Epub from DB...")
	err = json.Unmarshal(jsonBytes, b)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Loaded " + b.ShortString())
	return
}

// JSON returns a JSON representation of the Epub and its metadata.
func (b *Book) JSON() (JSONPart string, err error) {
	fmt.Println("Generationg JSON for " + b.ShortString())
	jsonEpub, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
		return
	}
	JSONPart = string(jsonEpub)
	return
}
