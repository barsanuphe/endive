package book

import (
	"fmt"
	"testing"

	u "github.com/barsanuphe/endive/ui"
	"github.com/stretchr/testify/assert"
)

var languages = []struct {
	candidate string
	expected  string
}{
	{"eng", "en"},
	{"engg", "engg"},
	{"fre", "fr"},
	{"fr", "fr"},
}

func TestEpubCleanLanguages(t *testing.T) {
	fmt.Println("+ Testing Info/CleanLanguages()...")
	assert := assert.New(t)
	for _, c := range languages {
		lg := cleanLanguage(c.candidate)
		assert.Equal(c.expected, lg, "Error cleaning language")
	}
}

const (
	validCategory = "valid category entered"
)

func TestCleanCategory(t *testing.T) {
	fmt.Println("+ Testing Info/TestCleanCategory()...")
	assert := assert.New(t)
	clean, err := cleanCategory(fiction)
	assert.Nil(err, validCategory)
	assert.Equal(fiction, clean)
	for _, nf := range []string{nonfiction, " " + nonfiction + "    ", "non fiction", "non-Fiction"} {
		clean, err = cleanCategory(nf)
		assert.Nil(err, validCategory)
		assert.Equal(nonfiction, clean)
	}
	_, err = cleanCategory("invalid category")
	assert.NotNil(err, "invalid category entered")
}

func TestCleanType(t *testing.T) {
	fmt.Println("+ Testing Info/TestCleanType()...")
	assert := assert.New(t)
	for _, v := range validTypes {
		clean, err := cleanType(v)
		assert.Nil(err, "valid type entered")
		assert.Equal(v, clean)
	}
	clean, err := cleanType(" " + essay + "    ")
	assert.Nil(err, "valid category entered")
	assert.Equal(essay, clean)
	_, err = cleanType("invalid type")
	assert.NotNil(err, "invalid category entered")
	clean, err = cleanType("short-story")
	assert.Nil(err, "valid type entered")
	assert.Equal(shortstory, clean)
	clean, err = cleanType("novella")
	assert.Nil(err, "valid type entered")
	assert.Equal(shortstory, clean)
}

func TestGetLargeImgURL(t *testing.T) {
	fmt.Println("+ Testing Info/getLargeGRUrl()...")
	assert := assert.New(t)

	mURL := "https://images.gr-assets.com/books/1426192671m/53732.jpg"
	lURL := "https://images.gr-assets.com/books/1426192671l/53732.jpg"
	badURL := "https://images.gr-assets.com/books/1426192671m/bad.jpg"

	assert.Equal(lURL, getLargeGRUrl(mURL))
	assert.Equal(lURL, getLargeGRUrl(lURL))
	assert.Equal(badURL, getLargeGRUrl(badURL))
}

const (
	option1              = "option1"
	option3              = "option3"
	option4              = "option4"
	option5              = "option5"
	unexpectedOption     = "unexpected option label"
	expectedOptionsCount = "should have %d element(s)"
	expectedLocal        = "should be tagged local"
	expectedRemote       = "should be tagged remote"
)

func TestOptions(t *testing.T) {
	fmt.Println("+ Testing CleanSliceAndTagEntries/RemoveDuplicates()...")
	assert := assert.New(t)
	u := &u.UI{}

	t1 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(u, option1, option3, &t1)
	assert.Equal(4, len(t1), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(u.Tag(option1, true), t1[0], expectedLocal)
	assert.Equal(u.Tag(option3, false), t1[1], expectedRemote)
	assert.Equal(option4, t1[2], unexpectedOption)
	assert.Equal(option5, t1[3], unexpectedOption)

	t2 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(u, option1, option1, &t2)
	assert.Equal(4, len(t2), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(u.Tag(u.Tag(option1, false), true), t2[0], expectedLocal+" & "+expectedRemote)
	assert.Equal(option3, t2[1], unexpectedOption)
	assert.Equal(option4, t2[2], unexpectedOption)
	assert.Equal(option5, t2[3], unexpectedOption)

	// + one tag is not found
	t3 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(u, option1, option1+"_", &t3)
	assert.Equal(4, len(t3), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(u.Tag(option1, true), t3[0], expectedLocal)
	assert.Equal(option3, t3[1], unexpectedOption)
	assert.Equal(option4, t3[2], unexpectedOption)
	assert.Equal(option5, t3[3], unexpectedOption)

	t4 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(u, option1, option1, &t4, option5)
	assert.Equal(3, len(t4), fmt.Sprintf(expectedOptionsCount, 3))
	assert.Equal(u.Tag(u.Tag(option1, false), true), t4[0], expectedLocal+" & "+expectedRemote)
	assert.Equal(option3, t4[1], unexpectedOption)
	assert.Equal(option4, t4[2], unexpectedOption)
}
