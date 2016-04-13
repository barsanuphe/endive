package main

import (
	"fmt"
	"testing"
)

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)
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
	e := NewBook(0, epubs[0].filename, standardTestConfig, isRetail)
	e.RetailEpub.GetHash()

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
