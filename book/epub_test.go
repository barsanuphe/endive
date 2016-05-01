package book

import (
	"errors"
	"fmt"
	"testing"
)

func TestEpubGetHash(t *testing.T) {
	fmt.Println("+ Testing Epub.GetHash()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)
		err := e.RetailEpub.GetHash()
		if err != nil {
			t.Errorf("Error calculating hash for %s", e.FullPath())
		}
		if e.RetailEpub.Hash != testEpub.expectedSha256 {
			t.Errorf("GetHash(%s) returned %s, expected %s!", testEpub.filename, e.RetailEpub.Hash, testEpub.expectedSha256)
		}
	}
}

func TestEpubFlagReplacement(t *testing.T) {
	fmt.Println("+ Testing Epub.FlagForReplacement()...")
	for i, testEpub := range epubs {
		e := NewBook(i, testEpub.filename, standardTestConfig, true)
		if e.RetailEpub.NeedsReplacement != "false" {
			t.Errorf("NeedsReplacement returned %s, expected false!", e.RetailEpub.NeedsReplacement)
		}
		err := e.RetailEpub.FlagForReplacement()
		if err != nil {
			t.Errorf("Error flagging for replacement")
		}
		if e.RetailEpub.NeedsReplacement != "true" {
			t.Errorf("NeedsReplacement returned %s, expected true!", e.RetailEpub.NeedsReplacement)
		}
	}
}

// TestEpubCheck tests for Check
func TestEpubCheck(t *testing.T) {
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

var isbns = []struct {
	candidate     string
	expectedISBN  string
	expectedError error
}{
	{
		"urn:ISBN: 12-2323-4-333-432  ",
		"1223234333432",
		nil,
	},
	{
		"1223234333432",
		"1223234333432",
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
		if isbn != c.expectedISBN {
			t.Errorf("Error cleaning isbn, got %s, expected %s.", isbn, c.expectedISBN)
		}
	}
}
