package book

import (
	"fmt"
	"testing"

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
