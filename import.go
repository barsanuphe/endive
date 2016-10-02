package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/tj/go-spin"

	en "github.com/barsanuphe/endive/endive"
)

//------------------------------------

type Candidate struct {
	filename           string
	hash               string
	imported           bool
	importedButMissing bool
}

func (c Candidate) String() string {
	return fmt.Sprintf("Candidate: %s | %s | %t | %t", c.filename, c.hash, c.imported, c.importedButMissing)
}

func NewCandidate(filename string, knownHashes en.KnownHashes, collection en.Collection) *Candidate {

	fmt.Println("New Candidate: " + filename)
	// calculate hash
	hash, err := en.CalculateSHA256(filename)
	if err != nil {
		return nil
	}
	var imported, importedButMissing bool
	// find if in known_hashes
	if knownHashes.IsIn(hash) {
		imported = true
	}

	// TODO if it is, try to find in collection.

	// build and return *Candidate with all fields
	return &Candidate{filename: filename, hash: hash, imported: imported, importedButMissing: importedButMissing}
}

//------------------------------------

type Candidates []Candidate

func (c *Candidates) New() Candidates {
	res := Candidates{}
	for _, e := range *c {
		if !e.imported {
			res = append(res, e)
		}
	}
	return res
}

func (c *Candidates) Missing() Candidates {
	res := Candidates{}
	for _, e := range *c {
		if e.imported && e.importedButMissing {
			res = append(res, e)
		}
	}
	return res
}

//------------------------------------

// listepubs ne fait plus que lister les epubs
// ListEpubs recursively.
func ListEpubs(root string) (epubPaths []string, err error) {
	if !en.DirectoryExists(root) {
		err = errors.New("Directory " + root + " does not exist")
		return
	}
	// spinner, defaults to s.Set(spin.Box1)
	s := spin.New()
	cpt := 0
	filepath.Walk(root, func(path string, f os.FileInfo, err error) (outErr error) {
		// only consider epub files
		if f.Mode().IsRegular() && filepath.Ext(path) == ".epub" {
			epubPaths = append(epubPaths, path)
			// show progress
			if cpt%10 == 0 {
				fmt.Printf("\rSearching %s ", s.Next())
			}
			cpt++
		}
		return
	})
	fmt.Print("\r")
	return
}

func GetCandidates(root string, known_hashes en.KnownHashes, collection en.Collection) ([]Candidate, error) {
	// list epubs
	paths, err := ListEpubs(root)
	if err != nil {
		return []Candidate{}, err
	}

	// for all epubs, build candidate
	jobs := make(chan string, len(paths))
	results := make(chan *Candidate, len(paths))

	// This starts up as many workers as there are detected cpus
	for w := 1; w <= runtime.NumCPU(); w++ {
		go func(id int, jobs <-chan string, results chan<- *Candidate) {
			for j := range jobs {
				fmt.Println("worker", id, "processing job", j)
				results <- NewCandidate(j, known_hashes, collection)
			}
		}(w, jobs, results)
	}

	// sending paths
	for _, p := range paths {
		jobs <- p
	}
	close(jobs)

	// Finally we collect all the results of the work.
	candidates := []Candidate{}
	for range paths {
		c := <- results
		candidates = append(candidates, *c)
	}
	return candidates, nil
}
