package book

import (
	"fmt"
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
func (e *Epub) ReadMetadata() (info Info, err error) {
	fmt.Println("Reading metadata from " + e.FullPath())
	book, err := epubgo.Open(e.FullPath())
	if err != nil {
		fmt.Println("Error parsing EPUB")
		return
	}
	defer book.Close()

	// year
	dateEvents, dateErr := book.MetadataElement("date")
	if dateErr != nil || len(dateEvents) == 0 {
		fmt.Println("Error parsing EPUB: no date found")
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
	results, err := book.MetadataElement("title")
	if err == nil && len(results) != 0 {
		info.MainTitle = results[0].Content
	}
	// authors
	results, err = book.MetadataElement("creator")
	if err == nil && len(results) != 0 {
		info.Authors = []string{}
		for _, t := range results {
			info.Authors = append(info.Authors, t.Content)
		}
	}
	// language
	results, err = book.MetadataElement("language")
	if err == nil && len(results) != 0 {
		info.Language = results[0].Content
	}
	// description
	results, err = book.MetadataElement("description")
	if err == nil && len(results) != 0 {
		info.Description = results[0].Content
	}
	// tags
	results, err = book.MetadataElement("subject")
	if err == nil && len(results) != 0 {
		info.Tags = Tags{}
		for _, t := range results {
			tag := Tag{Name: t.Content}
			info.Tags.Add(tag)
		}
	}
	// TODO !!!! identifier (isbn)

	// TODO show other included data:"publisher", "contributor", "type", "format",
	// TODO "source", "relation", "coverage", "rights", "meta",

	if info.Refresh(e.Config) {
		fmt.Println("Found author alias: " + info.Author())
	}

	return
}
