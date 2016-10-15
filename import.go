package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	b "github.com/barsanuphe/endive/book"
	en "github.com/barsanuphe/endive/endive"
)

//------------------------------------

// importFromSource all detected epubs, tagging them as retail or non-retail as requested.
func (e *Endive) importFromSource(sources []string, retail bool) error {
	defer en.TimeTrack(e.UI, time.Now(), "Imported")
	candidates, err := e.analyzeSources(sources, retail)
	if err != nil {
		return err
	}
	return e.ImportEpubs(candidates.Importable(), retail)
}

func (e *Endive) analyzeSources(sources []string, retail bool) (en.EpubCandidates, error) {
	defer en.TimeTrack(e.UI, time.Now(), "Analyzed")
	sourceType := "retail"
	if !retail {
		sourceType = "non-retail"
	}
	e.UI.Title("Importing " + sourceType + " epubs...")

	// checking all defined sources
	var candidates en.EpubCandidates
	for _, source := range sources {
		e.UI.SubTitle("Searching for " + sourceType + " epubs in " + source)
		c, err := en.ScanForEpubs(source, e.hashes, e.Library.Collection)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, c...)
	}
	newEpubs := candidates.New()
	missingEpubs := candidates.Missing()
	e.UI.SubTitle("Found %d new epubs and %d epubs previously imported and now missing.", len(newEpubs), len(missingEpubs))
	return candidates.Importable(), nil
}

// ImportRetail imports epubs from the Retail source.
func (e *Endive) ImportRetail() error {
	return e.importFromSource(e.Config.RetailSource, true)
}

// ImportNonRetail imports epubs from the Non-Retail source.
func (e *Endive) ImportNonRetail() error {
	return e.importFromSource(e.Config.NonRetailSource, false)
}

// ImportSpecific imports specific epubs
func (e *Endive) ImportSpecific(isRetail bool, paths ...string) error {
	var candidates en.EpubCandidates
	// for each path:
	for _, path := range paths {
		// verify it exists
		validPath, err := en.FileExists(path)
		if err == nil && filepath.Ext(strings.ToLower(validPath)) == en.EpubExtension {
			candidates = append(candidates, *en.NewCandidate(validPath, e.hashes, e.Library.Collection))
		}
	}
	return e.ImportEpubs(candidates.Importable(), isRetail)
}

// ImportEpubs files that are retail, or not.
func (e *Endive) ImportEpubs(candidates []en.EpubCandidate, isRetail bool) (err error) {
	// force reload if it has changed
	err = e.hashes.Load()
	if err != nil {
		return
	}

	newEpubs := 0
	// importing what is necessary
	for i, candidate := range candidates {
		intro := fmt.Sprintf("Considering importable epub %s", filepath.Base(candidate.Filename))
		if len(candidates) > 1 {
			intro += fmt.Sprintf(" (%d / %d)", i+1, len(candidates))
		}
		e.UI.SubTitle(intro)
		// new Epub
		ep := b.Epub{Filename: candidate.Filename, UI: e.UI}
		var unknownISBN bool
		// get Metadata from new epub
		info, err := ep.ReadMetadata()
		if err != nil {
			if err.Error() == "ISBN not found in epub" {
				unknownISBN = true
			} else {
				e.UI.Error("Could not analyze and import " + candidate.Filename)
				continue
			}
		}

		confirmText := fmt.Sprintf("Found: %s.\n", info.String())
		if !candidate.Imported {
			confirmText += "Import"
		} else {
			confirmText += "This epub has already been imported but, is not in the current library. Confirm importing again?"
		}
		if e.UI.Accept(confirmText) {
			// get isbn if not found automatically
			if unknownISBN {
				isbn, err := en.AskForISBN(e.UI)
				if err != nil {
					e.UI.Warning("Warning: ISBN still unknown.")
				} else {
					info.ISBN = isbn
				}
			}

			// get online data to prepare import
			if err = info.SearchOnline(e.UI, e.Config); err != nil {
				e.UI.Error("Could not merge metadata with online sources. Continuing importing nonetheless.")
			}

			// loop over Books to find similar Metadata
			var imported bool
			knownBook, err := e.Library.Collection.FindByMetadata(info.ISBN, info.Author(), info.Title())
			if err != nil {
				e.UI.Debug("Creating new book.")
				bk := b.NewBookWithMetadata(e.UI, e.Library.GenerateID(), candidate.Filename, e.Config, isRetail, info)
				imported, err = bk.Import(candidate.Filename, isRetail, candidate.Hash)
				if err != nil {
					return err
				}
				// adding new book
				e.Library.Collection.Add(bk)
				e.UI.SubTitle("Added epub %s to new book with ID %d", bk.ShortString(), bk.ID())
			} else {
				e.UI.Debug("Adding epub to " + knownBook.ShortString())
				imported, err = knownBook.AddEpub(candidate.Filename, isRetail, candidate.Hash)
				if err != nil {
					return err
				}
				e.UI.SubTitle("Added new epub %s to book with ID %d", knownBook.ShortString(), knownBook.ID())
			}

			if imported {
				// add hash to known hashes
				added, err := e.hashes.Add(candidate.Hash)
				if !added || err != nil {
					return err
				}
				// saving now == saving import progress, in case of interruption
				_, err = e.hashes.Save()
				if err != nil {
					return err
				}
				// saving database also
				_, err = e.Library.Save()
				if err != nil {
					return err
				}
				newEpubs++
			}
		} else {
			e.UI.Debug("Ignoring epub " + filepath.Base(candidate.Filename))
		}
	}
	e.UI.Debugf("Imported %d epubs (retail: %t).\n", newEpubs, isRetail)
	if newEpubs == 0 {
		err = errors.New("Nothing to import, epubs already in library")
	}
	return
}
