package book

import (
	"errors"
	"regexp"
	"strings"

	h "github.com/barsanuphe/endive/helpers"

	"github.com/kennygrant/sanitize"
	"github.com/moraes/isbn"
)

func cleanISBN(full string) (isbn13 string, err error) {
	// cleanup string, only keep numbers
	re := regexp.MustCompile("[0-9]+")
	candidate := strings.Join(re.FindAllString(full, -1), "")

	// if start of isbn detected, try to salvage the situation
	if len(candidate) > 13 && strings.HasPrefix(candidate, "978") {
		candidate = candidate[:13]
	}

	// validate and convert to ISBN13 if necessary
	if isbn.Validate(candidate) {
		if len(candidate) == 10 {
			isbn13, err = isbn.To13(candidate)
			if err != nil {
				isbn13 = ""
			}
		}
		if len(candidate) == 13 {
			isbn13 = candidate
		}
	} else {
		err = errors.New("ISBN-13 not found")
	}
	return
}

// tagAliases defines redundant tags and a main alias for them.
var languageAliases = map[string][]string{
	"en": []string{"en-US", "en-GB", "eng"},
	"fr": []string{"fr-FR", "fre"},
}

func cleanLanguage(language string) (clean string) {
	clean = strings.TrimSpace(language)
	// reducing to main alias
	for mainalias, aliasList := range languageAliases {
		_, isIn := h.StringInSlice(language, aliasList)
		if isIn {
			clean = mainalias
			break
		}
	}
	return
}

// tagAliases defines redundant tags and a main alias for them.
var tagAliases = map[string][]string{
	"science-fiction": []string{"sf", "sci-fi", "scifi-fantasy", "scifi", "science fiction", "sciencefiction", "sci-fi-fantasy", "sf-fantasy", "genre-sf", "science-fiction-speculative"},
	"fantasy":         []string{"fantasy-sci-fi", "fantasy-scifi", "fantasy-fiction"},
	"dystopia":        []string{"dystopian"},
	"nonfiction":      []string{"non-fiction"},
}

// TODO: names of months, dates
// remove shelf names that are obviously not genres
var forbiddenTags = []string{
	"own", "school", "favorite", "favourite", "book", "adult",
	"read", "kindle", "borrowed", "classic", "novel", "buy",
	"star", "release", "wait", "soon", "wish", "published", "want",
	"tbr", "series", "finish", "to-", "not-", "library", "audible",
	"coming", "anticipated", "default", "recommended", "-list", "sequel",
	"general", "have", "bundle", "maybe", "audio", "podcast", "calibre", "bks",
	"moved-on", "record", "arc", "z-", "livre",
}

func cleanTags(tags Tags) (cleanTags Tags, err error) {
	cleanTags = Tags{}
	for _, tag := range tags {
		clean := true
		tag.Name = strings.TrimSpace(tag.Name)

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
	// NOTE: this limit is completely arbitrary
	// only keep top10 tags, since they are ordered by popularity and will be increasingly wrong.
	if len(cleanTags) > 10 {
		cleanTags = cleanTags[:10]
	}
	return
}

func cleanHTML(desc string) (clean string) {
	return strings.Replace(sanitize.HTML(desc), "\n", "", -1)
}
