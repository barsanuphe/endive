/*
Package endive is the root package of endive.

It defines the main interfaces.
It manages the configuration file and also deals with the internal database of
already imported files (tracked through their SHA256 hashes).
It also contains functions that are used by other endive subpackages.
*/
package endive

import (
	i "github.com/barsanuphe/helpers/ui"
)

const (
	// EpubExtension is the lowercase extension for all epubs.
	EpubExtension = ".epub"
	// True as string
	True = "true"
	// False as string
	False = "false"
	// EmptyIndexError for Indexer
	EmptyIndexError = "Index is empty"
)

// GenericBook interface for Books
type GenericBook interface {
	ID() int
	HasHash(string) bool
	HasEpub() bool
	FullPath() string
	String() string
	CleanFilename() string
	Refresh() ([]bool, []string, error)
	AddEpub(string, bool, string) (bool, error)
	Check() (bool, bool, error)
	SetExported(bool)
}

// Indexer provides an interface for indexing books.
type Indexer interface {
	SetPath(path string)
	Rebuild(Collection) error
	Update(Collection, Collection, Collection) error
	Check(Collection) error
	Query(query string) ([]string, error)
	Count() uint64
}

// Collection interface for slices of Books
type Collection interface {
	// contents
	Books() []GenericBook
	Add(...GenericBook)
	Propagate(i.UserInterface, Config)
	RemoveByID(int) error
	Diff(Collection, Collection, Collection, Collection)
	// 	Check() error
	// search
	FindByID(int) (GenericBook, error)
	FindByHash(string) (GenericBook, error)
	FindByMetadata(string, string, string) (GenericBook, error)
	FindByFullPath(string) (GenericBook, error)
	// extracting information
	Retail() Collection
	NonRetailOnly() Collection
	Exported() Collection
	Progress(string) Collection
	Incomplete() Collection
	WithID(...int) Collection
	Authors() map[string]int
	Publishers() map[string]int
	Tags() map[string]int
	Series() map[string]int
	// output
	Table() string
	Sort(string)
	First(int) Collection
	Last(int) Collection
}

// Database is the interface for loading/saving Book information
type Database interface {
	SetPath(string)
	Path() string
	Equals(Database) bool
	Load(Collection) error
	Save(Collection) (bool, error)
	Backup(string) error
}
