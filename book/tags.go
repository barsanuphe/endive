package book

import (
	"strings"

	h "github.com/barsanuphe/endive/helpers"
)

// tagAliases defines redundant tags and a main alias for them.
var tagAliases = map[string][]string{
	"science-fiction": []string{"sci-fi", "scifi-fantasy", "scifi", "science fiction", "sciencefiction", "sci-fi-fantasy"},
	"fantasy":         []string{"fantasy-sci-fi", "fantasy-scifi", "fantasy-fiction"},
	"dystopia":        []string{"dystopian"},
}

// TODO: names of months, dates
// remove shelf names that are obviously not genres
var forbiddenTags = []string{
	"own", "school", "favorite", "favourite", "book", "adult",
	"read", "kindle", "borrowed", "classic", "novel", "buy",
	"star", "release", "wait", "soon", "wish", "published", "want",
	"tbr", "series", "finish", "to-", "not-", "library", "audible",
	"coming", "anticipated", "default", "recommended", "-list", "sequel",
	"general",
}

// Tag holds the name of a tag.
type Tag struct {
	Name string `json:"name" xml:"name,attr"`
}

// Tags can track a book's Tags
type Tags []Tag

// String give a string representation of Tags.
func (t *Tags) String() (text string) {
	tagNames := []string{}
	for _, tag := range *t {
		tagNames = append(tagNames, tag.Name)
	}
	return strings.Join(tagNames, ", ")
}

// Add Tags to the list
func (t *Tags) Add(tags ...Tag) (added bool) {
	for _, tag := range tags {
		if isIn, _ := t.Has(tag); !isIn {
			*t = append(*t, tag)
			added = true
		}
	}
	return
}

// AddFromNames Tags to the list, from []string
func (t *Tags) AddFromNames(tags ...string) (added bool) {
	newTags := Tags{}
	for _, tag := range tags {
		newTags = append(newTags, Tag{Name: tag})
	}
	return t.Add(newTags...)
}

// Remove Tags from the list
func (t *Tags) Remove(tags ...Tag) (removed bool) {
	for _, tag := range tags {
		if isIn, i := t.Has(tag); isIn {
			*t = append((*t)[:i], (*t)[i+1:]...)
			removed = true
		}
	}
	return
}

// RemoveFromNames Tags to the list, from []string
func (t *Tags) RemoveFromNames(tags ...string) (removed bool) {
	newTags := Tags{}
	for _, tag := range tags {
		newTags = append(newTags, Tag{Name: tag})
	}
	return t.Remove(newTags...)
}

// Has finds out if a Tag is already in list.
func (t *Tags) Has(o Tag) (isIn bool, index int) {
	for i, tag := range *t {
		if o.Name == tag.Name {
			return true, i
		}
	}
	return
}

// Clean a list of tags.
func (t *Tags) Clean() {
	cleanTags := Tags{}
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
		if clean {
			cleanTags.Add(tag)
		}
	}
	*t = cleanTags
}
