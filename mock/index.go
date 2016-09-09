package mock

import (
	"fmt"

	"github.com/barsanuphe/endive/endive"
)

// IndexService represents a mock implementation of endive.Indexer.
type IndexService struct {
	// additional function implementations...
}

// SetPath for mock Indexer
func (s *IndexService) SetPath(path string) {
	fmt.Println("mock Index: setPath" + path)
}

// Rebuild for mock Indexer
func (s *IndexService) Rebuild(all []endive.GenericBook) error {
	fmt.Println("mock Index: Rebuild")
	return nil
}

// Update for mock Indexer
func (s *IndexService) Update(new map[string]endive.GenericBook, mod map[string]endive.GenericBook, del map[string]endive.GenericBook) error {
	fmt.Println("mock Index: Update")
	return nil
}

// Query for mock Indexer
func (s *IndexService) Query(query string) ([]string, error) {
	fmt.Println("mock Index: Runquery")
	return []string{}, nil
}

// Count for mock Indexer
func (s *IndexService) Count() uint64 {
	fmt.Println("mock Index: Count")
	return 42
}
