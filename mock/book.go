package mock

// Book represents a mock implementation of endive.GenericBook.
type Book struct {
	// additional function implementations...
}

// FullPath implementation for tests
func (b *Book) FullPath() string {
	return "/tmp/book.epub"
}

// ShortString implementation for tests
func (b *Book) ShortString() string {
	return "Author (YEAR) Title"
}
