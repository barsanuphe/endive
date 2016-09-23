package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	b "github.com/barsanuphe/endive/book"
	"github.com/barsanuphe/endive/db"
	en "github.com/barsanuphe/endive/endive"
	i "github.com/barsanuphe/endive/index"
	l "github.com/barsanuphe/endive/library"
	u "github.com/barsanuphe/endive/ui"
)

// Endive is the main struct here.
type Endive struct {
	hashes  en.KnownHashes
	Config  en.Config
	UI      en.UserInterface
	Library l.Library
}

// NewEndive constructs a valid new Epub
func NewEndive() (*Endive, error) {
	// init ui
	var ui en.UserInterface
	ui = u.UI{}
	if err := ui.InitLogger(en.XdgLogPath); err != nil {
		return nil, err
	}

	// init known hashes
	hashesPath, err := en.GetKnownHashesPath()
	if err != nil {
		return nil, err
	}
	hashes := en.KnownHashes{Filename: hashesPath}
	err = hashes.Load()
	if err != nil {
		return nil, err
	}

	// create Endive
	e := &Endive{UI: ui, hashes: hashes}
	// open Config
	if err := e.openConfig(); err != nil {
		return e, err
	}
	// open library
	if err := e.openLibrary(); err != nil {
		return e, err
	}

	// set lock
	err = en.SetLock()
	return e, err
}

func (e *Endive) openConfig() error {
	// config
	configPath, err := en.GetConfigPath()
	if err != nil {
		return err
	}
	e.Config = en.Config{Filename: configPath}
	// config load
	e.UI.Debugf("Loading Config %s.\n", e.Config.Filename)
	err = e.Config.Load()
	if err != nil {
		if err == en.WarningGoodReadsAPIKeyMissing {
			e.UI.Warning(err.Error())
		} else {
			e.UI.Error(err.Error())
		}
		return err
	}
	// check config
	e.UI.Debug("Checking Config...")
	err = e.Config.Check()
	if err == en.WarningNonRetailSourceDoesNotExist || err == en.WarningRetailSourceDoesNotExist {
		e.UI.Warning(err.Error())
	}
	return err
}

// OpenLibrary constucts a valid new Library
func (e *Endive) openLibrary() error {
	// index
	index := &i.Index{}
	index.SetPath(en.GetIndexPath())
	// db
	db := &db.JSONDB{}
	db.SetPath(e.Config.DatabaseFile)
	e.Library = l.Library{Collection: &b.Books{}, Config: e.Config, Index: index, UI: e.UI, DB: db}
	return e.Library.Load()
}

// importFromSource all detected epubs, tagging them as retail or non-retail as requested.
func (e *Endive) importFromSource(sources []string, retail bool) error {
	defer en.TimeTrack(e.UI, time.Now(), "Imported")
	sourceType := "retail"
	if !retail {
		sourceType = "non-retail"
	}
	e.UI.Title("Importing " + sourceType + " epubs...")

	// checking all defined sources
	var allEpubs, allHashes []string
	for _, source := range sources {
		e.UI.SubTitle("Searching for " + sourceType + " epubs in " + source)
		epubs, hashes, err := en.ListEpubsInDirectory(source)
		if err != nil {
			return err
		}
		allEpubs = append(allEpubs, epubs...)
		allHashes = append(allHashes, hashes...)
	}
	return e.ImportEpubs(allEpubs, allHashes, retail)
}

// ImportRetail imports epubs from the Retail source.
func (e *Endive) ImportRetail() error {
	return e.importFromSource(e.Config.RetailSource, true)
}

// ImportNonRetail imports epubs from the Non-Retail source.
func (e *Endive) ImportNonRetail() error {
	return e.importFromSource(e.Config.NonRetailSource, false)
}

