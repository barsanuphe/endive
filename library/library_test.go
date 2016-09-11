package library

import (
	"fmt"
	"testing"

	e "github.com/barsanuphe/endive/endive"
	"github.com/barsanuphe/endive/mock"

	"github.com/stretchr/testify/assert"
)

func TestLibrarySearch(t *testing.T) {
	c := e.Config{}
	k := e.KnownHashes{}
	l := Library{Config: c, KnownHashes: k, DatabaseFile: "../test/endive.json", Index: &mock.IndexService{}, UI: &mock.UserInterface{}}
	assert := assert.New(t)

	err := l.Load()
	assert.Nil(err, "Error loading epubs from database")
	results, err := l.SearchAndPrint("language:fr", "default", false, false, 0)
	assert.Nil(err, "Error running query")
	fmt.Println(results)
	// TODO search all fields to check replacements
}
