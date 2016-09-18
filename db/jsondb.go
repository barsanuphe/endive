/*
Package db is the endive subpackage that implements the Database interface.
*/
package db

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/barsanuphe/endive/endive"
	"github.com/jhoonb/archivex"
)

// JSONDB implements endive.Database with a JSON backend.
type JSONDB struct {
	path string
}

// SetPath for database
func (db *JSONDB) SetPath(path string) {
	// TODO check if parent dir exists, create if necessary
	db.path = path
}

// Path of database
func (db *JSONDB) Path() string {
	return db.path
}

// Equals to another Database
func (db *JSONDB) Equals(o endive.Database) bool {
	// load local books into a collection
	var dbBooks endive.Collection
	if err := db.Load(dbBooks); err != nil {
		return false
	}

	var dbCollection endive.Collection
	dbCollection.Add(dbBooks.Books()...)

	// load books from the other database
	var oBooks endive.Collection
	if err := o.Load(oBooks); err != nil {
		return false
	}
	var oCollection endive.Collection
	oCollection.Add(oBooks.Books()...)
	// return true if diff of collections is blank
	newB, modB, delB := dbCollection.Diff(oCollection)
	if newB == nil && modB == nil && delB == nil {
		return true
	}

	return false
}

// Load database
func (db *JSONDB) Load(bks endive.Collection) error {
	jsonContent, err := ioutil.ReadFile(db.path)
	if err != nil {
		if os.IsNotExist(err) {
			// first run, it will be created later.
			return nil
		}
		return err
	}

	// load Books
	return json.Unmarshal(jsonContent, bks)
}

// Save database
func (db *JSONDB) Save(bks endive.Collection) (hasSaved bool, err error) {
	jsonToSave, err := json.Marshal(bks)
	if err != nil {
		return hasSaved, err
	}
	jsonInDB, err := ioutil.ReadFile(db.path)
	if err != nil && !os.IsNotExist(err) {
		return hasSaved, err
	}

	// if changes are detected, save
	if !bytes.Equal(jsonToSave, jsonInDB) {
		err = ioutil.WriteFile(db.path, jsonToSave, 0777)
		if err != nil {
			return false, err
		}
		hasSaved = true
	}
	return hasSaved, nil
}

// Backup database
func (db *JSONDB) Backup(path string) error {
	// TODO check path does not exist, but parent dirs do

	// creating tarball
	tar := new(archivex.TarFile)
	if err := tar.Create(path); err != nil {
		return err
	}
	if err := tar.AddFile(db.path); err != nil {
		return err
	}
	return tar.Close()
}
