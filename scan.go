package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tj/go-spin"

	en "github.com/barsanuphe/endive/endive"
)

// epubCandidate for import/export/refresh
type epubCandidate struct {
	filename           string
	hash               string
	imported           bool
	importedButMissing bool
}

func (c epubCandidate) String() string {
	return fmt.Sprintf("Candidate: %s | %s | %t | %t", c.filename, c.hash, c.imported, c.importedButMissing)
}

// newCandidate returns a filled epubCandidate struct
func newCandidate(filename string, knownHashes en.KnownHashes, collection en.Collection) *epubCandidate {
	// calculate hash
	hash, err := en.CalculateSHA256(filename)
	if err != nil {
		return nil
	}
	var imported, importedButMissing bool
	// find if in known_hashes
	if knownHashes.IsIn(hash) {
		imported = true
		// if it is, try to find in collection.
		if _, err := collection.FindByHash(hash); err != nil {
			importedButMissing = true
		}
	}
	// build and return *Candidate with all fields
	return &epubCandidate{filename: filename, hash: hash, imported: imported, importedButMissing: importedButMissing}
}

type epubCandidates []epubCandidate

func (c epubCandidates) new() epubCandidates {
	res := epubCandidates{}
	for _, e := range c {
		if !e.imported {
			res = append(res, e)
		}
	}
	return res
}

func (c epubCandidates) missing() epubCandidates {
	res := epubCandidates{}
	for _, e := range c {
		if e.imported && e.importedButMissing {
			res = append(res, e)
		}
	}
	return res
}

func (c epubCandidates) importable() epubCandidates {
	res := epubCandidates{}
	for _, e := range c {
		if !e.imported || e.importedButMissing {
			res = append(res, e)
		}
	}
	return res
}

//------------------------------------

// listEpubs recursively.
func listEpubs(root string) (epubPaths []string, err error) {
	if !en.DirectoryExists(root) {
		err = errors.New("Directory " + root + " does not exist")
		return
	}
	// spinner, defaults to s.Set(spin.Box1)
	s := spin.New()
	cpt := 0
	filepath.Walk(root, func(path string, f os.FileInfo, err error) (outErr error) {
		// only consider epub files
		if f.Mode().IsRegular() && filepath.Ext(path) == en.EpubExtension {
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

// getCandidates recursively in folder.
func getCandidates(root string, knownHashes en.KnownHashes, collection en.Collection) ([]epubCandidate, error) {
	// list epubs
	paths, err := listEpubs(root)
	if err != nil {
		return []epubCandidate{}, err
	}

	s := spin.New()
	cpt := 0
	// for all epubs, build candidate
	jobs := make(chan string, len(paths))
	results := make(chan *epubCandidate, len(paths))

	// This starts up as many workers as there are detected cpus
	for w := 1; w <= 25; w++ {
		go func(id int, jobs <-chan string, results chan<- *epubCandidate) {
			for j := range jobs {
				results <- newCandidate(j, knownHashes, collection)
			}
		}(w, jobs, results)
	}

	// sending paths
	for _, p := range paths {
		jobs <- p
	}
	close(jobs)

	// Finally we collect all the results of the work.
	candidates := []epubCandidate{}
	for range paths {
		c := <-results
		candidates = append(candidates, *c)
		// show progress
		if cpt%10 == 0 {
			fmt.Printf("\rAnalyzing %.2f%% %s ", float32(cpt)*100.0/float32(len(paths)), s.Next())
		}
		cpt++
	}
	fmt.Print("\r")
	return candidates, nil
}
