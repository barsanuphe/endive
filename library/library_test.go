package library

import (
	"fmt"
	"os"
	"testing"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// init logger
	err := h.GetLogger("log_testing")
	if err != nil {
		panic(err)
	}
	// do the actual testing
	retCode := m.Run()
	// cleanup
	h.LogFile.Close()
	if err := os.Remove("log_testing"); err != nil {
		panic(err)
	}
	os.Exit(retCode)
}

func TestLibrarySearch(t *testing.T) {
	c := cfg.Config{}
	k := cfg.KnownHashes{}
	ldb := DB{DatabaseFile: "../test/endive.json"}
	l := Library{Config: c, KnownHashes: k, DB: ldb}
	assert := assert.New(t)

	err := l.Load()
	assert.Nil(err, "Error loading epubs from database")
	results, err := l.Search("language:fr", "default", false, false, 0)
	assert.Nil(err, "Error running query")
	fmt.Println(results)
	// TODO search all fields to check replacements
}
