package main

import (
	"fmt"
	"os"
	"testing"
)

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	for _, testEpub := range epubs {
		e := NewBook(testEpub.filename, standardTestConfig, true)
		err := e.RetailEpub.GetHash()
		if err != nil {
			t.Errorf("Error calculating hash for %s", e.getMainFilename())
		}
		if e.RetailEpub.Hash != testEpub.expectedSha256 {
			t.Errorf("GetHash(%s) returned %s, expected %s!", testEpub.filename, e.RetailEpub.Hash, testEpub.expectedSha256)
		}
	}
}

// TestEpubRetail tests for SetRetail, SetNonRetail and Check
func TestEpubRetail(t *testing.T) {
	fmt.Println("+ Testing Epub.SetRetail()...")
	e := NewBook(epubs[0].filename, standardTestConfig, isRetail)
	e.RetailEpub.GetHash()

	// testing retail
	err := e.RetailEpub.SetRetail()
	if err != nil {
		t.Errorf("Error setting retail")
	}
	mode, err := os.Stat(e.getMainFilename())
	if mode.Mode() != 0444 {
		t.Errorf("Error: ebook should be read-only")
	}
	// checking retail
	hasChanged, err := e.RetailEpub.Check()
	if err != nil {
		t.Errorf("Error checking hash" + err.Error())
	}
	if hasChanged {
		t.Errorf("Error: ebook should be not have changed")
	}
	oldHash := e.RetailEpub.Hash
	e.RetailEpub.Hash = ""
	hasChanged, err = e.RetailEpub.Check()
	if err != nil {
		t.Errorf("Error checking retail hash")
	}
	if !hasChanged {
		t.Errorf("Error: ebook has changed")
	}

	// testing non-retail
	e.RetailEpub.Hash = oldHash
	err = e.RetailEpub.SetNonRetail()
	if err != nil {
		t.Errorf("Error setting non-retail")
	}
	mode, err = os.Stat(e.getMainFilename())
	if mode.Mode() != 0777 {
		t.Errorf("Error: ebook should be read-write")
	}

	// checking non retail
	hasChanged, err = e.RetailEpub.Check()
	if err != nil {
		t.Errorf("Error checking hash")
	}
	if hasChanged {
		t.Errorf("Error: ebook should be not have changed")
	}
	e.RetailEpub.Hash = ""
	hasChanged, err = e.RetailEpub.Check()
	if err != nil {
		t.Errorf("Error checking non retail hash, should have been ok")
	}
	if !hasChanged {
		t.Errorf("Error: ebook has changed")
	}
}
