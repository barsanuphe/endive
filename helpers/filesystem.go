package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tj/go-spin"
)

// DirectoryExists checks if a directory exists.
func DirectoryExists(path string) (res bool) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		return true
	}
	return
}

// IsDirectoryEmpty checks if files are present in directory.
func IsDirectoryEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// check if at least one file inside
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// AbsoluteFileExists checks if an absolute path is an existing file.
func AbsoluteFileExists(path string) (res bool) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.Mode().IsRegular() {
		return true
	}
	return
}

// FileExists checks if a path is valid and returns its absolute path
func FileExists(path string) (absolutePath string, err error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return
	}
	candidate := ""
	if filepath.IsAbs(path) {
		candidate = path
	} else {
		candidate = filepath.Join(currentDir, path)
	}

	if AbsoluteFileExists(candidate) {
		absolutePath = candidate
	} else {
		err = errors.New("File does not exist")
	}
	return
}

// DeleteEmptyFolders deletes empty folders that may appear after sorting albums.
func DeleteEmptyFolders(root string) (err error) {
	defer TimeTrack(time.Now(), "Scanning files")

	Debugf("Scanning for empty directories.\n\n")
	deletedDirectories := 0
	deletedDirectoriesThisTime := 0
	atLeastOnce := false

	// loops until all levels of empty directories are deleted
	for !atLeastOnce || deletedDirectoriesThisTime != 0 {
		atLeastOnce = true
		deletedDirectoriesThisTime = 0
		err = filepath.Walk(root, func(path string, fileInfo os.FileInfo, walkError error) (err error) {
			if path == root {
				// do not delete root, even if empty
				return
			}
			// when an directory has just been removed, Walk goes through it a second
			// time with an "file does not exist" error
			if os.IsNotExist(walkError) {
				return
			}
			if fileInfo.IsDir() {
				isEmpty, err := IsDirectoryEmpty(path)
				if err != nil {
					panic(err)
				}
				if isEmpty {
					Debugf("Removing empty directory ", path)
					if err := os.Remove(path); err == nil {
						deletedDirectories++
						deletedDirectoriesThisTime++
					}
				}
			}
			return
		})
		if err != nil {
			fmt.Printf("Error!")
		}
	}

	Debugf("\n### Removed %d albums.\n", deletedDirectories)
	return
}

// ListEpubsInDirectory recursively.
func ListEpubsInDirectory(root string) (epubPaths []string, hashes []string, err error) {
	if !DirectoryExists(root) {
		err = errors.New("Directory " + root + " does not exist")
		return
	}

	// spinner, defaults to s.Set(spin.Box1)
	s := spin.New()
	cpt := 0

	filepath.Walk(root, func(path string, f os.FileInfo, err error) (outErr error) {
		// only consider epub files
		if f.Mode().IsRegular() && filepath.Ext(path) == ".epub" {
			// check if already imported
			// calculate hash
			hash, err := CalculateSHA256(path)
			if err != nil {
				return
			}
			epubPaths = append(epubPaths, path)
			hashes = append(hashes, hash)
			// show progress
			if cpt%10 == 0 {
				fmt.Printf("\rSearching %s ", s.Next())
			}
			cpt++
		}
		return
	})
	fmt.Printf("\r")
	return
}

// CleanPath makes sure a string can be used as part of a path
func CleanPath(md string) string {
	md = strings.TrimSpace(md)
	// if it starts with a dot, remove it so it does not become
	// a hidden file. if it starts with /, weird things happen.
	if strings.HasPrefix(md, ".") || strings.HasPrefix(md, "/") {
		md = md[1:]
	}
	// clean characters which would be problematic in a filename
	r := strings.NewReplacer(
		"/", "-",
		"\\", "-",
	)
	return r.Replace(md)
}

// CleanPathForVFAT makes sure a string can be used as part of a path
func CleanPathForVFAT(md string) string {
	// clean characters which would be problematic in a filename
	r := strings.NewReplacer(
		":", "-",
		"?", "",
	)
	return r.Replace(md)
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// CalculateSHA256 calculates an epub's current hash
func CalculateSHA256(filename string) (hash string, err error) {
	var result []byte
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	hashBytes := sha256.New()
	_, err = io.Copy(hashBytes, file)
	if err != nil {
		return
	}
	hash = hex.EncodeToString(hashBytes.Sum(result))
	return
}

// GetUniqueTimestampedFilename for a given filename.
func GetUniqueTimestampedFilename(dir, filename string) (uniqueFilename string, err error) {
	// create dir if necessary
	if !DirectoryExists(dir) {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return
		}
	}
	// Mon Jan 2 15:04:05 -0700 MST 2006
	currentTime := time.Now().Local()
	uniqueNameFound := false
	ext := filepath.Ext(filename)
	filenameBase := strings.TrimSuffix(filepath.Base(filename), ext)
	attempts := 0
	for !uniqueNameFound || attempts > 50 {
		suffix := ""
		if attempts > 0 {
			suffix = fmt.Sprintf("_%d", attempts)
		}
		candidate := fmt.Sprintf("%s - %s%s.tar.gz", currentTime.Format("2006-01-02 15:04:05"), filenameBase, suffix)
		// While candidate already exists, change suffix.
		_, err := FileExists(filepath.Join(dir, candidate))
		if err != nil {
			// file not found
			uniqueFilename = filepath.Join(dir, candidate)
			uniqueNameFound = true
		} else {
			attempts++
		}
	}
	return
}
