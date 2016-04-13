package main

import (
	"errors"
	"fmt"

	"github.com/barsanuphe/epubgo"
)

// Metadata holds all of the recognized metadata fiels in the OPF file of an Epub.
type Metadata struct {
	Fields map[string][]string `json:"fields"`
}

// NewMetadata returns a properly initialized Metadata.
func NewMetadata() *Metadata {
	return &Metadata{Fields: make(map[string][]string)}
}

// Read from the epub
func (m *Metadata) Read(path string) (err error) {
	fmt.Println("Reading metadata from OPF in ..." + path)
	book, err := epubgo.Open(path)
	if err != nil {
		fmt.Println("Error parsing EPUB")
		return
	}
	defer book.Close()

	// get all possible fields except for date
	knownFields := []string{
		"title", "language", "identifier", "creator", "subject",
		"description", "publisher", "contributor", "type", "format",
		"source", "relation", "coverage", "rights", "meta",
	}
	for _, field := range knownFields {
		m.Fields[field] = []string{"Unknown"}
		results, err := book.MetadataElement(field)
		if err == nil {
			m.Fields[field] = []string{}
			for _, t := range results {
				m.Fields[field] = append(m.Fields[field], t.Content)
			}
		}
	}

	dateEvents, err := book.MetadataElement("date")
	if err != nil {
		fmt.Println("Error parsing EPUB")
		m.Fields["year"] = []string{"XXXX"}
	} else {
		found := false
		for _, el := range dateEvents {
			for _, evt := range el.Attr {
				if evt == "publication" {
					m.Fields["year"] = []string{el.Content[0:4]}
					if err != nil {
						panic(err)
					}
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			fmt.Println("Error parsing EPUB, no publication year")
			err = errors.New("No publication date")
		}
	}
	return
}

// HasField checks if a type of metadata is known
func (m *Metadata) HasField(field string) (hasField bool) {
	for f := range m.Fields {
		if f == field {
			return true
		}
	}
	return
}

// Get field values
func (m *Metadata) Get(field string) (values []string) {
	// test field
	if m.HasField(field) {
		return m.Fields[field]
	}
	return []string{}
}

// HasAny checks if metadata was parsed
func (m *Metadata) HasAny() (hasMetadata bool) {
	// if at least one field contains something else than N/A, return true
	for _, values := range m.Fields {
		if values[0] != "N/A" {
			return true
		}
	}

	return
}

// IsSimilar checks if metadata is similar to known Metadata
func (m *Metadata) IsSimilar(o *Metadata) (isSimilar bool) {
	// TODO do much better, try with isbn if available on both sides
	// similar == same author/title, for now
	if m.Get("creator")[0] == o.Get("creator")[0] && m.Get("title")[0] == o.Get("title")[0] {
		return true
	}
	return
}
