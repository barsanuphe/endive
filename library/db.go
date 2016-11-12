package library

import (
	"errors"
	"strconv"
	"time"

	b "github.com/barsanuphe/endive/book"
	e "github.com/barsanuphe/endive/endive"
)

// Backup current database.
func (l *Library) backup() (err error) {
	l.UI.Debug("Backup up database...")
	return l.DB.Backup(l.Config.LibraryRoot)
}

// Load current DB
func (l *Library) Load() error {
	l.UI.Debug("Loading database...")
	err := l.DB.Load(l.Collection)
	if err == nil {
		l.Collection.Propagate(l.UI, l.Config)
	}
	return err
}

// Save current DB
func (l *Library) Save() (hasSaved bool, err error) {
	l.UI.Debug("Determining if database should be saved...")
	// getting old contents for reference
	var oldBooks e.Collection
	oldBooks = &b.Books{}
	err = l.DB.Load(oldBooks)
	if err != nil {
		return
	}
	// saving
	hasSaved, err = l.DB.Save(l.Collection)
	if err != nil {
		return
	}
	if hasSaved {
		// index what is needed.
		// diff to check the changes
		var n, m, d e.Collection
		n = &b.Books{}
		m = &b.Books{}
		d = &b.Books{}
		l.Collection.Diff(oldBooks, n, m, d)

		// update the index
		err = l.Index.Update(n, m, d)
		if err != nil {
			l.UI.Error("Error updating index, it may be necessary to build it anew")
			return hasSaved, l.RebuildIndex()
		}
		l.UI.Debug("In index: " + strconv.FormatUint(l.Index.Count(), 10) + " epubs.")
	}
	return
}

// RebuildIndex from scratch if necessary
func (l *Library) RebuildIndex() error {
	defer e.TimeTrack(l.UI, time.Now(), "Indexing")
	f := func() error {
		return l.Index.Rebuild(l.Collection)
	}
	return e.SpinWhileThingsHappen("Indexing", f)
}

// CheckIndex from scratch if necessary
func (l *Library) CheckIndex() error {
	defer e.TimeTrack(l.UI, time.Now(), "Checking index")
	f := func() error {
		return l.Index.Check(l.Collection)
	}
	return e.SpinWhileThingsHappen("Checking index", f)
}

func (l *Library) checkBooks() error {
	for i := range l.Collection.Books() {
		l.UI.Debug("Checking " + l.Collection.Books()[i].String())
		retailChanged, nonRetailChanged, err := l.Collection.Books()[i].Check()
		if err != nil {
			return err
		}
		if retailChanged {
			err = errors.New("Retail epub has changed for book: " + l.Collection.Books()[i].String())
			return err
		}
		if nonRetailChanged {
			l.UI.Warning("Non-retail epub for book " + l.Collection.Books()[i].String() + " has changed, check if this is normal.")
		}
	}
	return nil
}

// Check all Books
func (l *Library) Check() error {
	defer e.TimeTrack(l.UI, time.Now(), "Checking")
	f := func() error {
		return l.checkBooks()
	}
	return e.SpinWhileThingsHappen("Checking", f)
}