// ImportEpubs files that are retail, or not.
func (e *Endive) ImportEpubs(allEpubs []string, allHashes []string, isRetail bool) (err error) {
	// force reload if it has changed
	err = e.hashes.Load()
	if err != nil {
		return
	}
	defer e.hashes.Save()

	newEpubs := 0
	// importing what is necessary
	for i, path := range allEpubs {
		hash := allHashes[i]
		importConfirmed := false
		ep := b.Epub{Filename: path, UI: e.UI}

		// compare with known hashes
		info := b.Metadata{}
		if !e.hashes.IsIn(hash) {
			// get Metadata from new epub
			info, err = ep.ReadMetadata()
			if err != nil {
				if err.Error() == "ISBN not found in epub" {
					isbn, err := en.AskForISBN(e.UI)
					if err != nil {
						e.UI.Warning("Warning: ISBN still unknown.")
					} else {
						info.ISBN = isbn
					}
				} else {
					e.UI.Error("Could not analyze and import " + path)
					continue
				}
			}
			// ask if user really wants to import it
			importConfirmed = e.UI.YesOrNo(fmt.Sprintf("Found %s (%s).\nImport", filepath.Base(path), info.String()))
		} else {
			_, err := e.Library.Collection.FindByHash(hash)
			if err != nil {
				// get Metadata from new epub
				info, err = ep.ReadMetadata()
				if err != nil {
					if err.Error() == "ISBN not found in epub" {
						isbn, err := en.AskForISBN(e.UI)
						if err != nil {
							e.UI.Warning("Warning: ISBN still unknown.")
						} else {
							info.ISBN = isbn
						}
					} else {
						e.UI.Error("Could not analyze and import " + path)
						continue
					}
				}
				//confirm force import
				importConfirmed = e.UI.YesOrNo(fmt.Sprintf("File %s has already been imported but is not in the current library. Confirm importing again?", filepath.Base(path)))
			}
		}

		if importConfirmed {
			// loop over Books to find similar Metadata
			var imported bool
			knownBook, err := e.Library.Collection.FindByMetadata(info.ISBN, info.Author(), info.Title())
			if err != nil {
				// new Book
				e.UI.Debug("Creating new book.")
				bk := b.NewBookWithMetadata(e.UI, e.Library.GenerateID(), path, e.Config, isRetail, info)
				imported, err = bk.Import(path, isRetail, hash)
				if err != nil {
					return err
				}
				e.Library.Collection.Add(bk)
			} else {
				// add to existing book
				e.UI.Debug("Adding epub to " + knownBook.ShortString())
				imported, err = knownBook.AddEpub(path, isRetail, hash)
				if err != nil {
					return err
				}
			}

			if imported {
				// add hash to known hashes
				// NOTE: otherwise it'll pop up every other time
				added, err := e.hashes.Add(hash)
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
			e.UI.Debug("Ignoring already imported epub " + filepath.Base(path))
		}
	}
	if isRetail {
		e.UI.Debugf("Imported %d retail epubs.\n", newEpubs)
	} else {
		e.UI.Debugf("Imported %d non-retail epubs.\n", newEpubs)
	}
	return
}

// Refresh current DB
func (e *Endive) Refresh() (renamed int, err error) {
	e.UI.Info("Refreshing database...")

	// scan for new epubs
	allEpubs, allHashes, err := en.ListEpubsInDirectory(e.Config.LibraryRoot)
	if err != nil {
		return
	}
	// compare allEpubs with l.Epubs
	newEpubs := []string{}
	newHashes := []string{}
	for i, epub := range allEpubs {
		_, err = e.Library.Collection.FindByFullPath(epub)
		// no error == found Epub
		if err != nil {
			// check if hash is known
			gBook, err := e.Library.Collection.FindByHash(allHashes[i])
			if err != nil {
				// else, it's a new epub, import
				e.UI.Info("NEW EPUB " + epub + " , will be imported as non-retail.")
				newEpubs = append(newEpubs, epub)
				newHashes = append(newHashes, allHashes[i])
			} else {
				var book *b.Book
				book = gBook.(*b.Book)
				// if it is, rename found file to filename in DB
				destination := book.RetailEpub.FullPath()
				if book.NonRetailEpub.Hash == allHashes[i] {
					destination = book.NonRetailEpub.FullPath()
				}
				// check if retail epub already exists
				_, err := en.FileExists(destination)
				if err == nil {
					// file already exists
					e.UI.Errorf("Found epub %s with the same hash as %s, ignoring.", epub, destination)
				} else {
					e.UI.Warningf("Found epub %s which is called %s in the database, renaming.", epub, destination)
					// rename found file to retail name in db
					err = os.Rename(epub, destination)
					if err != nil {
						return 0, err
					}
				}
			}
		}
	}
	// import new books as non-retail
	err = e.ImportEpubs(newEpubs, newHashes, false)
	if err != nil {
		return
	}
	// refresh all books
	deletedBooks := []int{}
	for i := range e.Library.Collection.Books() {
		wasRenamed, _, err := e.Library.Collection.Books()[i].Refresh()
		if err != nil {
			return renamed, err
		}
		if !e.Library.Collection.Books()[i].HasEpub() {
			// mark for deletion
			deletedBooks = append(deletedBooks, e.Library.Collection.Books()[i].ID())
		}
		if wasRenamed[0] {
			renamed++
		}
		if wasRenamed[1] {
			renamed++
		}
	}

	// remove empty books
	for _, id := range deletedBooks {
		e.UI.Infof("REMOVING from db Book with ID %s\n", id)
		err := e.Library.Collection.RemoveByID(id)
		if err != nil {
			return renamed, err
		}
	}
	// remove all empty dirs
	err = en.DeleteEmptyFolders(e.Config.LibraryRoot, e.UI)
	return
}
