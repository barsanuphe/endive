/*
Package book is a subpackage of Endive, that aims to manipulate epub files and their metadata.

A Book can hold at most 2 epubs: a retail version and/or a non-retail version.

Book Metadata starts with epub Metadata, and holds additionnal information retrieved from online sources (ie, Goodreads).

The Book struct controls where the files are and how they are named.
*/
package book

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	e "github.com/barsanuphe/endive/endive"
)

const (
	idField       = "id"
	filenameField = "filename"
	progressField = "progress"
	readDateField = "readdate"
	ratingField   = "rating"
	reviewField   = "review"
	versions      = "versions"
	exportedField = "exported"
	// progress values
	unread      = "unread"
	read        = "read"
	reading     = "reading"
	shortlisted = "shortlisted"
	// category values
	fiction    = "fiction"
	nonfiction = "nonfiction"
	// type values
	essay         = "essay"
	biography     = "biography"
	autobiography = "autobiography"
	novel         = "novel"
	shortstory    = "shortstory"
	anthology     = "anthology"
	poetry        = "poetry"
)

var bookFieldMap = map[string]string{
	progressField: "Progress",
	readDateField: "ReadDate",
	ratingField:   "Rating",
	reviewField:   "Review",
	exportedField: "Exported",
}

var validProgress = []string{unread, read, reading, shortlisted}
var validCategories = []string{fiction, nonfiction}
var validTypes = []string{essay, biography, autobiography, novel, shortstory, anthology, poetry}
var allFields = []string{idField, filenameField, authorField, titleField, yearField, editionYearField, publisherField, isbnField, descriptionField, numPagesField, languageField, categoryField, typeField, genreField, tagsField, seriesField, versions, progressField, readDateField, averageRatingField, ratingField, reviewField, exportedField}

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
	Metadata Metadata `json:"metadata"`
	// user info
	Progress   string `json:"progress"`
	ReadDate   string `json:"readdate"`
	Rating     string `json:"rating"`
	Review     string `json:"review"`
	IsExported string `json:"exported"`
}

// NewBook constructs a valid new Epub
func NewBook(ui e.UserInterface, id int, filename string, c e.Config, isRetail bool) *Book {
	return NewBookWithMetadata(ui, id, filename, c, isRetail, Metadata{})
}

// NewBookWithMetadata constructs a valid new Epub
func NewBookWithMetadata(ui e.UserInterface, id int, filename string, c e.Config, isRetail bool, i Metadata) *Book {
	f := Epub{Filename: filename, Config: c, UI: ui, NeedsReplacement: e.False}
	if isRetail {
		return &Book{BookID: id, RetailEpub: f, Config: c, UI: ui, Metadata: i, Progress: "unread", IsExported: e.False}
	}
	return &Book{BookID: id, NonRetailEpub: f, Config: c, UI: ui, Metadata: i, Progress: "unread", IsExported: e.False}
}

// ID returns the Books ID according to the GenericBook interface
func (b *Book) ID() int {
	return b.BookID
}

// HasHash returns true if it is associated with an Epub with the given hash, according to the GenericBook interface
func (b *Book) HasHash(hash string) bool {
	if hash == "" {
		return false
	}
	if (b.HasRetail() && b.RetailEpub.Hash == hash) || (b.HasNonRetail() && b.NonRetailEpub.Hash == hash) {
		return true
	}
	return false
}

// LongString returns a long string representation of Epub
func (b *Book) LongString() string {
	return b.FullPath() + ":\t" + b.Metadata.Author() + " (" + b.Metadata.OriginalYear + ") " + b.Metadata.Title() + " [" + b.Metadata.Language + "] "
}

// String returns a string representation of Epub
func (b *Book) String() string {
	return b.Metadata.Author() + " (" + b.Metadata.OriginalYear + ") " + b.Metadata.Title()
}

