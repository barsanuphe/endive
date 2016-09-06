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
	"strconv"
	"strings"
	"text/template"
	"time"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
)

var validProgress = []string{"unread", "read", "reading", "shortlisted"}
var validCategories = []string{"fiction", "nonfiction"}

// Book can manipulate a book.
// A Book can have multiple epub files.
type Book struct {
	Config cfg.Config `json:"-"`
	ID     int        `json:"id"`
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

// NewBook constructs a valid new Epub
func NewBook(id int, filename string, c cfg.Config, isRetail bool) *Book {
	return NewBookWithMetadata(id, filename, c, isRetail, Metadata{})
}

// NewBookWithMetadata constructs a valid new Epub
func NewBookWithMetadata(id int, filename string, c cfg.Config, isRetail bool, i Metadata) *Book {
	f := Epub{Filename: filename, Config: c, NeedsReplacement: "false"}
	if isRetail {
		return &Book{ID: id, RetailEpub: f, Config: c, EpubMetadata: i, Metadata: i, Progress: "unread"}
	}
	return &Book{ID: id, NonRetailEpub: f, Config: c, EpubMetadata: i, Metadata: i, Progress: "unread"}
}

// ShortString returns a short string representation of Epub
func (e *Book) ShortString() (desc string) {
	return e.Metadata.Author() + " (" + e.Metadata.OriginalYear + ") " + e.Metadata.Title()
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
	return e.FullPath() + ":\t" + e.Metadata.Author() + " (" + e.Metadata.OriginalYear + ") " + e.Metadata.Title() + " [" + e.Metadata.Language + "] " + tags
}

// ShowInfo returns a table with relevant information about a book.
func (e *Book) ShowInfo(fields ...string) (desc string) {
	if len(fields) == 0 {
		// select all fields
		fields = []string{"id", "filename", "author", "title", "year", "edition_year", "publisher", "isbn", "description", "numpages", "language", "category", "genre", "tag", "series", "versions", "progress", "readdate", "averagerating", "rating", "review"}
	}
	var rows [][]string
	for _, field := range fields {
		switch field {
		case "id":
			rows = append(rows, []string{"ID", strconv.Itoa(e.ID)})
		case "filename":
			rows = append(rows, []string{"Filename", e.MainEpub().Filename})
		case "author", "authors":
			rows = append(rows, []string{"Author", e.Metadata.Author()})
		case "title":
			rows = append(rows, []string{"Title", e.Metadata.Title()})
		case "year":
			rows = append(rows, []string{"Original Publication Year", e.Metadata.OriginalYear})
		case "edition_year":
			rows = append(rows, []string{"Publication Year", e.Metadata.EditionYear})
		case "publisher":
			rows = append(rows, []string{"Publisher", e.Metadata.Publisher})
		case "isbn":
			rows = append(rows, []string{"ISBN", e.Metadata.ISBN})
		case "description":
			rows = append(rows, []string{"Description", e.Metadata.Description})
		case "numpages":
			if e.Metadata.NumPages != "" {
				rows = append(rows, []string{"Number of pages", e.Metadata.NumPages})
			}
		case "language":
			rows = append(rows, []string{"Language", e.Metadata.Language})
		case "category":
			rows = append(rows, []string{"Category", e.Metadata.Category})
		case "genre", "maingenre":
			rows = append(rows, []string{"Main Genre", e.Metadata.MainGenre})
		case "tag", "tags":
			if len(e.Metadata.Tags) != 0 {
				rows = append(rows, []string{"Tags", e.Metadata.Tags.String()})
			}
		case "series":
			if len(e.Metadata.Series) != 0 {
				rows = append(rows, []string{"Series", e.Metadata.Series.String()})
			}
		case "versions":
			available := ""
			if e.HasRetail() {
				available += "retail "
				rows = append(rows, []string{"Retail hash", e.RetailEpub.Hash})
				if e.RetailEpub.NeedsReplacement == "true" {
					rows = append(rows, []string{"Retail needs replacement", "TRUE"})
				}
			}
			if e.HasNonRetail() {
				available += "non-retail"
				rows = append(rows, []string{"Non-Retail hash", e.NonRetailEpub.Hash})
				if e.NonRetailEpub.NeedsReplacement == "true" {
					rows = append(rows, []string{"Non-Retail needs replacement", "TRUE"})
				}
			}
			rows = append(rows, []string{"Available versions", available})
		case "progress":
			rows = append(rows, []string{"Progress", e.Progress})
		case "readdate", "read_date":
			if e.ReadDate != "" {
				rows = append(rows, []string{"Read Date", e.ReadDate})
			}
		case "averagerating":
			if e.Metadata.AverageRating != "" {
				rows = append(rows, []string{"Average Rating", e.Metadata.AverageRating})
			}
		case "rating":
			if e.Rating != "" {
				rows = append(rows, []string{"Rating", e.Rating})
			}
		case "review":
			if e.Review != "" {
				rows = append(rows, []string{"Review", e.Review})
			}
		}
	}
	return h.TabulateRows(rows, "Info", "Book")
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
	if len(e.Metadata.Series) != 0 {
		seriesString = h.CleanPath("[" + e.Metadata.Series.String() + "]")
	}
	retail := "nonretail"
	if isRetail {
		retail = "retail"
	}
	// replace with all valid epub parameters
	tmpl := fmt.Sprintf(`{{$a := "%s"}}{{$y := "%s"}}{{$t := "%s"}}{{$l := "%s"}}{{$i := "%s"}}{{$s := "%s"}}{{$p := "%s"}}{{$c := "%s"}}{{$g := "%s"}}{{$r := "%s"}}%s`,
		h.CleanPath(e.Metadata.Author()), e.Metadata.OriginalYear,
		h.CleanPath(e.Metadata.Title()), e.Metadata.Language,
		e.Metadata.ISBN, seriesString, e.Progress, e.Metadata.Category,
		e.Metadata.MainGenre, retail, r.Replace(fileTemplate))

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
	if filepath.Ext(newName) != ".epub" {
		newName += ".epub"
	}
	// making sure the path is relative
	if strings.HasPrefix(newName, "/") {
		newName = newName[1:]
	}
	// making sure the final filename is valid
	filename := filepath.Base(newName)
	if filename != h.CleanPath(filename) {
		newName = filepath.Join(filepath.Dir(newName), strings.TrimSpace(h.CleanPath(filename)))
	}
	return
}

// RefreshEpub one specific epub associated with this Book
func (e *Book) RefreshEpub(epub Epub, isRetail bool) (wasRenamed bool, newName string, err error) {
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
		h.Info("Renaming: \n\t" + origin + "\n   =>\n\t" + newName)
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

// ForceMetadataRefresh overwrites current Metadata
func (e *Book) ForceMetadataRefresh() (err error) {
	_, exists := h.FileExists(e.MainEpub().FullPath())
	if exists == nil {
		info, ok := e.MainEpub().ReadMetadata()
		if ok != nil {
			err = ok
			return
		}
		e.EpubMetadata = info
		e.Metadata = info
	} else {
		err = errors.New("Missing main epub for " + e.ShortString())
		return
	}

	// get online data
	err = e.SearchOnline()
	if err != nil {
		h.Warning(err.Error())
		return
	}
	return
}

// ForceMetadataFieldRefresh overwrites current Metadata for a specific field only.
func (e *Book) ForceMetadataFieldRefresh(field string) (err error) {
	info := Metadata{}
	_, exists := h.FileExists(e.MainEpub().FullPath())
	if exists == nil {
		info, err = e.MainEpub().ReadMetadata()
		if err != nil {
			return
		}
	} else {
		err = errors.New("Missing main epub for " + e.ShortString())
		return
	}
	// get online data
	onlineInfo, err := e.GetOnlineMetadata()
	if err != nil {
		return err
	}
	// merge field
	err = info.MergeField(onlineInfo, field, e.Config)
	if err != nil {
		return err
	}
	switch field {
	case "tags", "tag":
		e.Metadata.Tags = info.Tags
	case "series":
		e.Metadata.Series = info.Series
	case "author", "authors":
		e.Metadata.Authors = info.Authors
	case "year":
		e.Metadata.OriginalYear = info.OriginalYear
	case "edition_year":
		e.Metadata.EditionYear = info.EditionYear
	case "publisher":
		e.Metadata.Publisher = info.Publisher
	case "language":
		e.Metadata.Language = info.Language
	case "category":
		e.Metadata.Category = info.Category
	case "maingenre", "main_genre", "genre":
		e.Metadata.MainGenre = info.MainGenre
	case "isbn":
		e.Metadata.ISBN = info.ISBN
	case "title":
		e.Metadata.MainTitle = info.MainTitle
		e.Metadata.OriginalTitle = info.OriginalTitle
	case "description":
		e.Metadata.Description = info.Description
	default:
		h.Debug("Unknown field: " + field)
		return errors.New("Unknown field: " + field)
	}
	// cleaning all metadata
	e.Metadata.Clean(e.Config)
	return
}

// Refresh the filenames of the Epubs associated with this Book.
func (e *Book) Refresh() (wasRenamed []bool, newName []string, err error) {
	h.Debug("Refreshing Epub " + e.ShortString())

	// metadata is blank, run GetMetadata
	if hasMetadata := e.Metadata.HasAny(); !hasMetadata {
		_, exists := h.FileExists(e.MainEpub().FullPath())
		if exists == nil {
			info, ok := e.MainEpub().ReadMetadata()
			if ok != nil {
				err = ok
				return
			}
			e.Metadata = info
		} else {
			err = errors.New("Missing main epub for " + e.ShortString())
			return
		}
	}
	// refresh and clean Metadata
	e.Metadata.Clean(e.Config)

	// refresh both epubs
	var wasRenamedR, wasRenamedNR bool
	var newNameR, newNameNR string
	var errR, errNR error
	if e.HasRetail() {
		_, exists := h.FileExists(e.RetailEpub.FullPath())
		if exists == nil {
			wasRenamedR, newNameR, errR = e.RefreshEpub(e.RetailEpub, true)
			if wasRenamedR {
				e.RetailEpub.Filename = newNameR
			}
		} else {
			h.Warning("MISSING EPUB " + e.RetailEpub.FullPath())
			e.RetailEpub = Epub{}
		}
	}
	if e.HasNonRetail() {
		_, exists := h.FileExists(e.NonRetailEpub.FullPath())
		if exists == nil {
			wasRenamedNR, newNameNR, errNR = e.RefreshEpub(e.NonRetailEpub, false)
			if wasRenamedNR {
				e.NonRetailEpub.Filename = newNameNR
			}
		} else {
			h.Warning("MISSING EPUB " + e.NonRetailEpub.FullPath())
			e.NonRetailEpub = Epub{}
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
func (e *Book) HasRetail() (hasRetail bool) {
	return e.RetailEpub.Filename != ""
}

// HasNonRetail checks if a non-retail epub is available.
func (e *Book) HasNonRetail() (hasNonRetail bool) {
	return e.NonRetailEpub.Filename != ""
}

// HasEpub checks if the book has at least one epub
func (e *Book) HasEpub() bool {
	return e.HasRetail() || e.HasNonRetail()
}

// AddEpub to the Library
func (e *Book) AddEpub(path string, isRetail bool, hash string) (imported bool, err error) {
	// TODO tests
	if isRetail {
		if e.HasRetail() {
			h.Info("Trying to import retail epub although retail version already exists.")
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
			h.Warning("Non-retail version trumped, removing.")
			// replace ,nonretail
			err = e.removeEpub(!isRetail)
			if err != nil {
				return
			}
		}
	} else {
		if e.HasRetail() {
			h.Info("Trying to import non-retail epub although retail version exists, ignoring.")
		} else {
			if e.HasNonRetail() {
				h.Info("Trying to import non-retail epub although a non-retail version already exists.")
				if e.NonRetailEpub.NeedsReplacement == "true" {
					// replace ,nonretail
					h.Warning("Replacing non-retail version, flagged for replacement.")
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
	h.Debug("Importing " + path)
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
		h.Debug(err.Error())
		h.Warning("Could not retrieve information from GoodReads. Manual review.")
		err = e.EditField()
		if err != nil {
			h.Error(err.Error())
		}
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
		nonRetailHasChanged, err = e.NonRetailEpub.Check()
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

// GetOnlineMetadata retrieves the online info for this book.
func (e *Book) GetOnlineMetadata() (onlineInfo Metadata, err error) {
	if e.Config.GoodReadsAPIKey == "" {
		h.Error("Goodreads API key not found, not getting online information.")
		return
	}
	// TODO tests
	var g RemoteLibraryAPI
	g = GoodReads{}
	id := ""
	if e.Metadata.ISBN != "" {
		id, err = g.GetBookIDByISBN(e.Metadata.ISBN, e.Config.GoodReadsAPIKey)
		if err != nil {
			return
		}
	}
	// if no ISBN or nothing was found
	if id == "" {
		// TODO: if unsure, show hits
		id, err = g.GetBookIDByQuery(e.Metadata.Author(), e.Metadata.Title(), e.Config.GoodReadsAPIKey)
		if err != nil {
			return
		}
	}
	// if still nothing was found...
	if id == "" {
		// TODO ask for user input
		return Metadata{}, errors.New("Could not find online data for " + e.ShortString())
	}
	// get book info
	onlineInfo, err = g.GetBook(id, e.Config.GoodReadsAPIKey)
	if err == nil {
		onlineInfo.Clean(e.Config)
	}
	return
}

// SearchOnline tries to find metadata from online sources.
func (e *Book) SearchOnline() (err error) {
	onlineInfo, err := e.GetOnlineMetadata()
	if err != nil {
		return err
	}

	// show diff between epub and GR versions, then ask what to do.
	fmt.Println(e.Metadata.Diff(onlineInfo, "Epub Metadata", "GoodReads"))
	fmt.Printf(h.GreenBold("Choose: (1) Local version (2) Remote version (3) Edit (4) Abort "))
	validChoice := false
	errs := 0
	for !validChoice {
		scanner := bufio.NewReader(os.Stdin)
		choice, _ := scanner.ReadString('\n')
		choice = strings.TrimSpace(choice)
		switch choice {
		case "4":
			err = errors.New("Abort")
			validChoice = true
		case "3":
			h.Info("Going through every field.")
			err = e.Metadata.Merge(onlineInfo, e.Config)
			if err != nil {
				return err
			}
			validChoice = true
		case "2":
			h.Info("Accepting online version.")
			e.Metadata = onlineInfo
			validChoice = true
		case "1":
			h.Info("Keeping epub version.")
			validChoice = true
		default:
			fmt.Println("Invalid choice.")
			errs++
			if errs > 10 {
				return errors.New("Too many invalid choices.")
			}
		}
	}
	return
}

func (e *Book) editSpecificField(field string, values []string) (err error) {
	switch field {
	case "tags", "tag":
		fmt.Println("NOTE: tags can be edited as a comma-separated list of strings.")
		newValues, err := h.AssignNewValues(field, e.Metadata.Tags.String(), values)
		if err != nil {
			return err
		}
		// if user input was entered, we have to split the line
		if len(newValues) == 1 {
			newValues = strings.Split(newValues[0], ",")
		}
		for i := range newValues {
			newValues[i] = strings.TrimSpace(newValues[i])
		}
		// remove all tags
		e.Metadata.Tags = Tags{}
		// add new ones
		if e.Metadata.Tags.AddFromNames(newValues...) {
			h.Infof("Tags added to %s\n", e.ShortString())
		}
	case "series":
		fmt.Println("NOTE: series can be edited as a comma-separated list of 'series name:index' strings. Index can be empty, or a range.")
		newValues, err := h.AssignNewValues(field, e.Metadata.Series.String(), values)
		if err != nil {
			return err
		}
		// if user input was entered, we have to split the line
		if len(newValues) == 1 && strings.TrimSpace(newValues[0]) != "" {
			// remove all Series
			e.Metadata.Series = Series{}
			for _, s := range strings.Split(newValues[0], ",") {
				_, err := e.Metadata.Series.AddFromString(s)
				if err != nil {
					h.Warning("Could not parse series " + s + ", " + err.Error())
				}
			}
		}
	case "author", "authors":
		newValues, err := h.AssignNewValues(field, e.Metadata.Author(), values)
		if err != nil {
			return err
		}
		e.Metadata.Authors = strings.Split(newValues[0], ",")
		// trim spaces
		for j := range e.Metadata.Authors {
			e.Metadata.Authors[j] = strings.TrimSpace(e.Metadata.Authors[j])
		}
	case "year":
		newValues, err := h.AssignNewValues(field, e.Metadata.OriginalYear, values)
		if err != nil {
			return err
		}
		// check it's a valid date!
		_, err = strconv.Atoi(newValues[0])
		if err != nil {
			return err
		}
		e.Metadata.OriginalYear = newValues[0]
	case "edition_year":
		newValues, err := h.AssignNewValues(field, e.Metadata.EditionYear, values)
		if err != nil {
			return err
		}
		// check it's a valid date!
		_, err = strconv.Atoi(newValues[0])
		if err != nil {
			return err
		}
		e.Metadata.EditionYear = newValues[0]
	case "language":
		newValues, err := h.AssignNewValues(field, e.Metadata.Language, values)
		if err != nil {
			return err
		}
		e.Metadata.Language = newValues[0]
	case "category":
		newValues, err := h.AssignNewValues(field, e.Metadata.Category, values)
		if err != nil {
			return err
		}
		e.Metadata.Category = newValues[0]
	case "maingenre", "main_genre", "genre":
		newValues, err := h.AssignNewValues(field, e.Metadata.MainGenre, values)
		if err != nil {
			return err
		}
		e.Metadata.MainGenre = newValues[0]
	case "isbn":
		newValues, err := h.AssignNewValues(field, e.Metadata.ISBN, values)
		if err != nil {
			return err
		}
		// check it's a valid ISBN
		isbn, err := h.CleanISBN(newValues[0])
		if err != nil {
			return err
		}
		e.Metadata.ISBN = isbn
	case "title":
		newValues, err := h.AssignNewValues(field, e.Metadata.MainTitle, values)
		if err != nil {
			return err
		}
		e.Metadata.MainTitle = newValues[0]
		e.Metadata.OriginalTitle = newValues[0]
	case "description":
		newValues, err := h.AssignNewValues(field, e.Metadata.Description, values)
		if err != nil {
			return err
		}
		e.Metadata.Description = newValues[0]
	case "publisher":
		newValues, err := h.AssignNewValues(field, e.Metadata.Publisher, values)
		if err != nil {
			return err
		}
		e.Metadata.Publisher = newValues[0]
	case "progress":
		newValues, err := h.AssignNewValues(field, e.Progress, values)
		if err != nil {
			return err
		}
		if _, isIn := h.StringInSlice(newValues[0], validProgress); isIn {
			e.Progress = newValues[0]
		} else {
			return errors.New(newValues[0] + " is not a valid reading progress")
		}
	case "read_date", "readdate":
		newValues, err := h.AssignNewValues(field, e.ReadDate, values)
		if err != nil {
			return err
		}
		// check it's a valid date
		_, err = time.Parse("2006-01-02", newValues[0])
		if err != nil {
			return err
		}
		e.ReadDate = newValues[0]
	case "rating":
		newValues, err := h.AssignNewValues(field, e.Rating, values)
		if err != nil {
			return err
		}
		// checking rating is between 0 and 10
		val, err := strconv.Atoi(newValues[0])
		if err != nil || val > 10 || val < 0 {
			h.Error("Rating must be an integer between 0 and 10.")
			return err
		}
		e.Rating = newValues[0]
	case "review":
		newValues, err := h.AssignNewValues(field, e.Review, values)
		if err != nil {
			return err
		}
		e.Review = newValues[0]
	default:
		h.Debug("Unknown field: " + field)
		return errors.New("Unknown field: " + field)
	}
	// cleaning all metadata
	e.Metadata.Clean(e.Config)
	return
}

// EditField in current Metadata associated with the Book.
func (e *Book) EditField(args ...string) (err error) {
	if len(args) == 0 {
		// completely interactive edit
		for _, field := range []string{"author", "title", "year", "edition_year", "category", "genre", "tags", "series", "language", "isbn", "description", "progress"} {
			err = e.editSpecificField(field, []string{})
			if err != nil {
				fmt.Println("Could not assign new value to field " + field + ", continuing.")
			}
		}
	} else {
		field := strings.ToLower(args[0])
		values := []string{}
		if len(args) > 1 {
			values = args[1:]
		}
		err = e.editSpecificField(field, values)
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
