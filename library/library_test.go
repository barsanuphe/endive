package library

import (
	"fmt"
	"testing"

	"os"

	cfg "github.com/barsanuphe/endive/config"
	h "github.com/barsanuphe/endive/helpers"
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
	l := Library{c, k, ldb}

	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}
	results, err := l.RunQuery("language:fr")
	if err != nil {
		t.Errorf("Error running query: " + err.Error())
	}
	fmt.Println(results)
	// TODO search all fields to check replacements

}