// ShowInfo returns a table with relevant information about a book.
func (b *Book) ShowInfo(fields ...string) string {
	if len(fields) == 0 {
		// select all fields
		fields = allFields
	}
	var rows [][]string
	for _, field := range fields {
		switch field {
		case idField:
			rows = append(rows, []string{"ID", strconv.Itoa(b.BookID)})
		case filenameField:
			rows = append(rows, []string{strings.Title(filenameField), b.MainEpub().Filename})
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
		case typeField:
			rows = append(rows, []string{strings.Title(typeField), b.Metadata.Type})
		case genreField:
			rows = append(rows, []string{strings.Title(genreField), b.Metadata.Genre})
		case tagsField:
			if len(b.Metadata.Tags) != 0 {
				rows = append(rows, []string{strings.Title(tagsField), b.Metadata.Tags.String()})
			}
		case seriesField:
			if len(b.Metadata.Series) != 0 {
				rows = append(rows, []string{strings.Title(seriesField), b.Metadata.Series.String()})
			}
		case versions:
			available := ""
			if b.HasRetail() {
				available += "retail "
				rows = append(rows, []string{"Retail hash", b.RetailEpub.Hash})
				if b.RetailEpub.NeedsReplacement == e.True {
					rows = append(rows, []string{"Retail needs replacement", e.True})
				}
			}
			if b.HasNonRetail() {
				available += "non-retail"
				rows = append(rows, []string{"Non-Retail hash", b.NonRetailEpub.Hash})
				if b.NonRetailEpub.NeedsReplacement == e.True {
					rows = append(rows, []string{"Non-Retail needs replacement", e.True})
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
		case exportedField:
			if b.IsExported == e.True {
				rows = append(rows, []string{strings.Title(exportedField), e.True})
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
		b.UI.Warning("Book has no epub file!")
		return nil
	}
}

// SetExported set the main Epub as exported
func (b *Book) SetExported(isExported bool) {
	if isExported {
		b.IsExported = e.True
	} else {
		b.IsExported = e.False
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
		b.Metadata.Genre, retail, r.Replace(fileTemplate))

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
func (b *Book) RefreshEpub(epub Epub, isRetail bool) (bool, string, error) {
	// do nothing if file does not exist
	if epub.Filename == "" {
		return false, "", errors.New("Does not exist")
	}
	newName, err := b.generateNewName(b.Config.EpubFilenameFormat, isRetail)
	if err != nil {
		return false, epub.Filename, err
	}

	if epub.Filename != newName+e.EpubExtension {
		origin := epub.FullPath()
		// move to c.LibraryRoot + new name
		suffix := e.EpubExtension
		destination := ""

		uniqueNameFound := false
		isbnAdded := false
		// seed random number generator
		rand.Seed(time.Now().UTC().UnixNano())
		for !uniqueNameFound {
			destination = filepath.Join(b.Config.LibraryRoot, newName+suffix)
			_, errFileExists := e.FileExists(destination)
			if errFileExists != nil {
				uniqueNameFound = true
			} else {
				// file already exists and it's not epub.Filename
				// it belongs to another book.
				// trying to generate a unique name.

				// trying to add ISBN once to suffix, if it's not already in the filename.
				if !isbnAdded && !strings.Contains(b.Config.EpubFilenameFormat, "$i") && b.Metadata.ISBN != "" {
					suffix = "_" + b.Metadata.ISBN + e.EpubExtension
					isbnAdded = true
				} else {
					// add randint to suffix
					suffix = fmt.Sprintf("_%d%s", rand.Intn(100000), suffix)
				}
			}
		}
		// if parent directory does not exist, create
		err = os.MkdirAll(filepath.Dir(destination), os.ModePerm)
		if err != nil {
			return false, epub.Filename, err
		}
		b.UI.Info("Renaming: \n\t" + origin + "\n   =>\n\t" + newName + suffix)
		err = os.Rename(origin, destination)
		if err != nil {
			return false, epub.Filename, err
		}
		return true, newName + suffix, nil
	}
	return false, epub.Filename, nil
}

// Refresh the filenames of the Epubs associated with this Book.
func (b *Book) Refresh() (wasRenamed []bool, newName []string, err error) {
	b.UI.Debug("Refreshing Epub " + b.String())

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
			err = errors.New("Missing main epub for " + b.String())
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
		if _, exists := e.FileExists(b.RetailEpub.FullPath()); exists == nil {
			wasRenamedR, newNameR, errR = b.RefreshEpub(b.RetailEpub, true)
			if wasRenamedR {
				b.RetailEpub.Filename = newNameR
			}
		} else {
			b.UI.Warning("Missing retail epub " + b.RetailEpub.FullPath())
			b.RetailEpub = Epub{}
		}
	}
	if b.HasNonRetail() {
		if _, exists := e.FileExists(b.NonRetailEpub.FullPath()); exists == nil {
			wasRenamedNR, newNameNR, errNR = b.RefreshEpub(b.NonRetailEpub, false)
			if wasRenamedNR {
				b.NonRetailEpub.Filename = newNameNR
			}
		} else {
			b.UI.Warning("Missing nonretail epub " + b.NonRetailEpub.FullPath())
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
func (b *Book) HasRetail() bool {
	return b.RetailEpub.Filename != ""
}

// HasNonRetail checks if a non-retail epub is available.
func (b *Book) HasNonRetail() bool {
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
			if b.RetailEpub.NeedsReplacement == e.True {
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
				if b.NonRetailEpub.NeedsReplacement == e.True {
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
	// copy
	dest := filepath.Join(b.Config.LibraryRoot, filepath.Base(path))
	b.UI.Debug("Importing " + path + " to " + dest)
	err = e.CopyFile(path, dest)
	if err != nil {
		return
	}
	// make epub
	ep := Epub{Filename: filepath.Base(path), Hash: hash, Config: b.Config, UI: b.UI, NeedsReplacement: e.False}
	if isRetail {
		b.RetailEpub = ep
	} else {
		b.NonRetailEpub = ep
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
		b.UI.Error(err.Error())
		return
	}
	fmt.Println("Loaded " + b.String())
	return
}

// JSON returns a JSON representation of the Epub and its metadata.
func (b *Book) JSON() (JSONPart string, err error) {
	fmt.Println("Generationg JSON for " + b.String())
	jsonEpub, err := json.Marshal(b)
	if err != nil {
		b.UI.Error(err.Error())
		return
	}
	JSONPart = string(jsonEpub)
	return
}
