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
func (s *IndexService) Rebuild(all endive.Collection) error {
	fmt.Println("mock Index: Rebuild")
	return nil
}

// Update for mock Indexer
func (s *IndexService) Update(new endive.Collection, mod endive.Collection, del endive.Collection) error {
	fmt.Println("mock Index: Update")
	return nil
}

// Check for mock Indexer
func (s *IndexService) Check(all endive.Collection) error {
	fmt.Println("mock Index: Check")
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
