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
		"urn:ISBN: 9-780-5750-8-365-3  ",
		"9780575083653",
		nil,
	},
	{
		"9780575083653",
		"9780575083653",
		nil,
	},
	{
		"9780575083652",
		"",
		errors.New("ISBN-13 not found"),
	},
	{
		"0575083654",
		"9780575083653",
		nil,
	},
	{
		"0575083655",
		"",
		errors.New("ISBN-13 not found"),
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
	{
		"urn:uuid:0adf2006-7812-4675-9c27-47699d21c4a2",
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
		lg := cleanLanguage(c.candidate)
		assert.Equal(c.expected, lg, "Error cleaning language")
	}
}
