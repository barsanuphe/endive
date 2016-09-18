package mock

import (
	"fmt"

	"github.com/barsanuphe/endive/endive"
)

// Collection represents a mock implementation of endive.Collection.
type Collection struct {
	// additional function implementations...
}

// Books implementation for tests
func (c *Collection) Books() []endive.GenericBook {
	fmt.Println("mock Collection: Books")
	return []endive.GenericBook{}
}

// Add implementation for tests
func (c *Collection) Add(books ...endive.GenericBook) {
	fmt.Println("mock Collection: Add")
}

// Diff implementation for tests
func (c *Collection) Diff(Collection) (Collection, Collection, Collection) {
	fmt.Println("mock Collection: Diff")
	return Collection{}, Collection{}, Collection{}
}
