package book

import (
	"fmt"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
	"github.com/barsanuphe/epubgo"
)

// Metadata holds all of the recognized metadata fiels in the OPF file of an Epub.
type Metadata struct {
	Config cfg.Config          `json:"-"`
	Fields map[string][]string `json:"fields"`
}

// NewMetadata returns a properly initialized Metadata.
func NewMetadata(c cfg.Config) *Metadata {
	return &Metadata{Fields: make(map[string][]string), Config: c}
}

// Read from the epub.
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
		if err == nil && len(results) != 0 {
			m.Fields[field] = []string{}
			for _, t := range results {
				m.Fields[field] = append(m.Fields[field], t.Content)
			}
		}
	}

	// default value for publication year
	m.Fields["year"] = []string{"XXXX"}
	dateEvents, dateErr := book.MetadataElement("date")
	if dateErr != nil {
		fmt.Println("Error parsing EPUB")
	} else {
		found := false
		for _, el := range dateEvents {
			for _, evt := range el.Attr {
				if evt == "publication" {
					m.Fields["year"] = []string{el.Content[0:4]}
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			// using first date found
			m.Fields["year"] = []string{dateEvents[0].Content[0:4]}
		}
	}
	if m.Refresh() {
		fmt.Println("Found author alias: " + m.GetFirstValue("creator"))
	}
	return
}

// HasField checks if a type of metadata is known.
func (m *Metadata) HasField(field string) (hasField bool) {
	for f := range m.Fields {
		if f == field {
			return true
		}
	}
	return
}

// Get field values.
func (m *Metadata) Get(field string) (values []string) {
	// test field
	if m.HasField(field) {
		return m.Fields[field]
	}
	return []string{}
}

// GetFirstValue of a given field.
func (m *Metadata) GetFirstValue(field string) (value string) {
	// test field
	if m.HasField(field) {
		return m.Fields[field][0]
	}
	return
}

// HasAny checks if metadata was parsed.
func (m *Metadata) HasAny() (hasMetadata bool) {
	// if at least one field contains something else than N/A, return true
	for _, values := range m.Fields {
		if values[0] != "N/A" {
			return true
		}
	}

	return
}

// IsSimilar checks if metadata is similar to known Metadata.
func (m *Metadata) IsSimilar(o *Metadata) (isSimilar bool) {
	// TODO do much better, try with isbn if available on both sides
	// similar == same author/title, for now
	if m.GetFirstValue("creator") == o.GetFirstValue("creator") && m.GetFirstValue("title") == o.GetFirstValue("title") {
		return true
	}
	return
}

// Refresh updates Medatata fields, using the configuration file.
func (m *Metadata) Refresh() (hasChanged bool) {
	// for now, only taking into account author aliases
	for i, author := range m.Fields["creator"] {
		for mainalias, aliases := range m.Config.AuthorAliases {
			_, isIn := h.StringInSlice(author, aliases)
			if isIn {
				m.Fields["creator"][i] = mainalias
				break
			}
		}
	}
	return
}
