package book

import (
	"strings"

	c "github.com/barsanuphe/endive/config"
)

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
		newTags = append(newTags, Tag{Name: strings.TrimSpace(tag)})
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
func (t *Tags) Clean(cfg c.Config) {
	*t = cleanTags(*t, cfg)
	return
}
