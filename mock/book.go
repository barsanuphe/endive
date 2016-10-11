package mock

import "fmt"

// Book represents a mock implementation of endive.GenericBook.
type Book struct {
	// additional function implementations...
}

// ID implementation for tests
func (b *Book) ID() int {
	fmt.Println("mock Book: ID")
	return 1
}

// HasEpub implementation for tests
func (b *Book) HasEpub() bool {
	fmt.Println("mock Book: HasEpub")
	return true
}

// FullPath implementation for tests
func (b *Book) FullPath() string {
	fmt.Println("mock Book: FullPath")
	return "/tmp/book.epub"
}

// ShortString implementation for tests
func (b *Book) ShortString() string {
	fmt.Println("mock Book: ShortString")
	return "Author (YEAR) Title"
}

// CleanFilename implementation for tests
func (b *Book) CleanFilename() string {
	fmt.Println("mock Book: CleanFilename")
	return "Author (YEAR) Title.epub"
}

// Refresh implementation for tests
func (b *Book) Refresh() ([]bool, []string, error) {
	fmt.Println("mock Book: Refresh")
	return []bool{}, []string{}, nil
}

// AddEpub implementation for tests
func (b *Book) AddEpub(path string, isRetail bool, hash string) (bool, error) {
	fmt.Println("mock Book: AddEpub " + path)
	return true, nil
}

// Check implementation for tests
func (b *Book) Check() (bool, bool, error) {
	fmt.Println("mock Book: Check")
	return false, false, nil
}

// SetExported implementation for tests
func (b *Book) SetExported(bool) {
	fmt.Println("mock Book: SetExported")
}
