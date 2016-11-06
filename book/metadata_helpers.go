package book

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/kennygrant/sanitize"

	e "github.com/barsanuphe/endive/endive"
)

// tagAliases defines redundant tags and a main alias for them.
var languageAliases = map[string][]string{
	"en": []string{"en-US", "en-GB", "eng", "en-CA"},
	"fr": []string{"fr-FR", "fre"},
	"es": []string{"spa"},
}

func cleanLanguage(language string) (clean string) {
	clean = strings.TrimSpace(strings.ToLower(language))
	// reducing to main alias
	for mainalias, aliasList := range languageAliases {
		_, isIn := e.StringInSlice(language, aliasList)
		if isIn {
			clean = mainalias
			break
		}
	}
	return
}

// GR information return the "medium" version of the cover url. This generates the "large" URL.
func getLargeGRUrl(url string) string {
	re := regexp.MustCompile(`[0-9]+`)
	ids := re.FindAllString(url, -1)
	if len(ids) == 2 {
		return fmt.Sprintf("https://images.gr-assets.com/books/%sl/%s.jpg", ids[0], ids[1])
	}
	return url
}

// remove shelf names that are obviously not genres
var forbiddenTags = []string{
	"own", "school", "favorite", "favourite", "book", "adult",
	"read", "kindle", "borrowed", "classic", "buy",
	"star", "release", "wait", "soon", "wish", "published", "want",
	"tbr", "series", "finish", "to-", "not-", "library", "audible",
	"coming", "anticipated", "default", "recommended", "-list", "sequel",
	"general", "have", "bundle", "maybe", "audio", "podcast", "calibre", "bks",
	"moved-on", "record", "arc", "z-", "livre", "unsorted", "gave-up", "abandoned",
	"retelling", "middle-grade", "collection", "english", "netgalley", "available",
	"list", "stand-alone", "meh", "amazon", "paperback", "giveaways", "review-copy",
	"check", "queue", "dnf",
}

func cleanTags(tags Tags) (cleanTags Tags) {
	cleanTags = Tags{}
	for _, tag := range tags {
		cleanName, err := cleanTagName(tag.Name)
		if err == nil {
			cleanTags.Add(Tag{Name: cleanName})
		}
	}
	// NOTE: this limit is completely arbitrary
	// only keep top10 tags, since they are ordered by popularity and will be increasingly wrong.
	if len(cleanTags) > 10 {
		cleanTags = cleanTags[:10]
	}
	return
}

func cleanTagName(tagName string) (cleanTagName string, err error) {
	tagName = strings.TrimSpace(tagName)
	tagName = strings.ToLower(tagName)
	// checking if not forbidden
	for _, ft := range forbiddenTags {
		if strings.Contains(tagName, ft) {
			err = errors.New("Forbidden tag " + tagName)
			break
		}
	}
	// returning if not forbidden
	if err == nil {
		cleanTagName = tagName
	}
	if cleanTagName == "" {
		err = errors.New("Tag cannot be empty string")
	}
	return
}

// categoryAliases replaces category tags with the canonical version
var categoryAliases = map[string][]string{
	fiction:    []string{"fiction", "fic"},
	nonfiction: []string{"non fiction", "non-fiction", "nonfiction", "nonfic"},
}

func cleanCategory(category string) (string, error) {
	clean := strings.TrimSpace(strings.ToLower(category))
	// reducing to main alias
	for mainalias, aliasList := range categoryAliases {
		if _, isIn := e.StringInSlice(clean, aliasList); isIn {
			clean = mainalias
			break
		}
	}
	// testing if valid
	if _, isIn := e.StringInSlice(clean, validCategories); !isIn {
		return "", errors.New("Invalid category " + category)
	}
	return clean, nil
}

// typeAliases replaces category tags with the canonical version
var typeAliases = map[string][]string{
	essay:         []string{"essay"},
	biography:     []string{"biography"},
	autobiography: []string{"autobiography"},
	novel:         []string{"novel"},
	shortstory:    []string{"shortstory", "short story", "short-story", "novella", "short-stories", "shortstories"},
	anthology:     []string{"anthology", "anthologies"},
	poetry:        []string{"poetry", "poems"},
}

func cleanType(typ string) (string, error) {
	clean := strings.TrimSpace(strings.ToLower(typ))
	// reducing to main alias
	for mainalias, aliasList := range typeAliases {
		if _, isIn := e.StringInSlice(clean, aliasList); isIn {
			clean = mainalias
			break
		}
	}
	// testing if valid
	if _, isIn := e.StringInSlice(clean, validTypes); !isIn {
		return "", errors.New("Invalid type " + typ)
	}
	return clean, nil
}

func cleanHTML(desc string) string {
	return sanitize.HTML(desc)
}

// CleanSliceAndTagEntries as remote or local among a list
func CleanSliceAndTagEntries(ui e.UserInterface, local, remote string, options *[]string, otherStringsToClean ...string) {
	e.RemoveDuplicates(options, otherStringsToClean...)
	// NOTE: what to do is local or remote are not found?
	// NOTE: for now, ignore that
	for i, x := range *options {
		if x == remote {
			(*options)[i] = ui.Tag((*options)[i], false)
		}
		if x == local {
			(*options)[i] = ui.Tag((*options)[i], true)
		}
	}
}

// return public field name, field, canbeset, error
func getField(i interface{}, fieldMap map[string]string, name string) (string, reflect.Value, bool, error) {
	structFieldName := ""
	publicFieldName := ""

	// try to find struct name from public name
	for k, v := range fieldMap {
		if v == name || k == name {
			structFieldName = v
			publicFieldName = k
		}
	}
	if structFieldName == "" {
		// nothing was found, invalid field
		return "", reflect.Value{}, false, fmt.Errorf(invalidField, name)
	}

	structField := reflect.ValueOf(i).Elem().FieldByName(structFieldName)
	if !structField.IsValid() {
		return "", reflect.Value{}, false, fmt.Errorf(invalidField, name)
	}
	return publicFieldName, structField, structField.CanSet(), nil
}
