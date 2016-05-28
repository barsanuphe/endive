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
	"en": []string{"en-US", "en-GB", "eng", "en-CA"},
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
	"science-fiction": []string{"sf", "sci-fi", "scifi-fantasy", "scifi", "science fiction",
		"sciencefiction", "sci-fi-fantasy", "sf-fantasy", "genre-sf",
		"science-fiction-speculative", "sf-f", "sff",
		"science-fiction-fantasy", "sci-fi-and-fantasy"},
	"hard sf":       []string{"hard-sf", "hard-sci-fi"},
	"fantasy":       []string{"fantasy-sci-fi", "fantasy-scifi", "fantasy-fiction"},
	"dystopia":      []string{"dystopian"},
	"nonfiction":    []string{"non-fiction"},
	"young adult":   []string{"ya"},
	"fairy tale":    []string{"fairy-tales", "fairytales"},
	"historical":    []string{"historical-fiction"},
	"children":      []string{"childrens", "children-s", "kids", "juvenile"},
	"humor":         []string{"humour"},
	"anthology":     []string{"anthologies"},
	"steampunk":     []string{"steam-punk"},
	"new weird":     []string{"new-weird", "weird", "weird-fiction"},
	"time travel":   []string{"time-travel"},
	"space opera":   []string{"space-opera"},
	"urban fantasy": []string{"urban-fantasy"},
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
	"moved-on", "record", "arc", "z-", "livre", "unsorted", "gave-up", "abandoned",
	"retelling", "middle-grade", "collection", "english", "netgalley", "available",
	"list", "stand-alone", "meh", "amazon",
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
	// reducing to main alias
	for mainalias, aliasList := range tagAliases {
		_, isIn := h.StringInSlice(tagName, aliasList)
		if isIn {
			tagName = mainalias
			break
		}
	}
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
	return
}

func cleanCategory(category string) (clean string, err error) {
	cleanName, err := cleanTagName(category)
	if err != nil {
		return "", err
	}
	if _, isIn := h.StringInSlice(cleanName, validCategories); !isIn {
		err = errors.New("Invalid category " + category)
	} else {
		clean = cleanName
	}
	return
}

func cleanHTML(desc string) (clean string) {
	return strings.Replace(sanitize.HTML(desc), "\n", "", -1)
}
