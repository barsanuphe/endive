package book

import (
	"strings"

	h "github.com/barsanuphe/endive/helpers"
)

// Tag holds the name of a tag.
type Tag struct {
	Name string `json:"tagname" xml:"name,attr"`
}

// Tags can track a book's Tags
type Tags []Tag

// String give a string representation of Tags.
func (t *Tags) String() (text string) {
	for _, tag := range *t {
		text += tag.Name + " "
	}
	return
}

// Has finds out if a Tag is already in list.
func (t *Tags) Has(o Tag) (isIn bool) {
	for _, tag := range *t {
		if o.Name == tag.Name {
			return true
		}
	}
	return false
}

// Clean a list of tags.
func (t *Tags) Clean() {
	cleanTags := Tags{}
	// TODO: names of months, dates
	// remove shelf names that are obviously not genres
	forbiddenTags := []string{
		"own", "school", "favorite", "favourite", "book", "adult",
		"read", "kindle", "borrowed", "classic", "novel", "buy",
		"star", "release", "wait", "soon", "wish", "published", "want",
		"tbr", "series", "finish", "to-", "not-", "library", "audible",
		"coming", "anticipated", "default", "recommended", "-list", "sequel",
	}
	// remove duplicates
	tagAliases := make(map[string][]string)
	tagAliases["science-fiction"] = []string{"sci-fi", "scifi-fantasy", "scifi", "science fiction", "sciencefiction"}
	tagAliases["fantasy"] = []string{"fantasy-sci-fi", "fantasy-scifi", "fantasy-fiction"}
	tagAliases["dystopia"] = []string{"dystopian"}

	for _, tag := range *t {
		clean := true
		// reducing to main alias
		for mainalias, aliasList := range tagAliases {
			_, isIn := h.StringInSlice(tag.Name, aliasList)
			if isIn {
				tag.Name = mainalias
				break
			}
		}
		// checking if not forbidden
		for _, ft := range forbiddenTags {
			if strings.Contains(tag.Name, ft) {
				clean = false
				break
			}
		}
		// adding if not already present
		if clean && !cleanTags.Has(tag) {
			cleanTags = append(cleanTags, tag)
		}
	}
	*t = cleanTags
}
