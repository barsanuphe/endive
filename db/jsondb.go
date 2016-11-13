/*
Package db is the endive subpackage that implements the Database interface.

The current implementation saves all Book information as a simple JSON file.

That makes it:
- easy to index with bleve
- easy to check and, if desperate, edit for a human being
- easy to version with git

*/
package db

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/libgit2/git2go"

	"github.com/barsanuphe/endive/endive"
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
	jsonContent, err1 := ioutil.ReadFile(db.path)
	ojsonContent, err2 := ioutil.ReadFile(o.Path())
	if err1 != nil || err2 != nil {
		if os.IsNotExist(err1) && os.IsNotExist(err2) {
			return true
		}
		return false
	}
	return bytes.Equal(jsonContent, ojsonContent)
}

// Load database into a Collection
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

// Save database as a JSON file
func (db *JSONDB) Save(bks endive.Collection) (hasSaved bool, err error) {
	// Marshal into json with pretty print.
	// Use json.Marshal(bks) for more compressed format.
	jsonToSave, err := json.MarshalIndent(bks, "", "    ")
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

// Backup JSON database by versioning it in a git repository.
func (db *JSONDB) Backup(path string) error {
	// use git
	firstCommit := false
	repo, err := git.OpenRepository(path)
	if err != nil {
		repo, err = git.InitRepository(path, false)
		if err != nil {
			return err
		}
		firstCommit = true
	}
	index, err := repo.Index()
	if err != nil {
		return err
	}
	if err := index.AddByPath(filepath.Base(db.path)); err != nil {
		return err
	}
	treeID, err := index.WriteTree()
	if err != nil {
		return err
	}
	if err := index.Write(); err != nil {
		return err
	}
	tree, err := repo.LookupTree(treeID)
	if err != nil {
		return err
	}

	signature := &git.Signature{
		Name:  "endive",
		Email: "endive@endive.com",
		When:  time.Now(),
	}
	message := "endive automatic commit."
	if firstCommit {
		_, err = repo.CreateCommit("HEAD", signature, signature, message, tree)
	} else {
		head, err := repo.Head()
		if err != nil {
			return err
		}
		headCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}
		_, err = repo.CreateCommit("HEAD", signature, signature, message, tree, headCommit)
	}
	return err
}
