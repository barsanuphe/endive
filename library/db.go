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
	// generate archive filename with date.
	archiveName, err := e.GetArchiveUniqueName(l.DB.Path())
	if err != nil {
		return
	}
	return l.DB.Backup(archiveName)
}

// updateBooks to be aware of Config and UI.
func (l *Library) updateBooks() {
	// make each Book aware of current Config + UI
	for i := range l.Books {
		l.Books[i].Config = l.Config
		l.Books[i].UI = l.UI
		l.Books[i].NonRetailEpub.Config = l.Config
		l.Books[i].NonRetailEpub.UI = l.UI
		l.Books[i].RetailEpub.Config = l.Config
		l.Books[i].RetailEpub.UI = l.UI
	}
	return
}

// Load current DB
func (l *Library) Load() error {
	l.UI.Debug("Loading database...")
	var c e.Collection
	c = &l.Books
	err := l.DB.Load(c)
	if err == nil {
		l.updateBooks()
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
	var c e.Collection
	c = &l.Books
	hasSaved, err = l.DB.Save(c)
	if err != nil {
		return
	}
	if hasSaved {
		// index what is needed.
		// diff to check the changes
		n, m, d := l.Books.Diff(oldBooks)

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
		// convert to Collection
		var c e.Collection
		c = &l.Books
		return l.Index.Rebuild(c)
	}
	return e.SpinWhileThingsHappen("Indexing", f)
}

// CheckIndex from scratch if necessary
func (l *Library) CheckIndex() error {
	defer e.TimeTrack(l.UI, time.Now(), "Checking index")
	f := func() error {
		// convert to Collection
		var c e.Collection
		c = &l.Books
		return l.Index.Check(c)
	}
	return e.SpinWhileThingsHappen("Checking index", f)
}

func (l *Library) checkBooks() error {
	for i := range l.Books {
		l.UI.Debug("Checking " + l.Books[i].ShortString())
		retailChanged, nonRetailChanged, err := l.Books[i].Check()
		if err != nil {
			return err
		}
		if retailChanged {
			err = errors.New("Retail epub has changed for book: " + l.Books[i].ShortString())
			return err
		}
		if nonRetailChanged {
			l.UI.Warning("Non-retail epub for book " + l.Books[i].ShortString() + " has changed, check if this is normal.")
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

// RemoveByID a book from the db
func (l *Library) RemoveByID(id int) (err error) {
	var found bool
	removeIndex := -1
	for i := range l.Books {
		if l.Books[i].ID == id {
			found = true
			removeIndex = i
			break
		}
	}
	if found {
		l.UI.Info("REMOVING from db " + l.Books[removeIndex].ShortString())
		l.Books = append((l.Books)[:removeIndex], (l.Books)[removeIndex+1:]...)
	} else {
		err = errors.New("Did not find book with ID " + strconv.Itoa(id))
	}
	return
}

// ListNonRetailOnly among known epubs.
func (l *Library) ListNonRetailOnly() b.Books {
	return l.Books.FilterNonRetailOnly()
}

// ListRetail among known epubs.
func (l *Library) ListRetail() b.Books {
	return l.Books.FilterRetail()
}

// ListAuthors among known epubs.
func (l *Library) ListAuthors() (authors map[string]int) {
	authors = make(map[string]int)
	for _, book := range l.Books {
		author := book.Metadata.Author()
		authors[author]++
	}
	return
}

// ListPublishers among known epubs.
func (l *Library) ListPublishers() (publishers map[string]int) {
	publishers = make(map[string]int)
	for _, book := range l.Books {
		if book.Metadata.Publisher != "" {
			publisher := book.Metadata.Publisher
			publishers[publisher]++
		} else {
			publishers["Unknown"]++
		}

	}
	return
}

// ListTags associated with known epubs.
func (l *Library) ListTags() (tags map[string]int) {
	tags = make(map[string]int)
	for _, book := range l.Books {
		for _, tag := range book.Metadata.Tags {
			tags[tag.Name]++
		}
	}
	return
}

// ListSeries associated with known epubs.
func (l *Library) ListSeries() (series map[string]int) {
	series = make(map[string]int)
	for _, book := range l.Books {
		for _, s := range book.Metadata.Series {
			series[s.Name]++
		}
	}
	return
}

// ListUntagged among known epubs.
func (l *Library) ListUntagged() b.Books {
	return l.Books.FilterUntagged()
}

// ListByProgress returns a slice of Books with the given reading progress.
func (l *Library) ListByProgress(progress string) b.Books {
	return l.Books.FilterByProgress(progress)
}

// ListIncomplete among known epubs.
func (l *Library) ListIncomplete() b.Books {
	return l.Books.FilterIncomplete()
}
