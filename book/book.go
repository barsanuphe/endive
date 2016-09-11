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

	e "github.com/barsanuphe/endive/endive"
)

var validProgress = []string{"unread", "read", "reading", "shortlisted"}
var validCategories = []string{"fiction", "nonfiction"}

// Book can manipulate a book.
// A Book can have multiple epub files.
type Book struct {
	Config e.Config        `json:"-"`
	UI     e.UserInterface `json:"-"`
	ID     int             `json:"id"`
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
		return &Book{ID: id, RetailEpub: f, Config: c, UI: ui, EpubMetadata: i, Metadata: i, Progress: "unread"}
	}
	return &Book{ID: id, NonRetailEpub: f, Config: c, UI: ui, EpubMetadata: i, Metadata: i, Progress: "unread"}
}

// ShortString returns a short string representation of Epub
func (b *Book) ShortString() string {
	return b.Metadata.Author() + " (" + b.Metadata.OriginalYear + ") " + b.Metadata.Title()
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
			rows = append(rows, []string{"ID", strconv.Itoa(b.ID)})
		case "filename":
			rows = append(rows, []string{"Filename", b.MainEpub().Filename})
		case authorField:
			rows = append(rows, []string{"Author", b.Metadata.Author()})
		case titleField:
			rows = append(rows, []string{"Title", b.Metadata.Title()})
		case yearField:
			rows = append(rows, []string{"Original Publication Year", b.Metadata.OriginalYear})
		case editionYearField:
			rows = append(rows, []string{"Publication Year", b.Metadata.EditionYear})
		case publisherField:
			rows = append(rows, []string{"Publisher", b.Metadata.Publisher})
		case isbnField:
			rows = append(rows, []string{"ISBN", b.Metadata.ISBN})
		case descriptionField:
			rows = append(rows, []string{"Description", b.Metadata.Description})
		case numPagesField:
			if b.Metadata.NumPages != "" {
				rows = append(rows, []string{"Number of pages", b.Metadata.NumPages})
			}
		case languageField:
			rows = append(rows, []string{"Language", b.Metadata.Language})
		case categoryField:
			rows = append(rows, []string{"Category", b.Metadata.Category})
		case genreField:
			rows = append(rows, []string{"Main Genre", b.Metadata.MainGenre})
		case tagsField:
			if len(b.Metadata.Tags) != 0 {
				rows = append(rows, []string{"Tags", b.Metadata.Tags.String()})
			}
		case seriesField:
			if len(b.Metadata.Series) != 0 {
				rows = append(rows, []string{"Series", b.Metadata.Series.String()})
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
			rows = append(rows, []string{"Progress", b.Progress})
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
				rows = append(rows, []string{"Rating", b.Rating})
			}
		case reviewField:
			if b.Review != "" {
				rows = append(rows, []string{"Review", b.Review})
			}
		}
	}
	return e.TabulateRows(rows, "Info", "Book")
}

// FullPath of the main Epub of a Book.
func (b *Book) FullPath() string {
	// assuming at least one epub is defined
	if b.HasRetail() {
		return b.RetailEpub.FullPath()
	} else if b.HasNonRetail() {
		return b.NonRetailEpub.FullPath()
	} else {
		panic(errors.New("Book has no epub file!"))
	}
}

