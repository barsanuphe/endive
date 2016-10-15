package main

import (
	"os"

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
	ui = &u.UI{}
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

// Refresh current DB
func (e *Endive) Refresh() (renamed int, err error) {
	e.UI.Info("Refreshing database...")

	// scan for new epubs
	foundCandidates, err := en.ScanForEpubs(e.Config.LibraryRoot, e.hashes, e.Library.Collection)
	if err != nil {
		return
	}
	// compare allEpubs with l.Epubs
	var newEpubs en.EpubCandidates
	for _, epub := range foundCandidates {
		_, err = e.Library.Collection.FindByFullPath(epub.Filename)
		// no error == found Epub
		if err != nil {
			// check if hash is known
			gBook, err := e.Library.Collection.FindByHash(epub.Hash)
			if err != nil {
				// else, it's a new epub, import
				e.UI.Info("NEW EPUB " + epub.Filename + " , will be imported as non-retail.")
				newEpubs = append(newEpubs, epub)
			} else {
				var book *b.Book
				book = gBook.(*b.Book)
				// if it is, rename found file to filename in DB
				destination := book.RetailEpub.FullPath()
				if book.NonRetailEpub.Hash == epub.Hash {
					destination = book.NonRetailEpub.FullPath()
				}
				// check if retail epub already exists
				_, err := en.FileExists(destination)
				if err == nil {
					// file already exists
					e.UI.Errorf("Found epub %s with the same hash as %s, ignoring.", epub.Filename, destination)
				} else {
					e.UI.Warningf("Found epub %s which is called %s in the database, renaming.", epub.Filename, destination)
					// rename found file to retail name in db
					err = os.Rename(epub.Filename, destination)
					if err != nil {
						return 0, err
					}
				}
			}
		}
	}

	if len(newEpubs.Importable()) != 0 {
		// import new books as non-retail
		err = e.ImportEpubs(newEpubs.Importable(), false)
		if err != nil {
			return
		}
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
		e.UI.Infof("REMOVING from db Book with ID %d\n", id)
		err := e.Library.Collection.RemoveByID(id)
		if err != nil {
			return renamed, err
		}
	}
	// remove all empty dirs
	err = en.DeleteEmptyFolders(e.Config.LibraryRoot, e.UI)
	return
}
