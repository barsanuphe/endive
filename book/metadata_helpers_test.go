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
