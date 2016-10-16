package book

import (
	"errors"
	"strings"

	e "github.com/barsanuphe/endive/endive"

	"github.com/kennygrant/sanitize"
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

func cleanCategory(category string) (clean string, err error) {
	cleanName, err := cleanTagName(category)
	if err != nil {
		return "", err
	}
	if _, isIn := e.StringInSlice(cleanName, validCategories); !isIn {
		err = errors.New("Invalid category " + category)
	} else {
		clean = cleanName
	}
	return
}

func cleanType(bookType string) (clean string, err error) {
	cleanName, err := cleanTagName(bookType)
	if err != nil {
		return "", err
	}
	if _, isIn := e.StringInSlice(cleanName, validTypes); !isIn {
		err = errors.New("Invalid type " + bookType)
	} else {
		clean = cleanName
	}
	return
}

func cleanHTML(desc string) string {
	return sanitize.HTML(desc)
}
