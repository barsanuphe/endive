package mock

import "fmt"

// DB represents a mock implementation of endive.Database.
type DB struct {
	// additional function implementations...
}

// SetPath for mock Database
func (db *DB) SetPath(path string) {
	fmt.Println("mock DB: SetPath " + path)
}

// Equals for mock Database
func (db *DB) Equals(o DB) bool {
	fmt.Println("mock DB: Equals")
	return false
}

// Load for mock Database
func (db *DB) Load(Collection) error {
	fmt.Println("mock DB: Load")
	return nil
}

// Save for mock Database
func (db *DB) Save(c Collection) (bool, error) {
	fmt.Println("mock DB: Save")
	return true, nil
}

// Backup for mock Database
func (db *DB) Backup(path string) error {
	fmt.Println("mock DB: Backup to " + path)
	return nil
}
