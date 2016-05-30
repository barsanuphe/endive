package book

import (
	"errors"
	"path/filepath"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
	"github.com/barsanuphe/epubgo"
)

// Epub can manipulate an epub file.
type Epub struct {
	Filename         string     `json:"filename"` // relative to LibraryRoot
	Config           cfg.Config `json:"-"`
	Hash             string     `json:"hash"`
	NeedsReplacement string     `json:"replace"`
}

// FullPath returns the absolute file path.
// if it is in the library, prepends LibraryRoot.
// if it is outside, return Filename directly.
func (e *Epub) FullPath() (path string) {
	// TODO: tests
	if filepath.IsAbs(e.Filename) {
		return e.Filename
	}
	return filepath.Join(e.Config.LibraryRoot, e.Filename)
}

// GetHash calculates an epub's current hash
func (e *Epub) GetHash() (err error) {
	hash, err := h.CalculateSHA256(e.FullPath())
	if err != nil {
		return
	}
	e.Hash = hash
	return
}

// FlagForReplacement an epub of insufficient quality
func (e *Epub) FlagForReplacement() (err error) {
	e.NeedsReplacement = "true"
	return
}

// Check the retail epub integrity.
func (e *Epub) Check() (hasChanged bool, err error) {
	// get current hash
	currentHash, err := h.CalculateSHA256(e.FullPath())
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
	h.Debugf("Reading metadata from %s\n", e.FullPath())
	book, err := epubgo.Open(e.FullPath())
	if err != nil {
		h.Error("Error parsing EPUB")
		return
	}
	defer book.Close()

	// year
	dateEvents, nonFatalErr := book.MetadataElement("date")
	if nonFatalErr != nil || len(dateEvents) == 0 {
		h.Debug("Error parsing EPUB: no date found")
	} else {
		found := false
		// try to find date associated with "publication" event
		for _, el := range dateEvents {
			for _, evt := range el.Attr {
				if evt == "publication" {
					info.Year = el.Content[0:4]
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
			info.Year = dateEvents[0].Content[0:4]
		}
	}
	// title
	results, nonFatalErr := book.MetadataElement("title")
	if nonFatalErr == nil && len(results) != 0 {
		info.MainTitle = results[0].Content
	}
	// authors
	results, nonFatalErr = book.MetadataElement("creator")
	if nonFatalErr == nil && len(results) != 0 {
		info.Authors = []string{}
		for _, t := range results {
			info.Authors = append(info.Authors, t.Content)
		}
	}
	// language
	results, nonFatalErr = book.MetadataElement("language")
	if nonFatalErr == nil && len(results) != 0 {
		info.Language = results[0].Content
	}
	// description
	results, nonFatalErr = book.MetadataElement("description")
	if nonFatalErr == nil && len(results) != 0 {
		info.Description = results[0].Content
	}
	// tags
	results, nonFatalErr = book.MetadataElement("subject")
	if nonFatalErr == nil && len(results) != 0 {
		info.Tags = Tags{}
		for _, t := range results {
			tag := Tag{Name: t.Content}
			info.Tags.Add(tag)
		}
	}
	// ISBN
	nonFatalErr = e.findISBN(book, &info)
	if nonFatalErr != nil {
		h.Warning("ISBN could not be found in %s!!", e.FullPath())
	}
	// publisher
	results, nonFatalErr = book.MetadataElement("publisher")
	if nonFatalErr == nil && len(results) != 0 {
		info.Publisher = results[0].Content
	}

	// cleaning metadata
	info.Clean(e.Config)
	return
}

func (e *Epub) findISBN(book *epubgo.Epub, i *Metadata) (err error) {
	// get the identifier
	identifiers, nonFatalErr := book.MetadataElement("identifier")
	if nonFatalErr == nil && len(identifiers) != 0 {
		// try to find isbn
		for _, el := range identifiers {
			// try to find isbn in content
			isbn, err := cleanISBN(el.Content)
			if err == nil {
				i.ISBN = isbn
				return err
			}
			// try to find isbn in the attributes
			// it shouldn't be there, but retail epubs have awful metadata
			for _, evt := range el.Attr {
				isbn, err = cleanISBN(evt)
				if err == nil {
					i.ISBN = isbn
					return err
				}
			}
		}

		// try getting source
		sources, nonFatalErr := book.MetadataElement("source")
		if nonFatalErr == nil && len(sources) != 0 {
			// try to find isbn
			for _, el := range sources {
				// clean results
				isbn, err := cleanISBN(el.Content)
				if err == nil {
					i.ISBN = isbn
					return err
				}
			}
		}
	}
	// if no valid result, return err
	h.Debugf("ISBN not found in %s", e.FullPath())
	return errors.New("ISBN not found in epub")
}
