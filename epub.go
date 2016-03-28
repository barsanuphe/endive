package main

import (
	"github.com/meskio/epubgo"
	"fmt"
)

// Series can track a series and an epub's position.
type Series struct {
	Name  string
	Index int
}

// Tag an ebook.
// TODO: at library level?
type Tag struct {
	Name        string
	Description string
}

// Epub can manipulate an epub file.
type Epub struct {
	Filename        string
	RelativePath    string
	NewFilename     string
	NewRelativePath string
	Hash            string
	IsRetail        bool
	Progress        int
	Series          []Series
	Author          string
	Title           string
	PublicationYear int
	ReadDate        string // month
	Tags            []Tag
}

// GetHash calculates an epub's current hash
func (e *Epub) GetHash() (err error) {
	return
}

// SetProgress sets reading progress
func (e *Epub) SetProgress(progress int) (err error) {
	return
}

// GetProgress sets reading progress
func (e *Epub) GetProgress() (progress int, err error) {
	return
}

// AddSeries adds a series
func (e *Epub) AddSeries(seriesName string, index int) (err error) {
	return
}

// RemoveSeries removes a series
func (e *Epub) RemoveSeries(seriesName string) (err error) {
	return
}

// HasSeries checks if epub is part of a series
func (e *Epub) HasSeries(seriesName string) (isInThisSeries bool, index int, err error) {
	return
}

// AddTag adds a tag
func (e *Epub) AddTag(tagName string) (err error) {
	return
}

// RemoveTag removes a series
func (e *Epub) RemoveTag(tagName string) (err error) {
	return
}

// HasTag checks if epub is part of a series
func (e *Epub) HasTag(tagName string) (hasThisTag bool, err error) {
	return
}

// GetMetadata from the epub
func (e *Epub) GetMetadata() (err error) {
	book, err := epubgo.Open(e.Filename)
	if err != nil {
		fmt.Println("Error parsing EPUB")
		return
	}
	defer book.Close()
	title, err := book.Metadata("title")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		return
	}
	e.Title = title[0]
	// TODO
	return
}

