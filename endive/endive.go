/*
Package endive is the root package of endive.

It defines the main interfaces.
It manages the configuration file and also deals with the internal database of
already imported files (tracked through their SHA256 hashes).
It also contains functions that are used by other endive subpackages.
*/
package endive

// EpubExtension is the lowercase extension for all epubs.
const EpubExtension = ".epub"

// GenericBook interface for Books
type GenericBook interface {
	ID() int
	HasEpub() bool
	FullPath() string
	ShortString() string
	CleanFilename() string
	Refresh() ([]bool, []string, error)
	AddEpub(string, bool, string) (bool, error)
	Check() (bool, bool, error)
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
	Propagate(UserInterface, Config)
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
	Untagged() Collection
	Progress(string) Collection
	Incomplete() Collection
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

// UserInterface deals with user input, output and logging.
type UserInterface interface {
	// input
	GetInput() (string, error)
	Accept(string) bool
	Choose(string, string, string, string) (string, error)
	UpdateValues(string, string, []string, bool) ([]string, error)
	Edit(string) (string, error)
	// output
	Title(string, ...interface{})
	SubTitle(string, ...interface{})
	SubPart(string, ...interface{})
	Choice(string, ...interface{})
	Display(string)
	// log
	InitLogger(string) error
	CloseLog()
	Error(string)
	Errorf(string, ...interface{})
	Warning(string)
	Warningf(string, ...interface{})
	Info(string)
	Infof(string, ...interface{})
	Debug(string)
	Debugf(string, ...interface{})
}
