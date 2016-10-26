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

func TestCleanCategory(t *testing.T) {
	fmt.Println("+ Testing Info/TestCleanCategory()...")
	assert := assert.New(t)
	clean, err := cleanCategory(fiction)
	assert.Nil(err, "valid category entered")
	assert.Equal(fiction, clean)
	clean, err = cleanCategory(" " + nonfiction + "    ")
	assert.Nil(err, "valid category entered")
	assert.Equal(nonfiction, clean)
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

	t5 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(u, option1, option3, &t5)
	assert.Equal(4, len(t5), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(u.Tag(option1, true), t5[0], expectedLocal)
	assert.Equal(u.Tag(option3, false), t5[1], expectedRemote)
	assert.Equal(option4, t5[2], unexpectedOption)
	assert.Equal(option5, t5[3], unexpectedOption)

	t6 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(u, option1, option1, &t6)
	assert.Equal(4, len(t6), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(u.Tag(u.Tag(option1, false), true), t6[0], expectedLocal+" & "+expectedRemote)
	assert.Equal(option3, t6[1], unexpectedOption)
	assert.Equal(option4, t6[2], unexpectedOption)
	assert.Equal(option5, t6[3], unexpectedOption)

	// + one tag is not found
	t7 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(u, option1, option1+"_", &t7)
	assert.Equal(4, len(t7), fmt.Sprintf(expectedOptionsCount, 4))
	assert.Equal(u.Tag(option1, true), t7[0], expectedLocal)
	assert.Equal(option3, t7[1], unexpectedOption)
	assert.Equal(option4, t7[2], unexpectedOption)
	assert.Equal(option5, t7[3], unexpectedOption)

	t8 := []string{option1, option1, option3, option4, option5, option5}
	CleanSliceAndTagEntries(u, option1, option1, &t8, option5)
	assert.Equal(3, len(t8), fmt.Sprintf(expectedOptionsCount, 3))
	assert.Equal(u.Tag(u.Tag(option1, false), true), t8[0], expectedLocal+" & "+expectedRemote)
	assert.Equal(option3, t8[1], unexpectedOption)
	assert.Equal(option4, t8[2], unexpectedOption)
}
