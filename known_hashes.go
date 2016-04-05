package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

// KnownHashes keeps track of the hashes of already imported epubs.
type KnownHashes struct {
	Filename string   `json:"-"`
	Hashes   []string `json:"hashes"`
	Count    int      `json:"-"`
}

// Load the known hashes.
func (k *KnownHashes) Load() (err error) {
	fmt.Println("Loading known hashes database.")
	hashesBytes, err := ioutil.ReadFile(k.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			// first run
			return nil
		}
		return
	}
	err = json.Unmarshal(hashesBytes, k)
	if err != nil {
		fmt.Println(err)
		return
	}
	k.Count = len(k.Hashes)
	return
}

// Save the known hashes database.
func (k *KnownHashes) Save() (modified bool, err error) {
	// check if hashes have been added
	if k.Count != len(k.Hashes) {
		modified = true
		hashesJSON, err := json.Marshal(k)
		if err != nil {
			fmt.Println(err)
			return modified, err
		}
		// writing db
		fmt.Println("Saving known hashes database.")
		err = ioutil.WriteFile(k.Filename, hashesJSON, 0777)
		if err != nil {
			return modified, err
		}
	}
	return
}

// Add a hash to the database.
func (k *KnownHashes) Add(hash string) (added bool, err error) {
	// TODO check if valid hash? verif taille par ex
	if len(hash) != 64 {
		return false, errors.New("SHA256 hash should be 64 characters long, not " + strconv.Itoa(len(hash)))
	}
	// append if not IsIn
	if !k.IsIn(hash) {
		k.Hashes = append(k.Hashes, hash)
		added = true
	}
	return
}

// IsIn checks whether a hash is known.
func (k *KnownHashes) IsIn(hash string) (isIn bool) {
	if len(hash) != 64 {
		// no need to check
		return
	}
	_, isIn = stringInSlice(hash, k.Hashes)
	return
}