// MainEpub of a Book.
func (b *Book) MainEpub() (epub *Epub) {
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
func (b *Book) SetReadDate(date string) (err error) {
	b.ReadDate = date
	return
}

// SetReadDateToday sets date when finished reading
func (b *Book) SetReadDateToday() (err error) {
	currentDate := time.Now().Local()
	return b.SetReadDate(currentDate.Format("2006-01-02"))
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

// ForceMetadataRefresh overwrites current Metadata
func (b *Book) ForceMetadataRefresh() (err error) {
	_, exists := e.FileExists(b.MainEpub().FullPath())
	if exists == nil {
		info, ok := b.MainEpub().ReadMetadata()
		if ok != nil {
			err = ok
			return
		}
		b.EpubMetadata = info
		b.Metadata = info
	} else {
		err = errors.New("Missing main epub for " + b.ShortString())
		return
	}

	// get online data
	err = b.SearchOnline()
	if err != nil {
		b.UI.Warning(err.Error())
	}
	return
}

// ForceMetadataFieldRefresh overwrites current Metadata for a specific field only.
func (b *Book) ForceMetadataFieldRefresh(field string) (err error) {
	info := Metadata{}
	_, exists := e.FileExists(b.MainEpub().FullPath())
	if exists == nil {
		info, err = b.MainEpub().ReadMetadata()
		if err != nil {
			return
		}
	} else {
		err = errors.New("Missing main epub for " + b.ShortString())
		return
	}
	// get online data
	onlineInfo, err := b.GetOnlineMetadata()
	if err != nil {
		return err
	}
	// merge field
	err = info.MergeField(onlineInfo, field, b.Config, b.UI)
	if err != nil {
		return err
	}
	switch field {
	case tagsField:
		b.Metadata.Tags = info.Tags
	case seriesField:
		b.Metadata.Series = info.Series
	case authorField:
		b.Metadata.Authors = info.Authors
	case yearField:
		b.Metadata.OriginalYear = info.OriginalYear
	case editionYearField:
		b.Metadata.EditionYear = info.EditionYear
	case publisherField:
		b.Metadata.Publisher = info.Publisher
	case languageField:
		b.Metadata.Language = info.Language
	case categoryField:
		b.Metadata.Category = info.Category
	case genreField:
		b.Metadata.MainGenre = info.MainGenre
	case isbnField:
		b.Metadata.ISBN = info.ISBN
	case titleField:
		b.Metadata.MainTitle = info.MainTitle
		b.Metadata.OriginalTitle = info.OriginalTitle
	case descriptionField:
		b.Metadata.Description = info.Description
	default:
		return errors.New("Unknown field: " + field)
	}
	// cleaning all metadata
	b.Metadata.Clean(b.Config)
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
	// TODO tests
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

// GetOnlineMetadata retrieves the online info for this book.
func (b *Book) GetOnlineMetadata() (onlineInfo Metadata, err error) {
	if b.Config.GoodReadsAPIKey == "" {
		return Metadata{}, e.WarningGoodReadsAPIKeyMissing
	}
	var g RemoteLibraryAPI
	g = GoodReads{}
	id := ""

	// If not ISBN is found, ask for input
	if b.Metadata.ISBN == "" {
		b.UI.Warning("Could not find ISBN.")
		isbn, err := e.AskForISBN(b.UI)
		if err == nil {
			b.Metadata.ISBN = isbn
		}
	}
	// search by ISBN preferably
	if b.Metadata.ISBN != "" {
		id, err = g.GetBookIDByISBN(b.Metadata.ISBN, b.Config.GoodReadsAPIKey)
		if err != nil {
			return
		}
	}
	// if no ISBN or nothing was found
	if id == "" {
		// TODO: if unsure, show hits
		id, err = g.GetBookIDByQuery(b.Metadata.Author(), b.Metadata.Title(), b.Config.GoodReadsAPIKey)
		if err != nil {
			return
		}
	}
	// if still nothing was found...
	if id == "" {
		return Metadata{}, errors.New("Could not find online data for " + b.ShortString())
	}
	// get book info
	onlineInfo, err = g.GetBook(id, b.Config.GoodReadsAPIKey)
	if err == nil {
		onlineInfo.Clean(b.Config)
	}
	return
}

// SearchOnline tries to find metadata from online sources.
func (b *Book) SearchOnline() (err error) {
	onlineInfo, err := b.GetOnlineMetadata()
	if err != nil {
		return err
	}

	// show diff between epub and GR versions, then ask what to do.
	fmt.Println(b.Metadata.Diff(onlineInfo, "Epub Metadata", "GoodReads"))
	b.UI.Choice("Choose: (1) Local version (2) Remote version (3) Edit (4) Abort : ")
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
			err = b.Metadata.Merge(onlineInfo, b.Config, b.UI)
			if err != nil {
				return err
			}
			validChoice = true
		case "2":
			b.UI.Info("Accepting online version.")
			b.Metadata = onlineInfo
			validChoice = true
		case "1":
			b.UI.Info("Keeping epub version.")
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

func (b *Book) editSpecificField(field string, values []string) error {
	switch field {
	case tagsField:
		fmt.Println("NOTE: tags can be edited as a comma-separated list of strings.")
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Tags.String(), values)
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
		b.Metadata.Tags = Tags{}
		// add new ones
		if b.Metadata.Tags.AddFromNames(newValues...) {
			b.UI.Infof("Tags added to %s\n", b.ShortString())
		}
	case seriesField:
		fmt.Println("NOTE: series can be edited as a comma-separated list of 'series name:index' strings. Index can be empty, or a range.")
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Series.rawString(), values)
		if err != nil {
			return err
		}
		// if user input was entered, we have to split the line
		if len(newValues) == 1 && newValues[0] != b.Metadata.Series.rawString() && strings.TrimSpace(newValues[0]) != "" {
			// remove all Series
			b.Metadata.Series = Series{}
			for _, s := range strings.Split(newValues[0], ",") {
				_, err := b.Metadata.Series.AddFromString(s)
				if err != nil {
					b.UI.Warning("Could not parse series " + s + ", " + err.Error())
				}
			}
		}
	case authorField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Author(), values)
		if err != nil {
			return err
		}
		b.Metadata.Authors = strings.Split(newValues[0], ",")
		// trim spaces
		for j := range b.Metadata.Authors {
			b.Metadata.Authors[j] = strings.TrimSpace(b.Metadata.Authors[j])
		}
	case yearField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.OriginalYear, values)
		if err != nil {
			return err
		}
		// check it's a valid date!
		_, err = strconv.Atoi(newValues[0])
		if err != nil {
			return err
		}
		b.Metadata.OriginalYear = newValues[0]
	case editionYearField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.EditionYear, values)
		if err != nil {
			return err
		}
		// check it's a valid date!
		_, err = strconv.Atoi(newValues[0])
		if err != nil {
			return err
		}
		b.Metadata.EditionYear = newValues[0]
	case languageField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Language, values)
		if err != nil {
			return err
		}
		b.Metadata.Language = newValues[0]
	case categoryField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Category, values)
		if err != nil {
			return err
		}
		b.Metadata.Category = newValues[0]
	case genreField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.MainGenre, values)
		if err != nil {
			return err
		}
		b.Metadata.MainGenre = newValues[0]
	case isbnField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.ISBN, values)
		if err != nil {
			return err
		}
		// check it's a valid ISBN
		isbn, err := e.CleanISBN(newValues[0])
		if err != nil {
			return err
		}
		b.Metadata.ISBN = isbn
	case titleField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.MainTitle, values)
		if err != nil {
			return err
		}
		b.Metadata.MainTitle = newValues[0]
		b.Metadata.OriginalTitle = newValues[0]
	case descriptionField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Description, values)
		if err != nil {
			return err
		}
		b.Metadata.Description = newValues[0]
	case publisherField:
		newValues, err := b.UI.UpdateValues(field, b.Metadata.Publisher, values)
		if err != nil {
			return err
		}
		b.Metadata.Publisher = newValues[0]
	case progressField:
		newValues, err := b.UI.UpdateValues(field, b.Progress, values)
		if err != nil {
			return err
		}
		if _, isIn := e.StringInSlice(newValues[0], validProgress); isIn {
			b.Progress = newValues[0]
		} else {
			return errors.New(newValues[0] + " is not a valid reading progress")
		}
	case readDateField:
		newValues, err := b.UI.UpdateValues(field, b.ReadDate, values)
		if err != nil {
			return err
		}
		// check it's a valid date
		_, err = time.Parse("2006-01-02", newValues[0])
		if err != nil {
			return err
		}
		b.ReadDate = newValues[0]
	case ratingField:
		newValues, err := b.UI.UpdateValues(field, b.Rating, values)
		if err != nil {
			return err
		}
		// checking rating is between 0 and 10
		val, err := strconv.ParseFloat(newValues[0], 32)
		if err != nil || val > 5 || val < 0 {
			b.UI.Error("Rating must be between 0 and 5.")
			return err
		}
		b.Rating = newValues[0]
	case reviewField:
		newValues, err := b.UI.UpdateValues(field, b.Review, values)
		if err != nil {
			return err
		}
		b.Review = newValues[0]
	default:
		b.UI.Debug("Unknown field: " + field)
		return errors.New("Unknown field: " + field)
	}
	// cleaning all metadata
	b.Metadata.Clean(b.Config)
	return nil
}

// EditField in current Metadata associated with the Book.
func (b *Book) EditField(args ...string) (err error) {
	if len(args) == 0 {
		// completely interactive edit
		for _, field := range []string{"author", "title", "year", "edition_year", "category", "genre", "tags", "series", "language", "isbn", "description", "progress"} {
			err = b.editSpecificField(field, []string{})
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
		err = b.editSpecificField(field, values)
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
