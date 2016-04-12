package main

import (
	"os"
	"path/filepath"
)

// EpubFile can manipulate an epub file.
type Epub struct {
	Filename         string `json:"filename"` // relative to LibraryRoot
	Config           Config `json:"-"`
	Hash             string `json:"hash"`
	NeedsReplacement string `json:"replace"`
}

// getPath returns the absolute file path.
// if it is in the library, prepends LibraryRoot.
// if it is outside, return Filename directly.
func (e *Epub) getPath() (path string) {
	// TODO: tests
	if filepath.IsAbs(e.Filename) {
		return e.Filename
	} else {
		return filepath.Join(e.Config.LibraryRoot, e.Filename)
	}
}

// GetHash calculates an epub's current hash
func (e *Epub) GetHash() (err error) {
	hash, err := calculateSHA256(e.getPath())
	if err != nil {
		return
	}
	e.Hash = hash
	return
}

// FlagForReplacement an epub of insufficient quality
func (e *Epub) FlagForReplacement() (err error) {
	e.NeedsReplacement = "true"
	return
}

// SetRetail a retail epub ebook.
func (e *Epub) SetRetail() (err error) {
	// set read-only
	err = os.Chmod(e.getPath(), 0444)
	return
}

// SetNonRetail a non retail epub ebook.
func (e *Epub) SetNonRetail() (err error) {
	// set read-write
	err = os.Chmod(e.getPath(), 0777)
	return
}

// Check the retail epub integrity.
func (e *Epub) Check() (hasChanged bool, err error) {
	// get current hash
	currentHash, err := calculateSHA256(e.getPath())
	if err != nil {
		return
	}
	// compare with old
	if currentHash != e.Hash {
		hasChanged = true
	}
	return
}
