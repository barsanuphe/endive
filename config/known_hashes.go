package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	h "github.com/barsanuphe/endive/helpers"

	"launchpad.net/go-xdg"
)

const (
	hashes        = "endive_hashes"
	xdgHashesPath = Endive + "/" + hashes + ".json"
)

// KnownHashes keeps track of the hashes of already imported epubs.
type KnownHashes struct {
	Filename string   `json:"-"`
	Hashes   []string `json:"hashes"`
	Count    int      `json:"-"`
}

// GetKnownHashesPath gets the default path for known hashes.
func GetKnownHashesPath() (hashesFile string, err error) {
	hashesFile, err = xdg.Data.Find(xdgHashesPath)
	if err != nil {
		hashesFile, err = xdg.Data.Ensure(xdgHashesPath)
		if err != nil {
			return
		}
		// making sure it's a valid JSON file for next load
		err = ioutil.WriteFile(hashesFile, []byte("{}"), 0777)
		if err != nil {
			return
		}
		h.Logger.Debug("Known hashes file", hashesFile, "created.")
	}
	return
}

// Load the known hashes.
func (k *KnownHashes) Load() (err error) {
	h.Logger.Debug("Loading known hashes database.")
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
		h.Logger.Debug("Saving known hashes database.")
		err = ioutil.WriteFile(k.Filename, hashesJSON, 0777)
		if err != nil {
			return modified, err
		}
	}
	return
}

// Add a hash to the database.
func (k *KnownHashes) Add(hash string) (added bool, err error) {
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
	_, isIn = h.StringInSlice(hash, k.Hashes)
	return
}
