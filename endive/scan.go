package endive

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tj/go-spin"
)

// EpubCandidate for import/export/refresh
type EpubCandidate struct {
	Filename           string
	Hash               string
	Imported           bool
	ImportedButMissing bool
}

// String representation for EpubCandidate
func (c EpubCandidate) String() string {
	return fmt.Sprintf("Candidate: %s | %s | %t | %t", c.Filename, c.Hash, c.Imported, c.ImportedButMissing)
}

// NewCandidate returns a filled epubCandidate struct
func NewCandidate(filename string, knownHashes KnownHashes, collection Collection) *EpubCandidate {
	// calculate hash
	hash, err := CalculateSHA256(filename)
	if err != nil {
		return nil
	}
	var imported, importedButMissing bool
	// find if in known_hashes
	if knownHashes.IsIn(hash) {
		imported = true
		// if it is, try to find in collection.
		if collection != nil {
			if _, err := collection.FindByHash(hash); err != nil {
				importedButMissing = true
			}
		}
	}
	// build and return *Candidate with all fields
	return &EpubCandidate{Filename: filename, Hash: hash, Imported: imported, ImportedButMissing: importedButMissing}
}

// EpubCandidates is a slice of EpubCandidates.
type EpubCandidates []EpubCandidate

// New EpubCandidate-s in EpubCandidates
func (c EpubCandidates) New() EpubCandidates {
	res := EpubCandidates{}
	for _, e := range c {
		if !e.Imported {
			res = append(res, e)
		}
	}
	return res
}

// Missing EpubCandidate-s in EpubCandidates
func (c EpubCandidates) Missing() EpubCandidates {
	res := EpubCandidates{}
	for _, e := range c {
		if e.Imported && e.ImportedButMissing {
			res = append(res, e)
		}
	}
	return res
}

// Importable EpubCandidate-s in EpubCandidates
func (c EpubCandidates) Importable() EpubCandidates {
	res := EpubCandidates{}
	for _, e := range c {
		if !e.Imported || e.ImportedButMissing {
			res = append(res, e)
		}
	}
	return res
}

//------------------------------------

// listEpubs recursively.
func listEpubs(root string) (epubPaths []string, err error) {
	if !DirectoryExists(root) {
		err = errors.New("Directory " + root + " does not exist")
		return
	}
	// spinner, defaults to s.Set(spin.Box1)
	s := spin.New()
	cpt := 0
	filepath.Walk(root, func(path string, f os.FileInfo, err error) (outErr error) {
		// only consider epub files
		if f.Mode().IsRegular() && filepath.Ext(path) == EpubExtension {
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

// ScanForEpubs recursively in folder.
func ScanForEpubs(root string, knownHashes KnownHashes, collection Collection) ([]EpubCandidate, error) {
	// list epubs
	paths, err := listEpubs(root)
	if err != nil {
		return []EpubCandidate{}, err
	}

	s := spin.New()
	cpt := 0
	// for all epubs, build candidate
	jobs := make(chan string, len(paths))
	results := make(chan *EpubCandidate, len(paths))

	// This starts up as many workers as there are detected cpus
	for w := 1; w <= 25; w++ {
		go func(id int, jobs <-chan string, results chan<- *EpubCandidate) {
			for j := range jobs {
				results <- NewCandidate(j, knownHashes, collection)
			}
		}(w, jobs, results)
	}

	// sending paths
	for _, p := range paths {
		jobs <- p
	}
	close(jobs)

	// Finally we collect all the results of the work.
	candidates := []EpubCandidate{}
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
