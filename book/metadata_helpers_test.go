package book

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var isbns = []struct {
	candidate     string
	expectedISBN  string
	expectedError error
}{
	{
		"urn:ISBN: 97-8323-4-333-432  ",
		"9783234333432",
		nil,
	},
	{
		"9783234333432",
		"9783234333432",
		nil,
	},
	{
		"A223234333432",
		"",
		errors.New("ISBN-13 not found"),
	},
	{
		"urn:isbn: 12-23-4-333-432  ",
		"",
		errors.New("ISBN-13 not found"),
	},
}

func TestEpubCleanISBN(t *testing.T) {
	fmt.Println("+ Testing Info/CleanISBN()...")
	for _, c := range isbns {
		isbn, err := cleanISBN(c.candidate)
		if err == nil && c.expectedError != nil {
			t.Errorf("Unexpected error cleaning isbn %s", c.candidate)
		} else if err != nil && c.expectedError == nil {
			t.Errorf("Unexpected error cleaning isbn %s", c.candidate)
		} else if err != nil && c.expectedError != nil && c.expectedError.Error() != err.Error() {
			t.Errorf("Unexpected error cleaning isbn %s: got %s, expected %s", c.candidate, c.expectedError.Error(), err.Error())
		}
		assert.Equal(t, c.expectedISBN, isbn, "Error cleaning isbn")
	}
}

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
		lg, err := cleanLanguage(c.candidate)
		assert.Nil(err, "Error cleaning language")
		assert.Equal(c.expected, lg, "Error cleaning language")
	}
}
