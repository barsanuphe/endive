package main

import (
	"errors"
	"fmt"

	"github.com/barsanuphe/epubgo"
)

type Metadata struct {
	Fields map[string][]string
}

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

	// TODO map alias creator -- author
	for _, field := range []string{"title", "creator", "description", "source", "language"} {
		m.Fields[field] = []string{"N/A"}
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
	return
}

// Get field values
func (m *Metadata) Get(field string) (values []string) {
	// TODO test field
	return m.Fields[field]
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
