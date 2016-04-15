package library

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/barsanuphe/endive/book"
)

var testDbName = "../test/endive.json"
var l = DB{DatabaseFile: testDbName}

func TestLdbLoad(t *testing.T) {
	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}
	if len(l.Books) != 2 {
		t.Errorf("Error loading epubs, expected 2 epubs, got %d: ", len(l.Books))
	}
	for _, epub := range l.Books {
		if hasMetadata := epub.Metadata.HasAny(); !hasMetadata {
			t.Errorf("Error loading epubs, epub %s does not have metadata in db", epub.GetMainFilename())
		}
	}
}

func TestLdbSave(t *testing.T) {
	tempTestDbName := "../test/db2.json"

	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}

	// save unchanged
	hasSaved, err := l.Save()
	if err != nil {
		t.Errorf("Error saving epubs to database: " + err.Error())
	}
	if hasSaved {
		t.Errorf("Error, db should not have been saved")
	}

	// changing DatabaseFile will make Save() compare current db with an
	// empty file, forcing save + new index
	l.DatabaseFile = tempTestDbName
	hasSaved, err = l.Save()
	if err != nil {
		t.Errorf("Error saving epubs to database: " + err.Error())
	}
	if !hasSaved {
		t.Errorf("Error saving epubs to database")
	}

	// compare both jsons
	db1, err := ioutil.ReadFile(testDbName)
	db2, err2 := ioutil.ReadFile(tempTestDbName)
	if err != nil || err2 != nil {
		t.Errorf("Error reading db file")
	}
	if !bytes.Equal(db1, db2) {
		t.Errorf("Error: original db != saved db")
	}
	// remove db2
	err = os.Remove(tempTestDbName)
	if err != nil {
		t.Errorf("Error removing temp copy test/db2.json")
	}
	l.DatabaseFile = testDbName
}
