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

// Propagate implementation for tests
func (c *Collection) Propagate(u endive.UserInterface, cfg endive.Config) {
	fmt.Println("mock Collection: Propagate")
}


// RemoveByID implementation for tests
func (c *Collection) RemoveByID(u int) error {
	fmt.Println("mock Collection: RemoveByID")
	return nil
}

// Diff implementation for tests
func (c *Collection) Diff(endive.Collection, endive.Collection, endive.Collection, endive.Collection) {
	fmt.Println("mock Collection: Diff")
}

// FindByID implementation for tests
func (c *Collection) FindByID(id int) (endive.GenericBook, error) {
	fmt.Printf("mock Collection: FindByID : %d\n", id)
	return nil, nil
}

// FindByHash implementation for tests
func (c *Collection) FindByHash(string) (endive.GenericBook, error) {
	fmt.Println("mock Collection: FindByHash")
	return nil, nil
}

// FindByMetadata implementation for tests
func (c *Collection) FindByMetadata(a, b, d string) (endive.GenericBook, error) {
	fmt.Println("mock Collection: FindByMetadata")
	return nil, nil
}

// FindByFullPath implementation for tests
func (c *Collection) FindByFullPath(string) (endive.GenericBook, error) {
	fmt.Println("mock Collection: FindByFullPath")
	return nil, nil
}

// Retail implementation for tests
func (c *Collection) Retail() endive.Collection {
	fmt.Println("mock Collection: Retail")
	return nil
}

// NonRetailOnly implementation for tests
func (c *Collection) NonRetailOnly() endive.Collection {
	fmt.Println("mock Collection: NonRetailOnly")
	return nil
}

// Progress implementation for tests
func (c *Collection) Progress(p string) endive.Collection {
	fmt.Println("mock Collection: Progress " + p)
	return nil
}

// Incomplete implementation for tests
func (c *Collection) Incomplete() endive.Collection {
	fmt.Println("mock Collection: Incomplete")
	return nil
}

// Authors implementation for tests
func (c *Collection) Authors() map[string]int {
	fmt.Println("mock Collection: Authors")
	return nil
}

// Publishers implementation for tests
func (c *Collection) Publishers() map[string]int {
	fmt.Println("mock Collection: Publishers")
	return nil
}

// Tags implementation for tests
func (c *Collection) Tags() map[string]int {
	fmt.Println("mock Collection: Tags")
	return nil
}

// Series implementation for tests
func (c *Collection) Series() map[string]int {
	fmt.Println("mock Collection: Series")
	return nil
}

// Untagged implementation for tests
func (c *Collection) Untagged() endive.Collection {
	fmt.Println("mock Collection: Untagged")
	return nil
}

// Table implementation for tests
func (c *Collection) Table() string {
	fmt.Println("mock Collection: Table")
	return "TABLE"
}

// Sort implementation for tests
func (c *Collection) Sort(sortBy string) {
	fmt.Println("mock Collection: Sort " + sortBy)
}

// First implementation for tests
func (c *Collection) First(nb int) endive.Collection {
	fmt.Printf("mock Collection: First %d\n", nb)
	return nil
}

// Last implementation for tests
func (c *Collection) Last(nb int) endive.Collection {
	fmt.Printf("mock Collection: Last %d\n", nb)
	return nil
}

