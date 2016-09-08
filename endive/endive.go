package endive

// GenericBook interface for Books
type GenericBook interface {
	FullPath() string
	ShortString() string
}

// GenericBooks interface for slices of Books
type GenericBooks interface {
	FindByID(string) (*GenericBook, error)
	FindByHash(string) (*GenericBook, error)
	FindByMetadata(string) (*GenericBook, error)
	FindByFullPath(string) (*GenericBook, error)
	Books() []GenericBook
}

// Indexer provides an interface ofr indexing books.
type Indexer interface {
	SetPath(path string)
	Rebuild(all []GenericBook) error
	Update(new map[string]GenericBook, mod map[string]GenericBook, del map[string]GenericBook) error
	Query(query string) ([]string, error)
	Count() uint64
}

// TODO: Check(all []GenericBook) error : checks all books have an entry + no extraneous ones
