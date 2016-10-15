package book

import (
	"errors"
	"path/filepath"

	en "github.com/barsanuphe/endive/endive"

	"github.com/barsanuphe/epubgo"
)

// Epub can manipulate an epub file.
type Epub struct {
	Config en.Config        `json:"-"`
	UI     en.UserInterface `json:"-"`

	Filename         string `json:"filename"` // relative to LibraryRoot
	Hash             string `json:"hash"`
	NeedsReplacement string `json:"replace"`
}

// FullPath returns the absolute file path.
// if it is in the library, prepends LibraryRoot.
// if it is outside, return Filename directly.
func (e *Epub) FullPath() string {
	if filepath.IsAbs(e.Filename) {
		return e.Filename
	}
	return filepath.Join(e.Config.LibraryRoot, e.Filename)
}

// GetHash calculates an epub's current hash
func (e *Epub) GetHash() (err error) {
	hash, err := en.CalculateSHA256(e.FullPath())
	if err != nil {
		return
	}
	e.Hash = hash
	return
}

// FlagForReplacement an epub of insufficient quality
func (e *Epub) FlagForReplacement() (err error) {
	e.NeedsReplacement = en.True
	return
}

// Check the retail epub integrity.
func (e *Epub) Check() (hasChanged bool, err error) {
	// get current hash
	currentHash, err := en.CalculateSHA256(e.FullPath())
	if err != nil {
		return
	}
	// compare with old
	if currentHash != e.Hash {
		hasChanged = true
	}
	return
}

// ReadMetadata from epub file
func (e *Epub) ReadMetadata() (info Metadata, err error) {
	e.UI.Debugf("Reading metadata from %s\n", e.FullPath())
	book, err := epubgo.Open(e.FullPath())
	if err != nil {
		err = errors.New("Error parsing EPUB")
		return
	}
	defer book.Close()

	// year
	dateEvents, nonFatalErr := book.MetadataElement("date")
	if nonFatalErr != nil || len(dateEvents) == 0 {
		e.UI.Debug("Parsing EPUB: no date found")
	} else {
		found := false
		// try to find date associated with "publication" event
		for _, el := range dateEvents {
			for _, evt := range el.Attr {
				if evt == "publication" {
					info.EditionYear = el.Content[0:4]
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		// else reverting to first date found
		if !found {
			// using first date found
			info.EditionYear = dateEvents[0].Content[0:4]
		}
	}
	// by default, assuming it's a first edition
	info.OriginalYear = info.EditionYear
	// title
	if results, err := getFirstFieldValue(book, "title"); err == nil {
		info.BookTitle = results
	}
	// authors
	if results, err := getFieldValues(book, "creator"); err == nil {
		info.Authors = []string{}
		info.Authors = append(info.Authors, results...)
	}
	// language
	if results, err := getFirstFieldValue(book, "language"); err == nil {
		info.Language = results
	}
	// description
	if results, err := getFirstFieldValue(book, "description"); err == nil {
		info.Publisher = results
	}
	// tags
	if results, err := getFieldValues(book, "subject"); err == nil {
		info.Tags = Tags{}
		for _, t := range results {
			info.Tags.Add(Tag{Name: t})
		}
	}
	// publisher
	if results, err := getFirstFieldValue(book, "publisher"); err == nil {
		info.Publisher = results
	}

	// ISBN
	nonFatalErr = e.findISBN(book, &info)
	if nonFatalErr != nil {
		e.UI.Warningf("ISBN could not be found in %s!!", e.FullPath())
		err = nonFatalErr
	}

	// cleaning metadata
	info.Clean(e.Config)
	return
}

func getFirstFieldValue(epub *epubgo.Epub, field string) (string, error) {
	values, err := getFieldValues(epub, field)
	if err == nil && len(values) != 0 {
		return values[0], nil
	}
	return "", err
}
func getFieldValues(epub *epubgo.Epub, field string) ([]string, error) {
	values := []string{}
	results, err := epub.MetadataElement(field)
	if err == nil {
		for _, v := range results {
			values = append(values, v.Content)
		}
	}
	return values, err
}

func (e *Epub) findISBN(book *epubgo.Epub, i *Metadata) error {
	// get the identifier
	identifiers, nonFatalErr := book.MetadataElement("identifier")
	if nonFatalErr == nil && len(identifiers) != 0 {
		// try to find isbn
		for _, el := range identifiers {
			// try to find isbn in content
			isbn, err := en.CleanISBN(el.Content)
			if err == nil {
				i.ISBN = isbn
				return err
			}
			// try to find isbn in the attributes
			// it shouldn't be there, but retail epubs have awful metadata
			for _, evt := range el.Attr {
				isbn, err = en.CleanISBN(evt)
				if err == nil {
					i.ISBN = isbn
					return err
				}
			}
		}
	}
	// try getting source if not already found
	sources, nonFatalErr := book.MetadataElement("source")
	if nonFatalErr == nil && len(sources) != 0 {
		// try to find isbn
		for _, el := range sources {
			// clean results
			isbn, err := en.CleanISBN(el.Content)
			if err == nil {
				i.ISBN = isbn
				return err
			}
		}
	}

	// if no valid result, return err
	return errors.New("ISBN not found in epub")
}
